package groups

import (
	"errors"
	"fmt"

	fhir "github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/search"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	// Register the groupId parameter
	search.GlobalRegistry().RegisterParameterInfo(GroupParamInfo)
	search.GlobalRegistry().RegisterParameterParser(GroupParamInfo.Type, GroupParamParser)
	search.GlobalMongoRegistry().RegisterBSONBuilder(GroupParamInfo.Type, GroupBSONBuilder)

	// Register the _query parameter
	search.GlobalRegistry().RegisterParameterInfo(QueryParamInfo)
	search.GlobalRegistry().RegisterParameterParser(QueryParamInfo.Type, QueryParamParser)
	search.GlobalMongoRegistry().RegisterBSONBuilder(QueryParamInfo.Type, QueryBSONBuilder)
}

// GroupParam represents the group parameter.  Since this is an id, we simply embed a TokenParam. After fhir search is
// fully refactored, we may be able to get rid of GroupParam and just return a TokenParam from the parser.
type GroupParam struct {
	search.TokenParam
}

// GroupParamInfo represents the "groupId" parameter on patients.  This allows patients to be searched
// by their membership to a group.  Their membership may be explicit (they are explicitly named as a
// member of a group) or implicit (they match characteristics criteria of a group).
var GroupParamInfo = search.SearchParamInfo{
	Resource: "Patient",
	Name:     "groupId",
	Type:     "ie.group",
}

// GroupParamParser parses the parameter and returns a GroupParam
var GroupParamParser = func(info search.SearchParamInfo, data search.SearchParamData) (search.SearchParam, error) {
	return &GroupParam{
		TokenParam: *search.ParseTokenParam(data.Value, info),
	}, nil
}

// GroupBSONBuilder builds the Mongo BSON object corresponding to the query for group membership.
var GroupBSONBuilder = func(param search.SearchParam, searcher *search.MongoSearcher) (object bson.M, err error) {
	// IDs are tokens, so treat the custom group param as a token param
	gp, ok := param.(*GroupParam)
	if !ok {
		return nil, errors.New("Expected a GroupParam")
	}

	// First get the group
	var groupID string
	if bson.IsObjectIdHex(gp.Code) {
		groupID = bson.ObjectIdHex(gp.Code).Hex()
	} else {
		return nil, fmt.Errorf("Invalid BSON ID: %s", gp.Code)
	}
	var group fhir.Group
	if err := searcher.GetDB().C("groups").Find(bson.M{"_id": groupID}).One(&group); err != nil {
		return nil, err
	}

	// If it's an "actual" group, then just look at the list of members (this is the case for huddles)
	if group.Actual != nil && *group.Actual {
		// Return a BSON object (for patient) indicating the ID can be any of those in the group
		var ids []string
		for _, m := range group.Member {
			if m.Entity != nil && m.Entity.Type == "Patient" && m.Entity.ReferencedID != "" {
				ids = append(ids, m.Entity.ReferencedID)
			}
		}
		return bson.M{
			"_id": bson.M{
				"$in": ids,
			},
		}, nil
	}

	// It's not an "actual" group, so use the group characteristics instead
	info, err := LoadCharacteristicInfo(group.Characteristic)
	if err != nil {
		return nil, err
	}

	// If the info is only patient characteristics, then this is a simple, normal patient search
	if !info.HasConditionCharacteristics && !info.HasEncounterCharacteristics {
		q := search.Query{Resource: "Patient", Query: info.PatientQueryValues.Encode()}
		return searcher.CreateQueryObject(q), nil
	}

	// Otherwise, this cannot be expressed as a patient search without first resolving the patient IDs
	ids, err := resolveGroup(info, searcher)
	if err != nil {
		return nil, err
	}

	return bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}, nil
}

// QueryParam represents the _query parameter.  This exists only to provide backwards compatibility for the
// frontend, which sends "_query=group".  Without this, the FHIR server would reject the _query parameter as
// unsupported.  Since this is a string, we simply embed a StringParam. After fhir search is
// fully refactored, we may be able to get rid of GroupParam and just return a StringParam from the parser.
type QueryParam struct {
	search.StringParam
}

// QueryParamInfo represents the info for the Patient._query parameter.
var QueryParamInfo = search.SearchParamInfo{
	Resource: "Patient",
	Name:     "_query",
	Type:     "ie.noop",
}

// QueryParamParser parses the parameter and returns a QueryParam
var QueryParamParser = func(info search.SearchParamInfo, data search.SearchParamData) (search.SearchParam, error) {
	return &QueryParam{
		StringParam: *search.ParseStringParam(data.Value, info),
	}, nil
}

// QueryBSONBuilder returns an empty BSON object unless it detects an invalid query type
var QueryBSONBuilder = func(param search.SearchParam, searcher *search.MongoSearcher) (object bson.M, err error) {
	// IDs are tokens, so treat the custom group param as a token param
	qp, ok := param.(*QueryParam)
	if !ok {
		return nil, errors.New("Expected a QueryParam")
	}
	if qp.String != "group" {
		return nil, fmt.Errorf("Unknown _query type: %s", qp.String)
	}
	return bson.M{}, nil
}
