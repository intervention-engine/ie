package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var RiskAssessment = Type("riskAssessment", func() {
	Attribute("id", String, "Risk Assessment ID")
	Attribute("groupId", String, "Risk Assessment Group ID")
	Attribute("date", DateTime, "Date")
	Attribute("value", Integer, "Value")
})

var PatientName = Type("name", func() {
	Attribute("family", String, "Family Name")
	Attribute("given", String, "Given Name")
	Attribute("middleInitial", String, "Middle Initial")
	Attribute("full", String, "Full Name")
})

var Address = Type("address", func() {
	Attribute("street", ArrayOf(String), "Street Name")
	Attribute("city", String, "City Name")
	Attribute("state", String, "State Name")
	Attribute("postalCode", String, "Postal Code")
})

var CareTeamPayload = Type("CareTeamPayload", func() {
	Attribute("id", String, "Unique care team ID")
	Attribute("name", String, "Care team name")
	Attribute("leader", String, "Care team leader")
	Attribute("createdAt", DateTime, "Timestamp for care team creation")
})
