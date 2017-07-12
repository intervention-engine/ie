package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var RiskPie = Type("pie", func() {
	Attribute("slices", ArrayOf(RiskPieSlice), "Individual Pie sli")
})

var RiskPieSlice = Type("pieSlice", func() {
	Attribute("name", String, "Risk Category Name")
	Attribute("value", Integer, "Risk Category Value")
	Attribute("weight", Integer, "Weight On Overall Risk Value")
	Attribute("maxValue", Integer, "Maximum possible value")
})

var PatientName = Type("name", func() {
	Attribute("family", String, "Family Name")
	Attribute("given", String, "Given Name")
	Attribute("middle_initial", String, "Middle Initial")
	Attribute("full", String, "Full Name")
})

var Address = Type("address", func() {
	Attribute("street", ArrayOf(String), "Street Name")
	Attribute("city", String, "City Name")
	Attribute("state", String, "State Name")
	Attribute("postal_code", String, "Postal Code")
})

var CareTeamPayload = Type("CareTeamPayload", func() {
	Attribute("id", String, "Unique care team ID")
	Attribute("name", String, "Care team name")
	Attribute("leader", String, "Care team leader")
	Attribute("created_at", DateTime, "Timestamp for care team creation")
})

var SchedulePatientPayload = Type("SchedulePatientPayload", func() {
	Attribute("patient_id", String, "Unique patient ID")
	Attribute("date", String, "Date in YYYY-MM-dd format to schedule huddle")
	Required("patient_id", "date")
})

var PatientHuddle = Type("PatientHuddle", func() {
	Attribute("id", String, "patient id")
	Attribute("reason", String, "reason for why patient is in this huddle")
	Attribute("reason_type", String, "")
	Attribute("reviewed", Boolean, "")
	Attribute("reviewed_at", DateTime, "")
})
