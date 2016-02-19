package huddles

import "github.com/intervention-engine/fhir/search"

// RegisterCustomSearchDefinitions registers the custom search definitions needed for the Huddle profile on Group
func RegisterCustomSearchDefinitions() {
	search.SearchParameterDictionary["Group"]["activedatetime"] = search.SearchParamInfo{
		Resource: "Group",
		Name:     "activedatetime",
		Type:     "date",
		Paths: []search.SearchParamPath{
			{Path: "[]extension.activeDateTime", Type: "dateTime"},
		},
	}
	search.SearchParameterDictionary["Group"]["leader"] = search.SearchParamInfo{
		Resource: "Group",
		Name:     "leader",
		Type:     "reference",
		Paths: []search.SearchParamPath{
			{Path: "[]extension.leader", Type: "Reference"},
		},
		Targets: []string{
			"Group",
			"Practitioner",
			"Organization",
		},
	}
	search.SearchParameterDictionary["Group"]["member-reviewed"] = search.SearchParamInfo{
		Resource: "Group",
		Name:     "member-reviewed",
		Type:     "date",
		Paths: []search.SearchParamPath{
			{Path: "[]member.[]extension.reviewed", Type: "dateTime"},
		},
	}
}
