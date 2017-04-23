package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = API("api", func() {
	Title("The Intervention Engine Web API")
	Description("An api used to interact with Intervention Engine")
	License(func() {
		Name("Apache 2.0")
		URL("https://github.com/intervention-engine/ie/blob/master/LICENSE")
	})
	Scheme("http")
	Host("localhost:8080")
	BasePath("/api")
	ResponseTemplate("RNotFound", func() {
		Description("Resource not found")
		Status(404)
	})
	// TODO: might not need this template
	ResponseTemplate("RCreated", func() {
		Description("Resource created")
		Status(201)
		Headers(func() {
			Header("Link", func() {
			})
		})
	})
})

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

var _ = Resource("care_team", func() {
	DefaultMedia(CareTeamMedia)
	BasePath("/care_teams")
	Response(OK)
	Response(NotFound, ErrorMedia)
	Response(BadRequest, ErrorMedia)
	Response(InternalServerError)
	Action("show", func() {
		Description("Get care team by id.")
		Routing(GET("/:id"))
		Params(func() {
			Param("id", String, "Care team ID")
		})
	})
	Action("list", func() {
		Routing(
			GET(""),
		)
		Description("List all care teams.")
		Response(OK, func() {
			Media(CollectionOf(CareTeamMedia, func() {
				View("default")
			}))
		})
	})
	Action("create", func() {
		Routing(
			POST(""),
		)
		Description("Create care team.")
	})
	Action("update", func() {
		Routing(
			PUT(""),
		)
		Description("Update care team.")
	})
	Action("delete", func() {
		Routing(
			DELETE(""),
		)
		Description("Delete care team.")
	})
})

var _ = Resource("swagger", func() {
	Description("The API Swagger specification")
	Files("/swagger.json", "swagger/swagger.json")
	Files("/swagger-ui/*filepath", "swagger-ui/dist/")
})

var RiskAssessment = Type("riskAssessment", func() {
	Attribute("id", String, "Risk Assessment ID")
	Attribute("groupId", String, "Risk Assessment Group ID")
	Attribute("date", DateTime, "Date")
	Attribute("value", Integer, "Value")
})

var Address = Type("address", func() {
	Attribute("street", ArrayOf(String), "Street Name")
	Attribute("city", String, "City Name")
	Attribute("state", String, "State Name")
	Attribute("postalCode", String, "Postal Code")
})

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
	Attributes(func() {
		Attribute("id", String, "Unique care team ID")
		Attribute("name", String, "Care team name")
		Attribute("leader", String, "Care team leader")
		Attribute("createdAt", DateTime, "Timestamp for care team creation")
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
