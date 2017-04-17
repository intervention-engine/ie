package ie

import "github.com/gin-gonic/gin"

type Adapter func(gin.HandlerFunc) gin.HandlerFunc

func Adapt(h gin.HandlerFunc, adapters ...Adapter) gin.HandlerFunc {
	for _, wrapper := range adapters {
		h = wrapper(h)
	}
	return h
}

// Services Interface for services to be shared with the web handlers
type Services interface {
	CareTeamService() Adapter
	PatientService() Adapter
	MembershipService() Adapter
}
