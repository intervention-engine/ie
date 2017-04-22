package ie

import (
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

// Get runs the get action.
func (c *PatientController) Show(ctx *app.ShowPatientContext) error {
	s := GetStorageService(ctx.Context)
	ps := s.NewPatientService()
	p, err := ps.Patient(ctx.ID)
	if err != nil {
		if err.Error() == "bad id" {
			return ctx.BadRequest(err)
		}
		return ctx.NotFound()
	}

	return ctx.OK(p)
}

// List runs the list action.
func (c *PatientController) List(ctx *app.ListPatientContext) error {
	s := GetStorageService(ctx.Context)
	ps := s.NewPatientService()
	pp, err := ps.Patients()
	if err != nil {
		return ctx.BadRequest(err)
	}
	return ctx.OK(pp)
}
