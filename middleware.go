package ie

import (
	"context"
	"net/http"

	"github.com/goadesign/goa"
	"github.com/intervention-engine/ie/mongo"
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

func GetStorageService(ctx context.Context) StorageService {
	svc := ctx.Value("service")
	return svc.(StorageService)
}
