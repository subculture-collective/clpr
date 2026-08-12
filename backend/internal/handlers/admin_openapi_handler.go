package handlers

import (
	"net/http"

	adminopenapi "git.subcult.tv/subculture-collective/clpr/internal/openapi"
	"github.com/gin-gonic/gin"
)

type AdminOpenAPIHandler struct{}

func NewAdminOpenAPIHandler() *AdminOpenAPIHandler {
	return &AdminOpenAPIHandler{}
}

// Get serves the immutable OpenAPI document embedded in the running backend.
func (h *AdminOpenAPIHandler) Get(c *gin.Context) {
	c.Data(http.StatusOK, "application/json; charset=utf-8", adminopenapi.Document)
}
