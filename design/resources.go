package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = Resource("patient", func() {
	DefaultMedia(PatientMedia)
	BasePath("/patients")
	Response(OK)
	Response(NotFound)
	Response(BadRequest)
	Response(InternalServerError)
	Action("show", func() {
		Description("Get patient by id.")
		Routing(GET("/:id"))
		Params(func() {
			Param("id", String, "Patient ID")
		})
	})
	Action("list", func() {
		Routing(GET(""))
		Params(func() {
			Param("sort_by", String)
			Param("page", Integer, func() {
				Minimum(1)
			})
			Param("per_page", Integer, func() {
				Minimum(1)
			})
		})

		Description("List all patients.")
		Response(OK, func() {
			Media(CollectionOf(PatientMedia, func() {
				View("default")
			}))
		})
	})
})

var _ = Resource("risk_service", func() {
	DefaultMedia(RiskServiceMedia)
	BasePath("/risk_services")
	Response(OK)
	Action("list", func() {
		Description("List all risk services")
		Routing(GET("/"))
		Response(NotFound, ErrorMedia)
		Response(BadRequest, ErrorMedia)
		Response(InternalServerError, ErrorMedia)
		Response(OK, func() {
			Media(CollectionOf(RiskServiceMedia))
		})
	})
})

var _ = Resource("risk_assessment", func() {
	DefaultMedia(RiskAssessmentMedia)
	BasePath("/patients/:id/risk_assessments")
	Action("list", func() {
		Description("List of risk assessments for a patient")
		Routing(GET("/"))
		Params(func() {
			Param("risk_service_id", String, "Scopes list to a single risk service")
			Param("start_date", DateTime, "Include all risk assessments after this date")
			Param("end_date", DateTime, "Include all risk assessments before this date")
			Required("risk_service_id")
		})
		Response(OK, func() {
			Media(CollectionOf(RiskAssessmentMedia, func() {
				View("default")
			}))
		})
		Response(BadRequest, ErrorMedia)
		Response(InternalServerError, ErrorMedia)

	})
})

var _ = Resource("risk_categories", func() {
	DefaultMedia(RiskCategoryMedia)
	BasePath("/risk_assessments/:id/breakdown")
	Response(BadRequest, ErrorMedia)
	Response(InternalServerError, ErrorMedia)
	Action("list", func() {
		Description("List of subcategories of risk score")
		Routing(GET("/"))
		Response(OK, func() {
			Media(CollectionOf(RiskCategoryMedia, func() {
				View("default")
			}))
		})
		Response(BadRequest, ErrorMedia)
		Response(InternalServerError, ErrorMedia)
	})
})

var _ = Resource("care_team", func() {
	DefaultMedia(CareTeamMedia)
	BasePath("/care_teams")
	Response(OK)
	Action("show", func() {
		Description("Get care team by id.")
		Routing(GET("/:id"))
		Params(func() {
			Param("id", String, "Care team ID")
		})
		Response(NotFound, ErrorMedia)
		Response(BadRequest, ErrorMedia)
		Response(InternalServerError, ErrorMedia)
	})
	Action("list", func() {
		Routing(GET(""))
		Description("List all care teams.")
		Response(OK, func() {
			Media(CollectionOf(CareTeamMedia, func() {
				View("default")
			}))
		})
		Response(NotFound, ErrorMedia)
		Response(BadRequest, ErrorMedia)
		Response(InternalServerError, ErrorMedia)
	})
	Action("create", func() {
		Routing(POST(""))
		Payload(CareTeamPayload, func() {
			Required("name", "leader")
		})
		Description("Create care team.")
		Response(Created)
		Response(NotFound, ErrorMedia)
		Response(BadRequest, ErrorMedia)
		Response(InternalServerError, ErrorMedia)
	})
	Action("update", func() {
		Routing(PUT("/:id"))
		Payload(CareTeamPayload)
		Description("Update care team.")
		Response(NoContent)
		Response(NotFound, ErrorMedia)
		Response(BadRequest, ErrorMedia)
		Response(InternalServerError, ErrorMedia)
	})
	Action("delete", func() {
		Routing(DELETE("/:id"))
		Description("Delete care team.")
		Response(NoContent)
		Response(NotFound, ErrorMedia)
		Response(BadRequest, ErrorMedia)
		Response(InternalServerError, ErrorMedia)
	})
	Action("huddles", func() {
		Routing(GET("/:id/huddles"))
		Description("List all huddles for this care team.")
		Response(OK, func() {
			Media(CollectionOf(HuddleMedia, func() {
				View("default") // This should be like a "careteam" view
			}))
		})
		Response(NotFound, ErrorMedia)
		Response(BadRequest, ErrorMedia)
		Response(InternalServerError, ErrorMedia)
	})
	Action("schedule", func() {
		Routing(POST("/:id/huddles"))
		Payload(SchedulePatientPayload, func() {
			Required("patient_id", "date")
		})
		Description("Schedule a patient for a huddle with this care team.")
		Response(OK, HuddleMedia)
		Response(Created, HuddleMedia)
		Response(NotFound, ErrorMedia)
		Response(BadRequest, ErrorMedia)
		Response(InternalServerError, ErrorMedia)
	})
	Action("add_patient", func() {
		Routing(POST("/:id/patients/:patient_id"))
		Description("Add patient to care team")
		Response(NoContent)
		Response(NotFound, ErrorMedia)
		Response(BadRequest, ErrorMedia)
		Response(InternalServerError, ErrorMedia)
	})
	Action("remove_patient", func() {
		Routing(DELETE("/:id/patients/:patient_id"))
		Description("Remove patient from care team.")
		Response(NoContent)
		Response(NotFound, ErrorMedia)
		Response(BadRequest, ErrorMedia)
		Response(InternalServerError, ErrorMedia)
	})
})

var _ = Resource("huddle", func() {
	DefaultMedia(HuddleMedia)
	BasePath("/huddles")
	Response(OK)
	Response(NotFound, ErrorMedia)
	Response(BadRequest, ErrorMedia)
	Response(InternalServerError, ErrorMedia)
	Action("cancel", func() {
		Routing(DELETE(":id/patients/:patient_id"))
		Description("Cancel patient's huddle")
		Response(OK, HuddleMedia)
		Response(NoContent)
	})
})

var _ = Resource("swagger", func() {
	Description("The API Swagger specification")
	Origin("*", func() {
		Methods("GET", "OPTIONS")
	})
	NoSecurity()
	Files("/swagger.json", "swagger/swagger.json")
	Files("/swagger-ui/*filepath", "swagger-ui/")
})
