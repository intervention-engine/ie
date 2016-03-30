package huddles

import (
	"net/http"

	"github.com/intervention-engine/fhir/models"
	"github.com/labstack/echo"
)

type HuddleSchedulerController struct {
	configs []HuddleConfig
}

func (h *HuddleSchedulerController) AddConfig(config *HuddleConfig) {
	h.configs = append(h.configs, *config)
}

func (h *HuddleSchedulerController) ScheduleHandler(c *echo.Context) error {
	var scheduledHuddles []*models.Group
	for i := range h.configs {
		huddles, err := AutoScheduleHuddles(&h.configs[i])
		if err != nil {
			return err
		}
		scheduledHuddles = append(scheduledHuddles, huddles...)
	}
	return c.JSON(http.StatusOK, scheduledHuddles)
}
