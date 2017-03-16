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

// GetPatient Get a patient with given id.
func GetPatient(ctx *gin.Context) {
	s := getPatientService(ctx)
	id := ctx.Param("id")
	p, err := s.Patient(id)
	Render(ctx, gin.H{"patient": p}, err)
}

func getPatientService(ctx *gin.Context) ie.PatientService {
	svc := ctx.MustGet("service")
	return svc.(ie.PatientService)
}
