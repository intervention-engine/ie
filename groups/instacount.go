package groups

import (
	"net/http"

	"github.com/gin-gonic/gin"
	fhir "github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/search"
	"github.com/intervention-engine/fhir/server"
)

func InstaCountAllHandler(c *gin.Context) {
	group := &fhir.Group{}
	if err := server.FHIRBind(c, group); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	cInfo, err := LoadCharacteristicInfo(group.Characteristic)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// TODO: Get the searcher that is actually used in the FHIR server (requires registration refactoring)
	searcher := search.NewMongoSearcher(server.Database)
	patients, conditions, encounters, err := resolveGroupCounts(cInfo, searcher)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newResultMap := map[string]int{
		"patients":   patients,
		"conditions": conditions,
		"encounters": encounters,
	}

	c.JSON(http.StatusOK, newResultMap)
}
