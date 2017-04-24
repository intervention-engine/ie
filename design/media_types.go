package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var PatientMedia = MediaType("application/vnd.patient+json", func() {
	Description("A patient")
	Attributes(func() {
		Attribute("id", String, "Unique patient ID")
		Attribute("name", func() {
			Attribute("family", String, "Family Name")
			Attribute("given", String, "Given Name")
			Attribute("middleInitial", String, "Middle Initial")
			Attribute("full", String, "Full Name")
		})
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
	Description("A care team")
	Reference(CareTeamPayload)
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
