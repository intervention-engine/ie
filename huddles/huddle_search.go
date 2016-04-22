package huddles

import "github.com/intervention-engine/fhir/search"

// RegisterCustomSearchDefinitions registers the custom search definitions needed for the Huddle profile on Group
func init() {
	registry := search.GlobalRegistry()
	registry.RegisterParameterInfo(search.SearchParamInfo{
		Resource: "Group",
		Name:     "activedatetime",
		Type:     "date",
		Paths: []search.SearchParamPath{
			{Path: "[]extension.activeDateTime", Type: "dateTime"},
		},
	})
	registry.RegisterParameterInfo(search.SearchParamInfo{
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
	})
	registry.RegisterParameterInfo(search.SearchParamInfo{
		Resource: "Group",
		Name:     "member-reviewed",
		Type:     "date",
		Paths: []search.SearchParamPath{
			{Path: "[]member.[]extension.reviewed", Type: "dateTime"},
		},
	})
}
