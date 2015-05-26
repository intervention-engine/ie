package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/intervention-engine/fhir/server"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"net/url"
	"strconv"
	"text/template"
	"time"
)

type RiskModelParameter struct {
	Category string
	Weight   float64
	Method   string
}

func RiskQueryHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	query := r.URL.Query()
	queryValues := query["_query"]
	if len(queryValues) > 0 && queryValues[0] == "risk" {
		mapFn := `function() {
  emit(this.patientid, {factType: this.type, risk: 0})
}`
		reduceSource := `function(patient_id, values) {
  var reducedObject = {
    factType: 'risk',
    risk: 0
  }

  values.forEach(function(value) {
    {{ range . }}
    if (value.factType === '{{.Category}}') {
      {{if eq .Method "count"}}
      reducedObject.risk += 1 * {{.Weight}};
      {{ end }}
      {{if eq .Method "presence"}}
      reducedObject.risk += {{.Weight}};
      {{ end }}
    };
    {{ end }}
  });

  return reducedObject;
}`
		reduceTemplate, err := template.New("reduce").Parse(reduceSource)
		var buffer bytes.Buffer
		rmps := handleRiskModelParameters(query)
		reduceTemplate.Execute(&buffer, rmps)
		mrJob := mgo.MapReduce{Map: mapFn, Reduce: buffer.String()}
		facts := server.Database.C("facts")

		type valueStruct struct {
			Risk int `bson:"risk"`
		}

		var result []struct {
			Id    string      `bson:"_id"`
			Value valueStruct `bson:"value"`
		}

		query := facts.Find(bson.M{"$or": []bson.M{{"enddate.time": nil}, {"enddate.time": bson.M{"$gt": time.Now()}}}})
		_, err = query.MapReduce(&mrJob, &result)

		if err != nil {
			panic(err.Error())
		}

		json.NewEncoder(rw).Encode(result)
	} else {
		next(rw, r)
	}
}

func handleRiskModelParameters(queryParams url.Values) []RiskModelParameter {
	rmps := []RiskModelParameter{}
	categories := []string{"MedicationStatement", "Condition", "Encounter"}
	for _, category := range categories {
		method, ok := queryParams[category]
		if ok {
			rmp := RiskModelParameter{Category: category, Method: method[0]}
			weightSlice, weightPresent := queryParams[category+"Weight"]
			if weightPresent {
				weight := weightSlice[0]
				floatWeight, err := strconv.ParseFloat(weight, 64)
				if err != nil {
					rmp.Weight = 1.0
				} else {
					rmp.Weight = floatWeight
				}
			} else {
				rmp.Weight = 1.0
			}
			rmps = append(rmps, rmp)
		}
	}
	return rmps
}
