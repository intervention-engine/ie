package middleware

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/context"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/notifications"
)

type NotificationHandler struct {
	Registry *notifications.NotificationDefinitionRegistry
}

func (h *NotificationHandler) Handle(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	next(rw, r)
	resourceType, ok := context.GetOk(r, "Resource")
	if ok {
		resource := context.Get(r, resourceType)
		actionType := context.Get(r, "Action")

		var reg *notifications.NotificationDefinitionRegistry
		if h.Registry != nil {
			reg = h.Registry
		} else {
			reg = notifications.DefaultNotificationDefinitionRegistry
		}
		for _, def := range reg.GetAll() {
			if def.Triggers(resource, actionType.(string)) {
				notification := def.GetNotification(resource, actionType.(string), h.getBaseURL(r))
				err := server.Database.C("communicationrequests").Insert(notification)
				if err != nil {
					log.Printf("Error creating notification.\n\tNotification: %#v\n\tResource: %#v\n\tError: %#v", notification, resource, err)
					http.Error(rw, err.Error(), http.StatusInternalServerError)
				}
			}
		}
	}
}

func (h *NotificationHandler) getBaseURL(r *http.Request) string {
	newURL := url.URL(*r.URL)
	if newURL.Host == "" {
		newURL.Host = r.Host
	}
	if newURL.Scheme == "" {
		if strings.HasSuffix(newURL.Host, ":443") {
			newURL.Scheme = "https"
		} else {
			newURL.Scheme = "http"
		}
	}
	newURL.Path = ""

	return newURL.String()
}
