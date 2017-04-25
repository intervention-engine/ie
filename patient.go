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
			// return goa.ErrBadRequest("id was not a proper bson object id")
			return ctx.BadRequest()
		}
		if err.Error() == "not found" {
			// return goa.ErrNotFound("could not find resource")
			return ctx.NotFound()
		}
		// return goa.ErrInternal("internal server error trying to fulfill request")
		return ctx.InternalServerError()
	}

	return ctx.OK(p)
}

// List runs the list action.
func (c *PatientController) List(ctx *app.ListPatientContext) error {
	s := GetStorageService(ctx.Context)
	ps := s.NewPatientService()
	if ctx.Page != nil {
		// Time to do paging!
		sortby := "name.full"
		if ctx.SortBy != nil {
			sortby = *ctx.SortBy
		}
		list := strings.Split(sortby, ",")
		pp, err := ps.SortBy(list...)
		if err != nil {
			// return goa.ErrInternal("internal server error trying to list patients")
			return ctx.InternalServerError()
		}
		// grab the actual ones we need to send over the wire.
		total := len(pp)
		pageinfo, err := pagingInfo(*ctx.Page, *ctx.PerPage, total)
		if err != nil {
			// TODO: actually get error message
			return ctx.BadRequest()
		}
		links := linkInfo(pageinfo)
		ctx.ResponseWriter.Header().Set("Link", links)

		return ctx.OK(pp[pageinfo.dot:pageinfo.enddot])
	}
	pp, err := ps.Patients()
	if err != nil {
		// return goa.ErrInternal("internal server error trying to list patients")
		return ctx.InternalServerError()
	}
	return ctx.OK(pp)
}

type page struct {
	num    int
	per    int
	next   int
	prev   int
	first  int
	last   int
	total  int
	dot    int
	enddot int
}

func pagingInfo(num int, perpage int, total int) (page, error) {
	if perpage == 0 {
		perpage = 50
	}
	last := (total / perpage) + 1

	dot := (num * perpage) - perpage
	if dot > total {
		return page{}, goa.ErrBadRequest("requested page is out of bounds of what is available")
	}
	enddot := dot + perpage
	if enddot > total {
		// This is the last page, which will most likely go over the total patients.
		enddot = total
	}
	prev := 1
	if num != 1 {
		prev = num - 1
	}
	next := num + 1
	if num == last {
		next = last
	}

	return page{
		next:   next,
		prev:   prev,
		num:    num,
		total:  total,
		last:   last,
		first:  1,
		per:    perpage,
		dot:    dot,
		enddot: enddot,
	}, nil
}

// TODO: pass the context into linkInfo and make the URL dynamic.
var linkTmpl string = "<http://localhost:3001/patients?page=%d&per_page=%d>; rel=\"next\", <http://localhost:3001/patients?page=%d&per_page=%d>; rel=\"last\", <http://localhost:3001/patients?page=%d&per_page=%d>; rel=\"first\", <http://localhost:3001/patients?page=%d&per_page=%d>; rel=\"prev\", total=%d"

func linkInfo(info page) string {
	return fmt.Sprintf(linkTmpl, info.next, info.per, info.last, info.per, 1, info.per, info.prev, info.per, info.total)
}
