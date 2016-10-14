package huddles

import (
	"fmt"
	"net/http"
	"time"

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
		hs := NewHuddleScheduler(&h.configs[i])
		start := time.Now()
		huddles, err := hs.ScheduleHuddles()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		fmt.Println(time.Now().Sub(start))
		scheduledHuddles = append(scheduledHuddles, huddles...)
	}
	c.JSON(http.StatusOK, scheduledHuddles)
}
