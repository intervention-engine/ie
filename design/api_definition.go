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
	Host("localhost:3001")
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
