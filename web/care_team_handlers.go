package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/ie"
)

// ListAllCareTeams List all care teams
func ListAllCareTeams(ctx *gin.Context) {
	s := getCareTeamService(ctx)
	cc, err := s.CareTeams()

	Render(ctx, gin.H{"care_teams": cc}, err)
}

// GetCareTeam Get a care team with given id.
func GetCareTeam(ctx *gin.Context) {
	s := getCareTeamService(ctx)

	id := ctx.Param("id")
	ct, err := s.CareTeam(id)

	Render(ctx, gin.H{"care_team": ct}, err)
}

// CreateCareTeam Create a care team
func CreateCareTeam(ctx *gin.Context) {
	s := getCareTeamService(ctx)

	ct := &ie.CareTeam{}

	if bindFormData(ctx, ct) {
		err := s.CreateCareTeam(ct)
		Render(ctx, gin.H{"care_team": ct}, err)
	}
}

// UpdateCareTeam Update a care team with given id
func UpdateCareTeam(ctx *gin.Context) {
	s := getCareTeamService(ctx)

	id := ctx.Param("id")
	ct, err := s.CareTeam(id)

	if err != nil {
		status := ErrCode(err)
		ctx.AbortWithError(status, err)
		return
	}

	if bindFormData(ctx, ct) {
		err = s.UpdateCareTeam(ct)
		Render(ctx, gin.H{"care_team": ct}, err)
	}
}

// DeleteCareTeam Delete a care team
func DeleteCareTeam(ctx *gin.Context) {
	s := getCareTeamService(ctx)

	id := ctx.Param("id")
	err := s.DeleteCareTeam(id)

	Render(ctx, nil, err)
}

func getCareTeamService(ctx *gin.Context) ie.CareTeamService {
	svc := ctx.MustGet("service")
	return svc.(ie.CareTeamService)
}

func bindFormData(ctx *gin.Context, ct *ie.CareTeam) bool {
	var form ie.CareTeam

	if ctx.Bind(&form) != nil {
		ct.Name = form.Name
		ct.Leader = form.Leader
		return true
	}
	ctx.AbortWithError(http.StatusBadRequest, nil)
	return false
}
