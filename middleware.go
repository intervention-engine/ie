package main

import (
	"context"
	"net/http"

	"github.com/goadesign/goa"
	"github.com/intervention-engine/ie/mongo"
	"github.com/intervention-engine/ie/storage"
	mgo "gopkg.in/mgo.v2"
)

func exposeHeaderField(field string) goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			rw.Header().Set("Access-Control-Expose-Headers", field)
			return h(ctx, rw, req)
		}
	}
}

func withMongoService(session *mgo.Session) goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			newctx := context.WithValue(ctx, "serviceFactory", mongo.NewServiceFactory(session, "fhir"))
			return h(newctx, rw, req)
		}
	}
}

func withRiskServices(path string) goa.Middleware {
	riskServices := loadRiskServicesJSON(path)
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			newctx := context.WithValue(ctx, "riskServices", riskServices)
			return h(newctx, rw, req)
		}
	}
}

func GetServiceFactory(ctx context.Context) storage.ServiceFactory {
	svc := ctx.Value("serviceFactory")
	return svc.(storage.ServiceFactory)
}
