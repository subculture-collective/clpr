package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// AdHandler handles ad delivery endpoints
type AdHandler struct {
	adService *services.AdService
}

// NewAdHandler creates a new AdHandler
func NewAdHandler(adService *services.AdService) *AdHandler {
	return &AdHandler{
		adService: adService,
	}
}

// parseDaysParameter parses and validates the days query parameter
// Returns the number of days and true if valid, or 0 and false if invalid (error already sent)
func parseDaysParameter(c *gin.Context, defaultDays int) (int, bool) {
	days := defaultDays
	if daysStr := c.Query("days"); daysStr != "" {
		d, err := strconv.Atoi(daysStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "INVALID_PARAMETER",
					Message: "days parameter must be a valid integer",
				},
			})
			return 0, false
		}
		if d <= 0 || d > 365 {
			c.JSON(http.StatusBadRequest, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "INVALID_PARAMETER",
					Message: "days parameter must be between 1 and 365",
				},
			})
			return 0, false
		}
		days = d
	}
	return days, true
}

// SelectAd handles GET /ads/select
// Selects an appropriate ad for display based on request parameters
func (h *AdHandler) SelectAd(c *gin.Context) {
	var req models.AdSelectionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// Get user ID if authenticated
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	// Get IP address for fraud prevention
	ipAddress := c.ClientIP()

	// Select an ad
	response, err := h.adService.SelectAd(c.Request.Context(), req, userID, ipAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "AD_SELECTION_FAILED",
				Message: "Failed to select ad",
			},
		})
		return
	}

	// If no ad available, return empty response with consistent structure
	if response.Ad == nil {
		c.JSON(http.StatusOK, StandardResponse{
			Success: true,
			Data:    models.AdSelectionResponse{},
		})
		return
	}

	// Return ad with impression tracking info
	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"ad":            response.Ad,
			"impression_id": response.ImpressionID,
			"tracking_url":  response.TrackingURL,
		},
	})
}

// TrackImpression handles POST /ads/track/:id
// Updates impression with viewability and click tracking data
func (h *AdHandler) TrackImpression(c *gin.Context) {
	impressionID := c.Param("id")
	if impressionID == "" {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_IMPRESSION_ID",
				Message: "Impression ID is required",
			},
		})
		return
	}

	// Validate UUID format
	if _, err := uuid.Parse(impressionID); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_IMPRESSION_ID",
				Message: "Invalid impression ID format",
			},
		})
		return
	}

	var req models.AdTrackingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid tracking data",
			},
		})
		return
	}

	// Set the impression ID from URL
	req.ImpressionID = impressionID

	// Track the impression
	if err := h.adService.TrackImpression(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "TRACKING_FAILED",
				Message: "Failed to track impression",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"message": "Tracking recorded",
		},
	})
}

// GetAd handles GET /ads/:id
// Returns ad details by ID
func (h *AdHandler) GetAd(c *gin.Context) {
	adID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_AD_ID",
				Message: "Invalid ad ID format",
			},
		})
		return
	}

	ad, err := h.adService.GetAdByID(c.Request.Context(), adID)
	if err != nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "AD_NOT_FOUND",
				Message: "Ad not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    ad,
	})
}

// GetCTRReportByCampaign handles GET /ads/analytics/campaigns
// Returns CTR report grouped by campaign/ad
func (h *AdHandler) GetCTRReportByCampaign(c *gin.Context) {
	days, ok := parseDaysParameter(c, 30)
	if !ok {
		return
	}

	since := time.Now().AddDate(0, 0, -days)
	reports, err := h.adService.GetCTRReportByCampaign(c.Request.Context(), since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "ANALYTICS_FAILED",
				Message: "Failed to get campaign CTR report",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"reports": reports,
			"days":    days,
		},
	})
}

// GetCTRReportBySlot handles GET /ads/analytics/slots
// Returns CTR report grouped by ad slot
func (h *AdHandler) GetCTRReportBySlot(c *gin.Context) {
	days, ok := parseDaysParameter(c, 30)
	if !ok {
		return
	}

	since := time.Now().AddDate(0, 0, -days)
	reports, err := h.adService.GetCTRReportBySlot(c.Request.Context(), since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "ANALYTICS_FAILED",
				Message: "Failed to get slot CTR report",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"reports": reports,
			"days":    days,
		},
	})
}

// GetExperimentReport handles GET /ads/experiments/:id/report
// Returns analytics for a specific experiment
func (h *AdHandler) GetExperimentReport(c *gin.Context) {
	experimentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_EXPERIMENT_ID",
				Message: "Invalid experiment ID format",
			},
		})
		return
	}

	days, ok := parseDaysParameter(c, 30)
	if !ok {
		return
	}

	since := time.Now().AddDate(0, 0, -days)
	report, err := h.adService.GetExperimentReport(c.Request.Context(), experimentID, since)
	if err != nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "EXPERIMENT_NOT_FOUND",
				Message: "Experiment not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"report": report,
			"days":   days,
		},
	})
}

// ListExperiments handles GET /ads/experiments
// Returns all running experiments
func (h *AdHandler) ListExperiments(c *gin.Context) {
	experiments, err := h.adService.GetRunningExperiments(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "LIST_EXPERIMENTS_FAILED",
				Message: "Failed to list experiments",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"experiments": experiments,
		},
	})
}

// Campaign CRUD Handlers

// CreateCampaignRequest represents the request to create a campaign
type CreateCampaignRequest struct {
	Name              string                 `json:"name" binding:"required"`
	AdvertiserName    string                 `json:"advertiser_name" binding:"required"`
	AdType            string                 `json:"ad_type" binding:"required,oneof=banner video native"`
	ContentURL        string                 `json:"content_url" binding:"required,url"`
	ClickURL          *string                `json:"click_url,omitempty"`
	AltText           *string                `json:"alt_text,omitempty"`
	Width             *int                   `json:"width,omitempty"`
	Height            *int                   `json:"height,omitempty"`
	Priority          int                    `json:"priority,omitempty"`
	Weight            int                    `json:"weight,omitempty"`
	DailyBudgetCents  *int64                 `json:"daily_budget_cents,omitempty"`
	TotalBudgetCents  *int64                 `json:"total_budget_cents,omitempty"`
	CPMCents          int                    `json:"cpm_cents,omitempty"`
	IsActive          bool                   `json:"is_active"`
	StartDate         *time.Time             `json:"start_date,omitempty"`
	EndDate           *time.Time             `json:"end_date,omitempty"`
	TargetingCriteria map[string]interface{} `json:"targeting_criteria,omitempty"`
}

// UpdateCampaignRequest represents the request to update a campaign
type UpdateCampaignRequest struct {
	Name              *string                `json:"name,omitempty"`
	AdvertiserName    *string                `json:"advertiser_name,omitempty"`
	AdType            *string                `json:"ad_type,omitempty"`
	ContentURL        *string                `json:"content_url,omitempty"`
	ClickURL          *string                `json:"click_url,omitempty"`
	AltText           *string                `json:"alt_text,omitempty"`
	Width             *int                   `json:"width,omitempty"`
	Height            *int                   `json:"height,omitempty"`
	Priority          *int                   `json:"priority,omitempty"`
	Weight            *int                   `json:"weight,omitempty"`
	DailyBudgetCents  *int64                 `json:"daily_budget_cents,omitempty"`
	TotalBudgetCents  *int64                 `json:"total_budget_cents,omitempty"`
	CPMCents          *int                   `json:"cpm_cents,omitempty"`
	IsActive          *bool                  `json:"is_active,omitempty"`
	StartDate         *time.Time             `json:"start_date,omitempty"`
	EndDate           *time.Time             `json:"end_date,omitempty"`
	TargetingCriteria map[string]interface{} `json:"targeting_criteria,omitempty"`
}

// ListCampaigns handles GET /admin/ads/campaigns
// Returns paginated list of campaigns
func (h *AdHandler) ListCampaigns(c *gin.Context) {
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p < 1 {
			c.JSON(http.StatusBadRequest, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "INVALID_PARAMETER",
					Message: "page must be a positive integer",
				},
			})
			return
		}
		page = p
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l < 1 || l > 100 {
			c.JSON(http.StatusBadRequest, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "INVALID_PARAMETER",
					Message: "limit must be between 1 and 100",
				},
			})
			return
		}
		limit = l
	}

	var status *string
	if statusStr := c.Query("status"); statusStr != "" {
		validStatuses := map[string]bool{"active": true, "inactive": true, "ended": true, "scheduled": true}
		if !validStatuses[statusStr] {
			c.JSON(http.StatusBadRequest, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "INVALID_PARAMETER",
					Message: "status must be one of: active, inactive, ended, scheduled",
				},
			})
			return
		}
		status = &statusStr
	}

	campaigns, total, err := h.adService.ListCampaigns(c.Request.Context(), page, limit, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "LIST_CAMPAIGNS_FAILED",
				Message: "Failed to list campaigns",
			},
		})
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"campaigns": campaigns,
		},
		Meta: gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
		},
	})
}

// GetCampaign handles GET /admin/ads/campaigns/:id
// Returns a single campaign by ID
func (h *AdHandler) GetCampaign(c *gin.Context) {
	campaignID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_CAMPAIGN_ID",
				Message: "Invalid campaign ID format",
			},
		})
		return
	}

	campaign, err := h.adService.GetAdByID(c.Request.Context(), campaignID)
	if err != nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "CAMPAIGN_NOT_FOUND",
				Message: "Campaign not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    campaign,
	})
}

// CreateCampaign handles POST /admin/ads/campaigns
// Creates a new campaign
func (h *AdHandler) CreateCampaign(c *gin.Context) {
	var req CreateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// Validate creative if it's a banner
	if req.AdType == "banner" {
		if err := h.adService.ValidateCreative(c.Request.Context(), req.ContentURL, req.AdType, req.Width, req.Height); err != nil {
			c.JSON(http.StatusBadRequest, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "INVALID_CREATIVE",
					Message: err.Error(),
				},
			})
			return
		}
	}

	ad := &models.Ad{
		Name:              req.Name,
		AdvertiserName:    req.AdvertiserName,
		AdType:            req.AdType,
		ContentURL:        req.ContentURL,
		ClickURL:          req.ClickURL,
		AltText:           req.AltText,
		Width:             req.Width,
		Height:            req.Height,
		Priority:          req.Priority,
		Weight:            req.Weight,
		DailyBudgetCents:  req.DailyBudgetCents,
		TotalBudgetCents:  req.TotalBudgetCents,
		CPMCents:          req.CPMCents,
		IsActive:          req.IsActive,
		StartDate:         req.StartDate,
		EndDate:           req.EndDate,
		TargetingCriteria: req.TargetingCriteria,
	}

	if err := h.adService.CreateCampaign(c.Request.Context(), ad); err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "CREATE_CAMPAIGN_FAILED",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, StandardResponse{
		Success: true,
		Data:    ad,
	})
}

// UpdateCampaign handles PUT /admin/ads/campaigns/:id
// Updates an existing campaign
func (h *AdHandler) UpdateCampaign(c *gin.Context) {
	campaignID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_CAMPAIGN_ID",
				Message: "Invalid campaign ID format",
			},
		})
		return
	}

	var req UpdateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// Get existing campaign
	existing, err := h.adService.GetAdByID(c.Request.Context(), campaignID)
	if err != nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "CAMPAIGN_NOT_FOUND",
				Message: "Campaign not found",
			},
		})
		return
	}

	// Apply updates
	ad := existing
	ad.ID = campaignID

	if req.Name != nil {
		ad.Name = *req.Name
	}
	if req.AdvertiserName != nil {
		ad.AdvertiserName = *req.AdvertiserName
	}
	if req.AdType != nil {
		ad.AdType = *req.AdType
	}
	if req.ContentURL != nil {
		ad.ContentURL = *req.ContentURL
	}
	if req.ClickURL != nil {
		ad.ClickURL = req.ClickURL
	}
	if req.AltText != nil {
		ad.AltText = req.AltText
	}
	if req.Width != nil {
		ad.Width = req.Width
	}
	if req.Height != nil {
		ad.Height = req.Height
	}
	if req.Priority != nil {
		ad.Priority = *req.Priority
	}
	if req.Weight != nil {
		ad.Weight = *req.Weight
	}
	if req.DailyBudgetCents != nil {
		ad.DailyBudgetCents = req.DailyBudgetCents
	}
	if req.TotalBudgetCents != nil {
		ad.TotalBudgetCents = req.TotalBudgetCents
	}
	if req.CPMCents != nil {
		ad.CPMCents = *req.CPMCents
	}
	if req.IsActive != nil {
		ad.IsActive = *req.IsActive
	}
	if req.StartDate != nil {
		ad.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		ad.EndDate = req.EndDate
	}
	if req.TargetingCriteria != nil {
		ad.TargetingCriteria = req.TargetingCriteria
	}

	// Validate creative if it's a banner
	if ad.AdType == "banner" {
		if err := h.adService.ValidateCreative(c.Request.Context(), ad.ContentURL, ad.AdType, ad.Width, ad.Height); err != nil {
			c.JSON(http.StatusBadRequest, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "INVALID_CREATIVE",
					Message: err.Error(),
				},
			})
			return
		}
	}

	if err := h.adService.UpdateCampaign(c.Request.Context(), ad); err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UPDATE_CAMPAIGN_FAILED",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    ad,
	})
}

// DeleteCampaign handles DELETE /admin/ads/campaigns/:id
// Deletes a campaign
func (h *AdHandler) DeleteCampaign(c *gin.Context) {
	campaignID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_CAMPAIGN_ID",
				Message: "Invalid campaign ID format",
			},
		})
		return
	}

	if err := h.adService.DeleteCampaign(c.Request.Context(), campaignID); err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "DELETE_CAMPAIGN_FAILED",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"message": "Campaign deleted successfully",
		},
	})
}

// ValidateCreative handles POST /admin/ads/validate-creative
// Validates a creative URL and dimensions
func (h *AdHandler) ValidateCreative(c *gin.Context) {
	var req struct {
		ContentURL string `json:"content_url" binding:"required"`
		AdType     string `json:"ad_type" binding:"required,oneof=banner video native"`
		Width      *int   `json:"width,omitempty"`
		Height     *int   `json:"height,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	if err := h.adService.ValidateCreative(c.Request.Context(), req.ContentURL, req.AdType, req.Width, req.Height); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_CREATIVE",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"valid":   true,
			"message": "Creative is valid",
		},
	})
}

// GetCampaignReportByDate handles GET /admin/ads/reports/by-date
// Returns campaign performance report by date range
func (h *AdHandler) GetCampaignReportByDate(c *gin.Context) {
	var adID *uuid.UUID
	if idStr := c.Query("ad_id"); idStr != "" {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "INVALID_PARAMETER",
					Message: "Invalid ad_id format",
				},
			})
			return
		}
		adID = &id
	}

	// Default to last 30 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startStr := c.Query("start_date"); startStr != "" {
		parsed, err := time.Parse("2006-01-02", startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "INVALID_PARAMETER",
					Message: "Invalid start_date format. Use YYYY-MM-DD",
				},
			})
			return
		}
		startDate = parsed
	}

	if endStr := c.Query("end_date"); endStr != "" {
		parsed, err := time.Parse("2006-01-02", endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "INVALID_PARAMETER",
					Message: "Invalid end_date format. Use YYYY-MM-DD",
				},
			})
			return
		}
		endDate = parsed
	}

	reports, err := h.adService.GetCampaignReportByDate(c.Request.Context(), adID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "REPORT_FAILED",
				Message: "Failed to get campaign report",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"reports":    reports,
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
		},
	})
}

// GetCampaignReportByPlacement handles GET /admin/ads/reports/by-placement
// Returns campaign performance report grouped by placement/slot
func (h *AdHandler) GetCampaignReportByPlacement(c *gin.Context) {
	var adID *uuid.UUID
	if idStr := c.Query("ad_id"); idStr != "" {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "INVALID_PARAMETER",
					Message: "Invalid ad_id format",
				},
			})
			return
		}
		adID = &id
	}

	days, ok := parseDaysParameter(c, 30)
	if !ok {
		return
	}

	since := time.Now().AddDate(0, 0, -days)
	reports, err := h.adService.GetCampaignReportByPlacement(c.Request.Context(), adID, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "REPORT_FAILED",
				Message: "Failed to get placement report",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"reports": reports,
			"days":    days,
		},
	})
}
