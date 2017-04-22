package ie

import (
	"github.com/goadesign/goa"
	"github.com/intervention-engine/ie/app"
)

// CareTeamController implements the care_team resource.
type CareTeamController struct {
	*goa.Controller
}

// NewCareTeamController creates a care_team controller.
func NewCareTeamController(service *goa.Service) *CareTeamController {
	return &CareTeamController{Controller: service.NewController("CareTeamController")}
}

// Create runs the create action.
func (c *CareTeamController) Create(ctx *app.CreateCareTeamContext) error {
	s := GetStorageService(ctx.Context)
	cs := s.NewCareTeamService()

	return ctx.OK(res)
}

// Delete runs the delete action.
func (c *CareTeamController) Delete(ctx *app.DeleteCareTeamContext) error {
	res := &app.Careteam{}
	return ctx.OK(res)
}

// List runs the list action.
func (c *CareTeamController) List(ctx *app.ListCareTeamContext) error {
	s := GetStorageService(ctx.Context)
	cs := s.NewCareTeamService()
	cc, err := cs.CareTeams()
	if err != nil {
		return ctx.BadRequest(err)
	}

	return ctx.OK(cc)
}

// Show runs the show action.
func (c *CareTeamController) Show(ctx *app.ShowCareTeamContext) error {
	s := GetStorageService(ctx.Context)
	cs := s.NewCareTeamService()
	c, err := cs.CareTeam(ctx.ID)
	if err != nil {
		if err.Error() == "bad id" {
			return ctx.BadRequest(err)
		}
		return ctx.NotFound()
	}

	return ctx.OK(c)
}

// Update runs the update action.
func (c *CareTeamController) Update(ctx *app.UpdateCareTeamContext) error {
	res := &app.Careteam{}
	return ctx.OK(res)
}
