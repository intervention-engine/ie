package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/goadesign/goa"
	"github.com/intervention-engine/ie/app"
	"github.com/intervention-engine/ie/storage"
)

// EventController implements the event resource.
type EventController struct {
	*goa.Controller
}

// NewEventController creates an event controller.
func NewEventController(service *goa.Service) *EventController {
	return &EventController{Controller: service.NewController("EventController")}
}

func (c *EventController) List(ctx *app.ListEventContext) error {
	s := GetServiceFactory(ctx.Context)
	es := s.NewEventService()
	var query storage.EventFilterQuery
	query.PatientID = ctx.ID
	var err error
	log.Println(ctx.Type)
	if ctx.Type == nil || *ctx.Type == "" {
		query.Type = "condition,medication,encounter,risk_change"
	} else {
		query.Type = *ctx.Type
	}
	if strings.Contains(query.Type, "risk_change") {
		invalid := false
		message := "risk_change was one of the types, but did not provide:"
		if ctx.RiskServiceID == nil {
			invalid = true
			message = fmt.Sprintf(message+" %s", "risk_service_id")
		}
		if ctx.StartTime == nil {
			invalid = true
			message = fmt.Sprintf(message+" %s", "start_time")
		}
		if invalid {
			return ctx.BadRequest(goa.ErrBadRequest(message))
		}
		query.RiskServiceID = *ctx.RiskServiceID
	}
	if ctx.StartTime != nil {
		query.Start = *ctx.StartTime
	}
	if ctx.EndTime != nil {
		query.End = *ctx.EndTime
	}
	ee, err := es.EventsFilterBy(query)
	if err != nil {
		return ctx.InternalServerError(goa.ErrInternal("error trying to get events for patient", "error", err))
	}
	return ctx.OK(ee)
}
