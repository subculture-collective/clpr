package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/config"
)

// ConfigHandler handles public configuration endpoints
type ConfigHandler struct {
	cfg *config.Config
}

// NewConfigHandler creates a new config handler
func NewConfigHandler(cfg *config.Config) *ConfigHandler {
	return &ConfigHandler{
		cfg: cfg,
	}
}

// PublicConfigResponse represents public configuration data exposed to frontend
type PublicConfigResponse struct {
	Karma KarmaConfigResponse `json:"karma"`
}

// KarmaConfigResponse represents public karma configuration
type KarmaConfigResponse struct {
	InitialKarmaPoints        int  `json:"initial_karma_points"`
	SubmissionKarmaRequired   int  `json:"submission_karma_required"`
	RequireKarmaForSubmission bool `json:"require_karma_for_submission"`
}

// GetPublicConfig returns public configuration for the frontend
// GET /api/v1/config
func (h *ConfigHandler) GetPublicConfig(c *gin.Context) {
	c.JSON(http.StatusOK, PublicConfigResponse{
		Karma: KarmaConfigResponse{
			InitialKarmaPoints:        h.cfg.Karma.InitialKarmaPoints,
			SubmissionKarmaRequired:   h.cfg.Karma.SubmissionKarmaRequired,
			RequireKarmaForSubmission: h.cfg.Karma.RequireKarmaForSubmission,
		},
	})
}
