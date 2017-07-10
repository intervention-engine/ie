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

var RiskAssessmentMedia = MediaType("application/vnd.riskassessment+json", func() {
	TypeName("RiskAssessment")
	Description("An single overall assessment score of a patient's risk for a risk service")
	ContentType("application/json")
	Attributes(func() {
		Attribute("id", String, "Unique assessment identifier")
		Attribute("risk_service_id", String, "Identifier for risk service that produced the assessment")
		Attribute("date", DateTime, "Date assessment was created")
		Attribute("value", Number, "Risk Score")
	})
	View("default", func() {
		Attribute("id", String, "Unique assessment identifier")
		Attribute("risk_service_id", String, "Identifier for risk service that produced the assessment")
		Attribute("date", DateTime, "Date assessment was created")
		Attribute("value", Integer, "Risk Score")
	})
})

var RiskCategoryMedia = MediaType("applicaiton/vnd.riskassessment+json", func() {
	TypeName("RiskCategory")
	Description("A component score of an overall risk asessment")
	ContentType("application/json")
	Attributes(func() {
		Attribute("name", String, "Risk Category Name")
		Attribute("value", Integer, "Risk Category Value")
		Attribute("weight", Integer, "Weight On Overall Risk Value")
		Attribute("maxValue", Integer, "Maximum possible value")
	})
	View("default", func() {
		Attribute("name", String, "Risk Category Name")
		Attribute("value", Integer, "Risk Category Value")
		Attribute("weight", Integer, "Weight On Overall Risk Value")
		Attribute("maxValue", Integer, "Maximum possible value")
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
		Attribute("riskAssessments", ArrayOf(RiskAssessmentMedia))
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

var RiskServiceMedia = MediaType("applicatoin/vnd.riskservice+json", func() {
	TypeName("RiskService")
	Description("Service providing risk scores for patients")
	ContentType("application/json")
	Attributes(func() {
		Attribute("id")
		Attribute("name")
		Attribute("url")
	})
	View("default", func() {
		Attribute("id")
		Attribute("name")
		Attribute("url")
	})
	View("list", func() {
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
