package huddles

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HuddleSchedulerController struct {
	configs []HuddleConfig
}

func (h *HuddleSchedulerController) AddConfig(config *HuddleConfig) {
	h.configs = append(h.configs, *config)
}

func (h *HuddleSchedulerController) ScheduleHandler(c *gin.Context) {
	var scheduledHuddles []*Huddle
	for i := range h.configs {
		huddles, err := ScheduleHuddles(&h.configs[i])
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		scheduledHuddles = append(scheduledHuddles, huddles...)
	}
	c.JSON(http.StatusOK, scheduledHuddles)
}
