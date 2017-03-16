package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Render handles errors and renders appropriate content
func Render(ctx *gin.Context, json map[string]interface{}, err error) {
	if err != nil {
		ctx.AbortWithError(ErrCode(err), err)
	} else {
		if json != nil {
			ctx.JSON(http.StatusOK, json)
		} else {
			ctx.Status(http.StatusNoContent)
		}

	}
}

// ErrCode Matches the service error to the corresponding HTTP Error Code
func ErrCode(err error) int {
	if err.Error() == "not found" {
		return http.StatusNotFound
	} else if err.Error() == "bad id" {
		return http.StatusBadRequest
	}

	return http.StatusInternalServerError
}
