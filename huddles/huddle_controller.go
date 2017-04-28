package huddles

//
// type HuddleSchedulerController struct {
// 	configs []HuddleConfig
// }
//
// func (h *HuddleSchedulerController) AddConfig(config *HuddleConfig) {
// 	h.configs = append(h.configs, *config)
// }
//
// func (h *HuddleSchedulerController) ScheduleHandler(c *gin.Context) {
// 	var scheduledHuddles []*ie.Huddle
// 	for i := range h.configs {
// 		hs := NewHuddleScheduler(&h.configs[i])
// 		huddles, err := hs.ScheduleHuddles()
// 		if err != nil {
// 			c.AbortWithError(http.StatusInternalServerError, err)
// 			return
// 		}
// 		scheduledHuddles = append(scheduledHuddles, huddles...)
// 	}
// 	c.JSON(http.StatusOK, scheduledHuddles)
// }
