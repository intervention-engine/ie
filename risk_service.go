package main

import (
	"github.com/goadesign/goa"
	"github.com/intervention-engine/ie/app"
)

func toListCollection(rs []*app.RiskService) []*app.RiskServiceList {
	rsl := make([]*app.RiskServiceList, len(rs))
	for i := 0; i < len(rsl); i++ {
		rsl[0] = toList(rs[0])
	}
	return rsl
}

func toList(rs *app.RiskService) *app.RiskServiceList {
	return &app.RiskServiceList{ID: rs.ID, Name: rs.Name}
}

// RiskServiceController implements the risk_service resource.
type RiskServiceController struct {
	*goa.Controller
}

// NewRiskServiceController creates a risk_service controller.
func NewRiskServiceController(service *goa.Service) *RiskServiceController {
	return &RiskServiceController{Controller: service.NewController("RiskServiceController")}
}

// List runs the list action.
func (c *RiskServiceController) List(ctx *app.ListRiskServiceContext) error {
	s := ctx.Value("riskServices").([]*app.RiskService)
	return ctx.OKList(toListCollection(s))
}
