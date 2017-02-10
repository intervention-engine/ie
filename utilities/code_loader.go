package utilities

import (
	"archive/zip"
	"bufio"
	"bytes"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// CodeEntry represents a searchable code in the Mongo database
type CodeEntry struct {
	CodeSystem string `bson:"codeSystem",json:"codeSystem"`
	Code       string `bson:"code",json:"code"`
	Name       string `bson:"name",json:"name"`
	Count      int    `bson:"count,omitempty",json:"count,omitempty"`
}

// LoadICDFromCMS downloads CMS-provided ICD-9 or ICD-10 codes and loads them into the Mongo database
func LoadICDFromCMS(mongoHost, lookupURL, codeSystem string) {

	mongoSession, err := mgo.Dial(mongoHost)
	if err != nil {
		log.Fatal(err)
	}
	defer mongoSession.Close()
	database := mongoSession.DB("fhir")
	codeCollection := database.C("codelookup")

	urlReader, err := getReaderFromURL(lookupURL)
	if err != nil {
		log.Fatal(err)
	}

	zr, err := zip.NewReader(urlReader, urlReader.Size())
	if err != nil {
		log.Fatal("Unable to read zip:", err)
	}

	var fileName string
	switch codeSystem {
	case "ICD-9":
		fileName = "CMS32_DESC_LONG_DX.txt"
	case "ICD-10":
		fileName = "icd10cm_codes_2016.txt"
	}

	for _, zf := range zr.File {
		if zf.Name == fileName {
			zipFile, err := zf.Open()
			if err != nil {
				log.Fatal("Unable to read zip file:", err)
			}

			// Use bulk ops to clear the codes then re-insert (over 1000 times faster than upserting).
			bulk := codeCollection.Bulk()
			bulk.RemoveAll(bson.M{"codeSystem": codeSystem})

			re := regexp.MustCompile(" +")
			scanner := bufio.NewScanner(zipFile)
			for scanner.Scan() {
				splitLine := re.Split(scanner.Text(), 2)
				bulk.Insert(CodeEntry{
					CodeSystem: codeSystem,
					Code:       fixCode(splitLine[0], codeSystem),
					Name:       splitLine[1],
				})
			}

			if _, err := bulk.Run(); err != nil {
				log.Fatal("Failed to update database:", err)
			}

			log.Printf("Successfully loaded %s codes from %s", codeSystem, lookupURL)
		}
	}
}

func fixCode(code, codeSystem string) string {
	switch codeSystem {
	case "ICD-9":
		if strings.HasPrefix(code, "E") && len(code) > 4 {
			return code[:4] + "." + code[4:]
		} else if len(code) > 3 {
			return code[:3] + "." + code[3:]
		}
	case "ICD-10":
		if len(code) > 3 {
			return code[:3] + "." + code[3:]
		}
	}
	return code
}

func getReaderFromURL(url string) (*bytes.Reader, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	buf := &bytes.Buffer{}

	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(buf.Bytes()), nil
}
