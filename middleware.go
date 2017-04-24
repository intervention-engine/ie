package main

import (
	"context"
	"net/http"

	"github.com/goadesign/goa"
	"github.com/intervention-engine/ie/mongo"
	"github.com/intervention-engine/ie/storage"
	mgo "gopkg.in/mgo.v2"
)

func WithMongoService(session *mgo.Session) goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			newctx := context.WithValue(ctx, "service", mongo.NewMongoService(session))
			return h(newctx, rw, req)
		}
	}
}

func GetStorageService(ctx context.Context) storage.Service {
	svc := ctx.Value("service")
	return svc.(storage.Service)
}
