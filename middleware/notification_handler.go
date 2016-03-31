package middleware

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/notifications"
)

type NotificationHandler struct {
	Registry *notifications.NotificationDefinitionRegistry
}

func (h *NotificationHandler) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.IsAborted() {
			return
		}
		if c.Request.Method != "POST" {
			return
		}

		if resourceType, ok := c.Get("Resource"); ok {
			rs := resourceType.(string)
			resource, ok := c.Get(rs)
			if !ok {
				log.Printf("Error creating notification for resource type: %#v", resourceType)
				return
			}
			actionType, ok := c.Get("Action")
			if !ok {
				log.Printf("Error creating notification for resource: %#v", resource)
				return
			}

			var reg *notifications.NotificationDefinitionRegistry
			if h.Registry != nil {
				reg = h.Registry
			} else {
				reg = notifications.DefaultNotificationDefinitionRegistry
			}
			for _, def := range reg.GetAll() {
				if def.Triggers(resource, actionType.(string)) {
					notification := def.GetNotification(resource, actionType.(string), h.getBaseURL(c.Request))
					if err := server.Database.C("communicationrequests").Insert(notification); err != nil {
						log.Printf("Error creating notification.\n\tNotification: %#v\n\tResource: %#v\n\tError: %#v", notification, resource, err)
						return
					}
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
