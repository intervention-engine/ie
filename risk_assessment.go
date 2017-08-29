package main

import (
	"time"

	"github.com/goadesign/goa"
	"github.com/intervention-engine/ie/app"
)

// RiskAssessmentController implements the risk_assessment resource.
type RiskAssessmentController struct {
	*goa.Controller
}

// NewRiskAssessmentController creates a risk_assessment controller.
func NewRiskAssessmentController(service *goa.Service) *RiskAssessmentController {
	return &RiskAssessmentController{Controller: service.NewController("RiskAssessmentController")}
}

func toListView(ra *app.RiskAssessment) *app.RiskAssessmentList {
	return &app.RiskAssessmentList{Date: ra.Date, ID: ra.ID,
		RiskServiceID: ra.RiskServiceID, Value: ra.Value}
}

func allToListView(ra []*app.RiskAssessment) []*app.RiskAssessmentList {
	ral := make([]*app.RiskAssessmentList, len(ra))
	for i, r := range ra {
		ral[i] = toListView(r)
	}
	return ral
}

// List runs the list action.
func (c *RiskAssessmentController) List(ctx *app.ListRiskAssessmentContext) error {
	// RiskAssessmentController_List: start_implement
	s := GetServiceFactory(ctx.Context)
	rs := s.NewRiskAssessmentService()

	start := time.Time{}
	end := time.Now()

	// RiskAssessmentController_List: end_implement

	if ctx.StartDate != nil {
		start = *ctx.StartDate
	}

	if ctx.EndDate != nil {
		end = *ctx.EndDate
	}

	res, err := rs.RiskAssessments(ctx.ID, ctx.RiskServiceID, start, end)

	if err != nil {
		return ctx.InternalServerError(err)
	}
	return ctx.OKList(allToListView(res))
}
