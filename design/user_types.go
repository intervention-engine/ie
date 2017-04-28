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

var HuddleMembership = Type("HuddleMembership", func() {
	Attribute("id", String, "Relationship ID")
	Attribute("huddle_id", String, "huddle ID")
	Attribute("patient_id", String, "Patient ID")
	Attribute("created_at", DateTime, "Timestamp of membership")
	Attribute("reason", String, "Reason patient was added to huddle")
	Attribute("reviewed", Boolean, "Has patient been reviewed in this huddle")
})

var CareTeamMembership = Type("CareTeamMembership", func() {
	Attribute("id", String, "Relationship ID")
	Attribute("care_team_id", String, "Care Team ID")
	Attribute("patient_id", String, "Patient ID")
	Attribute("created_at", DateTime, "Timestamp of membership")
})
