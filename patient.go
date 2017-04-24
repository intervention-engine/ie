package main

import (
	"fmt"
	"strings"

	"github.com/goadesign/goa"
	"github.com/intervention-engine/ie/app"
)

// PatientController implements the patient resource.
type PatientController struct {
	*goa.Controller
}

// NewPatientController creates a patient controller.
func NewPatientController(service *goa.Service) *PatientController {
	return &PatientController{Controller: service.NewController("PatientController")}
}

// Show runs the show action.
func (c *PatientController) Show(ctx *app.ShowPatientContext) error {
	s := GetStorageService(ctx.Context)
	ps := s.NewPatientService()
	p, err := ps.Patient(ctx.ID)
	if err != nil {
		if err.Error() == "bad id" {
			return goa.ErrBadRequest("id was not a proper bson object id")
		}
		if err.Error() == "not found" {
			return goa.ErrNotFound("could not find patient")
		}
		return goa.ErrInternal("internal server error trying to find patient")
	}

	return ctx.OK(p)
}

// List runs the list action.
func (c *PatientController) List(ctx *app.ListPatientContext) error {
	s := GetStorageService(ctx.Context)
	ps := s.NewPatientService()
	if ctx.Page != nil {
		// Time to do paging!
		page := *ctx.Page
		perpage := 50
		if ctx.PerPage != nil {
			perpage = *ctx.PerPage
		}
		sortby := "name.full"
		if ctx.SortBy != nil {
			sortby = *ctx.SortBy
		}
		list := strings.Split(sortby, ",")
		pp, err := ps.SortBy(list...)
		if err != nil {
			return goa.ErrInternal("internal server error trying to list patients")
		}
		// grab the actual ones we need to send over the wire.
		total := len(pp)
		last := (total / perpage) + 1
		dot := (page * perpage) - perpage
		if dot > total {
			return goa.ErrBadRequest("requested page is out of bounds of what is available")
		}
		enddot := dot + perpage
		if enddot > total {
			// This is the last page, which will most likely go over the total patients.
			enddot = total
		}

		// TODO: remove hard links
		var linkTmpl string
		linkTmpl = "<http://localhost:3001/patients?page=%d&per_page=%d>; rel=\"next\", <http://localhost:3001/patients?page=%d&per_page=%d>; rel=\"last\", <http://localhost:3001/patients?page=%d&per_page=%d>; rel=\"first\", <http://localhost:3001/patients?page=%d&per_page=%d>; rel=\"prev\", total=%d"
		links := fmt.Sprintf(linkTmpl, page+1, perpage, last, perpage, 1, perpage, page-1, perpage, total)
		ctx.ResponseWriter.Header().Set("Link", links)

		return ctx.OK(pp[dot:enddot])
	}
	pp, err := ps.Patients()
	if err != nil {
		return goa.ErrInternal("internal server error trying to list patients")
	}
	return ctx.OK(pp)
}
