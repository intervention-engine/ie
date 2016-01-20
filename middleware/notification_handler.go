package middleware

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/notifications"
	"github.com/labstack/echo"
)

type NotificationHandler struct {
	Registry *notifications.NotificationDefinitionRegistry
}

func (h *NotificationHandler) Handle() echo.MiddlewareFunc {
	return func(hf echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			hf(c)
			resourceType := c.Get("Resource")
			if resourceType != nil {
				rs := resourceType.(string)
				resource := c.Get(rs)
				actionType := c.Get("Action")

				var reg *notifications.NotificationDefinitionRegistry
				if h.Registry != nil {
					reg = h.Registry
				} else {
					reg = notifications.DefaultNotificationDefinitionRegistry
				}
				for _, def := range reg.GetAll() {
					if def.Triggers(resource, actionType.(string)) {
						notification := def.GetNotification(resource, actionType.(string), h.getBaseURL(c.Request()))
						err := server.Database.C("communicationrequests").Insert(notification)
						if err != nil {
							log.Printf("Error creating notification.\n\tNotification: %#v\n\tResource: %#v\n\tError: %#v", notification, resource, err)
							return err
						}
					}
				}
			}
			return nil
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
