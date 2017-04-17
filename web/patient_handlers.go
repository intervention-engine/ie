package web

import (
	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/ie"
)

// ListAllPatients List all patients
func ListAllPatients(ctx *gin.Context) {
	s := getPatientService(ctx)
	pp, err := s.Patients()
	Render(ctx, gin.H{"patients": pp}, err)
}

// ListAllCareTeamPatients that belong to a given care team
func ListAllCareTeamPatients(ctx *gin.Context) {
	id := ctx.Param("care_team_id")
	pp, err := patientsForCareTeam(ctx, id)
	Render(ctx, gin.H{"patients": pp}, err)
}

// AddPatientToCareTeam create a membership for a patient
func AddPatientToCareTeam(ctx *gin.Context) {
	ct := ctx.Param("care_team_id")
	p := ctx.Param("id")
	mem := ie.Membership{CareTeamID: ct, PatientID: p}
	err := getMembershipService(ctx).CreateMembership(mem)
	Render(ctx, gin.H{"membership": mem}, err)
}

// GetPatient Get a patient with given id.
func GetPatient(ctx *gin.Context) {
	s := getPatientService(ctx)
	id := ctx.Param("id")
	p, err := s.Patient(id)
	Render(ctx, gin.H{"patient": p}, err)
}

func getPatientService(ctx *gin.Context) ie.PatientService {
	svc := ctx.MustGet("patientService")
	return svc.(ie.PatientService)
}

func getMembershipService(ctx *gin.Context) ie.MembershipService {
	svc := ctx.MustGet("membershipService")
	return svc.(ie.MembershipService)
}

// PatientsForCareTeam Utility Function to get Patients that belong to a care team
func patientsForCareTeam(ctx *gin.Context, id string) ([]ie.Patient, error) {

	mems, err := getMembershipService(ctx).PatientMemberships(id)

	if err != nil {
		return nil, err
	}

	ids := make([]string, len(mems))

	for i, mem := range mems {
		ids[i] = mem.PatientID
	}

	pp, err := getPatientService(ctx).PatientsByID(ids)

	if err != nil {
		return nil, err
	}

	return pp, nil
}
