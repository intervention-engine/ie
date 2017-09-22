package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/goadesign/goa"
	"github.com/intervention-engine/ie/app"
	"github.com/intervention-engine/ie/storage"
)

// PatientController implements the patient resource.
type PatientController struct {
	*goa.Controller
}

// NewPatientController creates a patient controller.
func NewPatientController(service *goa.Service) *PatientController {
	return &PatientController{Controller: service.NewController("PatientController")}
}

func (c *PatientController) allToList(ps []*app.Patient) []*app.PatientList {
	p2 := make([]*app.PatientList, len(ps))
	for i := range ps {
		p2[i] = c.toList(ps[i])
	}
	return p2
}

func (c *PatientController) toList(p *app.Patient) *app.PatientList {
	return &app.PatientList{Address: p.Address, Age: p.Age,
		BirthDate: p.BirthDate, Gender: p.Gender, ID: p.ID, Name: p.Name,
		RecentRiskAssessment: p.RecentRiskAssessment, NextHuddle: p.NextHuddle}
}

// Show runs the show action.
func (c *PatientController) Show(ctx *app.ShowPatientContext) error {
	s := GetServiceFactory(ctx.Context)
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
	s := GetServiceFactory(ctx.Context)
	ps := s.NewPatientService()
	filter := make(map[string]string)
	if ctx.CareTeamID != nil {
		filter["care_team_id"] = *ctx.CareTeamID
	}
	if ctx.HuddleID != nil {
		filter["huddle_id"] = *ctx.HuddleID
	}
	if ctx.SearchTerm != nil {
		filter["search_term"] = *ctx.SearchTerm
	}
	var pp []*app.Patient
	var err error
	if ctx.SortBy != nil {
		log.Println("sort_by is: ", *ctx.SortBy)
		query, err := c.parseSortQuery(*ctx.SortBy)
		if err != nil {
			// err has good error messages
			return ctx.BadRequest()
		}
		log.Println("query is: ", query)
		pp, err = ps.PatientsSortBy(filter, query...)
		if err != nil {
			// "internal server error trying to list patients"
			return ctx.InternalServerError()
		}
	} else {
		pp, err = ps.Patients(filter)
		if err != nil {
			// "internal server error trying to list patients"
			return ctx.InternalServerError()
		}
	}

	if ctx.Page != nil {
		// Time to do paging!
		// grab the actual ones we need to send over the wire.
		total := len(pp)
		var perpage int
		if ctx.PerPage != nil {
			perpage = *ctx.PerPage
		}
		pageinfo, err := pagingInfo(*ctx.Page, perpage, total)
		if err != nil {
			// TODO: actually get error message
			return ctx.BadRequest()
		}
		links := linkInfo(pageinfo)
		ctx.ResponseWriter.Header().Set("Link", links)

		return ctx.OKList(c.allToList(pp[pageinfo.dot:pageinfo.enddot]))
	}
	return ctx.OKList(c.allToList(pp))
}

func (c *PatientController) parseSortQuery(query string) ([]string, error) {
	if query == "" {
		log.Println("sortby string was empty")
		return nil, errors.New("sortby string was empty")
	}
	q := strings.Split(query, ",")
	if len(q) > 2 {
		log.Println("sortby string had more than two fields to sort by")
		return nil, errors.New("sortby string had more than two fields to sort by")
	}
	valid := false
	for _, arg := range q {
		for _, goodField := range storage.AcceptedQueryFields {
			if arg == goodField {
				valid = true
			}
		}
		if !valid {
			log.Println("sortby string contained invalid field: ", arg)
			return nil, errors.New("sortby string contained invalid field: " + arg)
		}
		valid = false
	}
	return q, nil
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

const defaultPerPage = 50

func pagingInfo(num int, perpage int, total int) (page, error) {
	if perpage == 0 {
		perpage = defaultPerPage
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
