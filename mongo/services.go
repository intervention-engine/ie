package mongo

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/ie"
	mgo "gopkg.in/mgo.v2"
)

const dbName = "fhir"

type Services struct {
	S *mgo.Session
}

// NewMongoService creates new Mongo service from MongodB URL and returns
// service and func used to clean up connection
func NewMongoService(url string) (Services, func()) {
	s, err := mgo.Dial(url)
	srv := Services{S: s}
	if err != nil {
		log.Fatalln("dialing mongo failed for ie session at url: %s", url)
	}

	return srv, srv.S.Close
}

// CareTeamService puts a mgo.Collection named "care_teams" into the context for the actual handler to use.
func (s *Services) CareTeamService() ie.Adapter {
	return func(h gin.HandlerFunc) gin.HandlerFunc {
		return func(ctx *gin.Context) {
			session := s.S.Copy()
			defer session.Close()
			c := col(session, "care_teams")
			service := &CareTeamService{C: c}
			ctx.Set("careTeamService", service)
			h(ctx)
		}
	}
}

func (s *Services) PatientService() ie.Adapter {
	return func(h gin.HandlerFunc) gin.HandlerFunc {
		return func(ctx *gin.Context) {
			session := s.S.Copy()
			defer session.Close()
			col := col(session, "patients")
			service := &PatientService{C: col}
			ctx.Set("patientService", service)
			h(ctx)
		}
	}
}

func (s *Services) MembershipService() ie.Adapter {
	return func(h gin.HandlerFunc) gin.HandlerFunc {
		return func(ctx *gin.Context) {
			session := s.S.Copy()
			defer session.Close()
			col := col(session, "care_team_memberships")
			service := &MembershipService{C: col}
			ctx.Set("membershipService", service)
		}
	}
}

func col(sess *mgo.Session, col string) *mgo.Collection {
	return sess.DB(dbName).C(col)
}
