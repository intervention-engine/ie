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
		return ctx.InternalServerError(err)
	}

	if ra == nil {
		return ctx.BadRequest(errors.New("Invalid Risk Assessment ID"))
	}

	rs := GetRiskService(ctx, *ra.RiskServiceID)

	pieURL, err := url.Parse("pies/" + *ra.PieID)

	if err != nil {
		return ctx.BadRequest(err)
	}

	if rs == nil {
		return ctx.InternalServerError(errors.New("Unable to Locate Risk Service for Risk Assessment"))
	}

	rsURL, err := url.Parse(*rs.URL)

	if err != nil {
		return ctx.InternalServerError(err)
	}

	endpoint := rsURL.ResolveReference(pieURL)

	resp, err := http.Get(endpoint.String())

	if resp.StatusCode == 404 {
		return goa.NewErrorClass("bad_gateway", 502)("Unable to Reach Risk Server")
	}

	if err != nil {
		return ctx.InternalServerError(err)
	}

	var pie app.Pie

	pieJSON, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return ctx.InternalServerError(err)
	}

	json.Unmarshal(pieJSON, &pie)

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
