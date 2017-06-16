package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var HuddleMedia = MediaType("application/vnd.huddle+json", func() {
	Description("A single gathering of a care team at a specific time")
	Attributes(func() {
		Attribute("id", String, "Unique Huddle ID")
		Attribute("date", DateTime, "Creation timestamp")
		Attribute("careTeamId", String, "ID of the care team associated with this Huddle")
		Attribute("patients", ArrayOf(PatientHuddle), "List of patients scheduled for this huddle")
	})
	View("default", func() {
		Attribute("id", String, "Unique Huddle ID")
		Attribute("date", DateTime, "Creation timestamp")
		Attribute("careTeamId", String, "ID of the care team associated with this Huddle")
		Attribute("patients", ArrayOf(PatientHuddle), "List of patients scheduled for this huddle")
	})
})

var PatientMedia = MediaType("application/vnd.patient+json", func() {
	TypeName("Patient")
	Description("A patient")
	ContentType("application/json")
	Attributes(func() {
		Attribute("id", String, "Unique patient ID")
		Attribute("name", PatientName)
		Attribute("address", Address)
		Attribute("riskAssessments", ArrayOf(RiskAssessment))
		Attribute("age", Integer, "Age of Patient")
		Attribute("gender", String, "Gender of Patient")
		Attribute("birthDate", DateTime, "Birth Date of Patient")
		Required("id", "name")
	})
	View("default", func() {
		Attribute("id")
		Attribute("name")
		Attribute("address")
		Attribute("riskAssessments")
		Attribute("age")
		Attribute("gender")
		Attribute("birthDate")
	})
	View("link", func() {
		Attribute("id")
		Attribute("name")
	})
})

var CareTeamMedia = MediaType("application/vnd.careteam+json", func() {
	TypeName("CareTeam")
	Description("A care team")
	Reference(CareTeamPayload)
	ContentType("application/json")
	Attributes(func() {
		// Inherit from CareTeamPayload
		Attribute("id")
		Attribute("name")
		Attribute("leader")
		Attribute("createdAt")
	})
	View("default", func() {
		Attribute("id")
		Attribute("name")
		Attribute("leader")
		Attribute("createdAt")
	})
	View("link", func() {
		Attribute("id")
		Attribute("name")
	})
})
