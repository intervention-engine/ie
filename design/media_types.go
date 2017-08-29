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
		Attribute("care_team_id", String, "ID of the care team associated with this Huddle")
		Attribute("patients", ArrayOf(PatientHuddle), "List of patients scheduled for this huddle")
	})
	View("default", func() {
		Attribute("id", String, "Unique Huddle ID")
		Attribute("date", DateTime, "Creation timestamp")
		Attribute("care_team_id", String, "ID of the care team associated with this Huddle")
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
		Attribute("pie_id", String, "ID for the associated Risk breakdown the score is based on")
	})
	View("default", func() {
		Attribute("id", String, "Unique assessment identifier")
		Attribute("risk_service_id", String, "Identifier for risk service that produced the assessment")
		Attribute("date", DateTime, "Date assessment was created")
		Attribute("value", Integer, "Risk Score")
		Attribute("pie_id", String, "ID for the associated Risk breakdown the score is based on")
	})
	View("list", func() {
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
		Attribute("max_value", Integer, "Maximum possible value")
	})
	View("default", func() {
		Attribute("name", String, "Risk Category Name")
		Attribute("value", Integer, "Risk Category Value")
		Attribute("weight", Integer, "Weight On Overall Risk Value")
		Attribute("max_value", Integer, "Maximum possible value")
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
		Attribute("recent_risk_assessment", RiskAssessmentMedia)
		Attribute("age", Integer, "Age of Patient")
		Attribute("gender", String, "Gender of Patient")
		Attribute("birth_date", DateTime, "Birth Date of Patient")
		Attribute("current_conditions", ArrayOf(ActivePatientData))
		Attribute("current_medications", ArrayOf(ActivePatientData))
		Attribute("current_allergies", ArrayOf(ActivePatientData))
		Required("id", "name")
	})
	View("default", func() {
		Attribute("id")
		Attribute("name")
		Attribute("address")
		Attribute("recent_risk_assessment")
		Attribute("age")
		Attribute("gender")
		Attribute("birth_date")
		Attribute("current_conditions")
		Attribute("current_medications")
		Attribute("current_allergies")
	})
	View("list", func() {
		Attribute("id")
		Attribute("name")
		Attribute("address")
		Attribute("recent_risk_assessment")
		Attribute("age")
		Attribute("gender")
		Attribute("birth_date")
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
		Attribute("id", String, "Unique care team ID")
		Attribute("name")
		Attribute("leader")
		Attribute("created_at", DateTime, "Timestamp for care team creation")
	})
	View("default", func() {
		Attribute("id")
		Attribute("name")
		Attribute("leader")
		Attribute("created_at")
	})
	View("link", func() {
		Attribute("id")
		Attribute("name")
	})
})
