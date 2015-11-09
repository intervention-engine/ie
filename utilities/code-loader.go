package utilities

import (
	"archive/zip"
	"bufio"
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type CodeEntry struct {
	CodeSystem string `bson:"codeSystem",json:"codeSystem"`
	Code       string `bson:"code",json:"code"`
	Name       string `bson:"name",json:"name"`
}

func LoadICD9FromCMS(mongoHost, lookupURL string) {

	mongoSession, err := mgo.Dial(mongoHost)
	if err != nil {
		log.Fatal(err)
	}

	database := mongoSession.DB("fhir")
	codeCollection := database.C("codelookup")

	urlReader, err := getReaderFromUrl(lookupURL)
	if err != nil {
		log.Fatal(err)
	}

	zr, err := zip.NewReader(urlReader, urlReader.Size())
	if err != nil {
		log.Fatalf("Unable to read zip: %s", err)
	}

	for _, zf := range zr.File {
		if zf.Name == "CMS32_DESC_LONG_DX.txt" {
			zipFile, err := zf.Open()
			if err != nil {
				log.Fatalf("Unable to read zip file: %s", err)
			}
			scanner := bufio.NewScanner(zipFile)
			for scanner.Scan() {
				codeString := scanner.Text()
				splitEntry := strings.SplitN(codeString, " ", 2)
				code := splitEntry[0]
				name := splitEntry[1]

				if strings.HasPrefix(code, "E") && len(code) > 4 {
					code = code[:4] + "." + code[4:]
				} else if strings.HasPrefix(code, "V") && len(code) > 3 {
					code = code[:3] + "." + code[3:]
				} else {
					if len(code) > 3 {
						code = code[:3] + "." + code[3:]
					}
				}

				codeEntry := CodeEntry{
					"ICD-9",
					code,
					name,
				}

				codeCollection.Upsert(bson.M{"codeSystem": "ICD-9", "code": code}, codeEntry)
			}
		}
	}
}

func getReaderFromUrl(url string) (*bytes.Reader, error) {
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
