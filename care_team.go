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
	// CareTeamController_Create: start_implement

	// Put your logic here

	// CareTeamController_Create: end_implement
	res := &app.Careteam{}
	return ctx.OK(res)
}

// Delete runs the delete action.
func (c *CareTeamController) Delete(ctx *app.DeleteCareTeamContext) error {
	// CareTeamController_Delete: start_implement

	// Put your logic here

	// CareTeamController_Delete: end_implement
	res := &app.Careteam{}
	return ctx.OK(res)
}

// List runs the list action.
func (c *CareTeamController) List(ctx *app.ListCareTeamContext) error {
	// CareTeamController_List: start_implement

	// Put your logic here

	// CareTeamController_List: end_implement
	res := app.CareteamCollection{}
	return ctx.OK(res)
}

// Show runs the show action.
func (c *CareTeamController) Show(ctx *app.ShowCareTeamContext) error {

	res := &app.Careteam{}
	return ctx.OK(res)
}

// Update runs the update action.
func (c *CareTeamController) Update(ctx *app.UpdateCareTeamContext) error {
	// CareTeamController_Update: start_implement

	// Put your logic here

	// CareTeamController_Update: end_implement
	res := &app.Careteam{}
	return ctx.OK(res)
}
