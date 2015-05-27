package notifications

import "github.com/intervention-engine/fhir/models"

/* The NotificationDefinition interface should be implemented by all notification definitions */
type NotificationDefinition interface {
	Name() string
	Triggers(resource interface{}, action string) bool
	GetNotification(resource interface{}, action string, baseURL string) *models.CommunicationRequest
}

// Setup the registry that keeps track of all the notification definitions

type NotificationDefinitionRegistry struct {
	defs []NotificationDefinition
}

func (r *NotificationDefinitionRegistry) Register(def NotificationDefinition) {
	r.defs = append(r.defs, def)
}

func (r *NotificationDefinitionRegistry) RegisterAll(slice []NotificationDefinition) {
	r.defs = append(r.defs, slice...)
}

func (r *NotificationDefinitionRegistry) GetAll() []NotificationDefinition {
	return r.defs
}

var DefaultNotificationDefinitionRegistry = new(NotificationDefinitionRegistry)
