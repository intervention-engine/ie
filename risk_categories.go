package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/goadesign/goa"
	"github.com/intervention-engine/ie/app"
)

// RiskCategoriesController implements the risk_categories resource.
type RiskCategoriesController struct {
	*goa.Controller
}

// NewRiskCategoriesController creates a risk_categories controller.
func NewRiskCategoriesController(service *goa.Service) *RiskCategoriesController {
	return &RiskCategoriesController{Controller: service.NewController("RiskCategoriesController")}
}

// List runs the list action.
func (c *RiskCategoriesController) List(ctx *app.ListRiskCategoriesContext) error {

	s := GetServiceFactory(ctx.Context)
	ras := s.NewRiskAssessmentService()

	ra, err := ras.RiskAssessment(ctx.ID)

	if err != nil {
		ctx.InternalServerError(err)
	}

	if ra == nil {
		ctx.BadRequest(errors.New("Invalid Risk Assessment ID"))
	}

	rs := GetRiskService(ctx, *ra.RiskServiceID)

	// Put your logic here
	pieURL, err := url.Parse("pies/" + ctx.ID)

	if err != nil {
		ctx.BadRequest(err)
	}

	rsURL, err := url.Parse(*rs.URL)

	if err != nil {
		return ctx.InternalServerError(err)
	}

	endpoint := rsURL.ResolveReference(pieURL)
	resp, err := http.Get(endpoint.String())

	if err != nil {
		return ctx.InternalServerError(err)
	}

	var pie app.Pie

	pieJSON, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return ctx.InternalServerError(err)
	}

	json.Unmarshal(pieJSON, &pie)

	// RiskBreakdownController_List: end_implement

	rc := pieToBreakdown(pie)

	return ctx.OK(rc)
}

func pieToBreakdown(pie app.Pie) []*app.RiskCategory {
	c := make([]*app.RiskCategory, len(pie.Slices))
	for i, s := range pie.Slices {
		c[i] = &app.RiskCategory{MaxValue: s.MaxValue, Value: s.Value, Weight: s.Weight, Name: s.Name}
	}
	return c
}
