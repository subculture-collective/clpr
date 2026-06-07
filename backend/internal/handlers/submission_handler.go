package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const uploadMultipartOverheadBytes int64 = 1 << 20

// SubmissionHandler handles clip submission operations
type SubmissionHandler struct {
	submissionService submissionServiceAPI
	uploadValidator   *services.UploadValidator
	clipStorage       storage.ClipStorage
}

type submissionServiceAPI interface {
	SubmitClip(ctx context.Context, userID uuid.UUID, req *services.SubmitClipRequest, ip string, deviceFingerprint string) (*models.ClipSubmission, error)
	SubmitUpload(ctx context.Context, userID uuid.UUID, req *services.SubmitUploadRequest, ip string, deviceFingerprint string) (*models.ClipSubmission, error)
	GetUserSubmissions(ctx context.Context, userID uuid.UUID, page, limit int) ([]*models.ClipSubmission, int, error)
	GetSubmissionStats(ctx context.Context, userID uuid.UUID) (*models.SubmissionStats, error)
	GetClipMetadata(ctx context.Context, clipURLOrID string) (*services.ClipMetadata, error)
	CheckClipExistence(ctx context.Context, twitchClipID string) (*services.ClipExistenceResult, error)
	GetPendingSubmissionsWithFilters(ctx context.Context, filters repository.SubmissionFilters, page, limit int) ([]*models.ClipSubmissionWithUser, int, error)
	ApproveSubmission(ctx context.Context, submissionID, reviewerID uuid.UUID) error
	RejectSubmission(ctx context.Context, submissionID, reviewerID uuid.UUID, reason string) error
	BulkApproveSubmissions(ctx context.Context, submissionIDs []uuid.UUID, reviewerID uuid.UUID) error
	BulkRejectSubmissions(ctx context.Context, submissionIDs []uuid.UUID, reviewerID uuid.UUID, reason string) error
}

type submissionResponse struct {
	ID                      string     `json:"id"`
	UserID                  string     `json:"user_id"`
	ClipID                  *string    `json:"clip_id,omitempty"`
	TwitchClipID            string     `json:"twitch_clip_id"`
	TwitchClipURL           string     `json:"twitch_clip_url"`
	Title                   *string    `json:"title,omitempty"`
	CustomTitle             *string    `json:"custom_title,omitempty"`
	BroadcasterNameOverride *string    `json:"broadcaster_name_override,omitempty"`
	Tags                    []string   `json:"tags,omitempty"`
	IsNSFW                  bool       `json:"is_nsfw"`
	SubmissionReason        *string    `json:"submission_reason,omitempty"`
	Status                  string     `json:"status"`
	RejectionReason         *string    `json:"rejection_reason,omitempty"`
	ReviewedBy              *string    `json:"reviewed_by,omitempty"`
	ReviewedAt              *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
	SourceType              string     `json:"source_type"`
	SourcePlatform          string     `json:"source_platform"`
	SourceURL               *string    `json:"source_url,omitempty"`
	SourceID                *string    `json:"source_id,omitempty"`
	DurationSeconds         *int       `json:"duration_seconds,omitempty"`
	DurationVerified        bool       `json:"duration_verified"`
	CreatorName             *string    `json:"creator_name,omitempty"`
	CreatorID               *string    `json:"creator_id,omitempty"`
	BroadcasterName         *string    `json:"broadcaster_name,omitempty"`
	BroadcasterID           *string    `json:"broadcaster_id,omitempty"`
	GameID                  *string    `json:"game_id,omitempty"`
	GameName                *string    `json:"game_name,omitempty"`
	ThumbnailURL            *string    `json:"thumbnail_url,omitempty"`
	Duration                *float64   `json:"duration,omitempty"`
	ViewCount               int        `json:"view_count"`
}

func toSubmissionResponse(submission *models.ClipSubmission) submissionResponse {
	var clipID *string
	if submission.ClipID != nil {
		value := submission.ClipID.String()
		clipID = &value
	}
	var reviewedBy *string
	if submission.ReviewedBy != nil {
		value := submission.ReviewedBy.String()
		reviewedBy = &value
	}
	return submissionResponse{
		ID:                      submission.ID.String(),
		UserID:                  submission.UserID.String(),
		ClipID:                  clipID,
		TwitchClipID:            submission.TwitchClipID,
		TwitchClipURL:           submission.TwitchClipURL,
		Title:                   submission.Title,
		CustomTitle:             submission.CustomTitle,
		BroadcasterNameOverride: submission.BroadcasterNameOverride,
		Tags:                    submission.Tags,
		IsNSFW:                  submission.IsNSFW,
		SubmissionReason:        submission.SubmissionReason,
		Status:                  submission.Status,
		RejectionReason:         submission.RejectionReason,
		ReviewedBy:              reviewedBy,
		ReviewedAt:              submission.ReviewedAt,
		CreatedAt:               submission.CreatedAt,
		UpdatedAt:               submission.UpdatedAt,
		SourceType:              submission.SourceType,
		SourcePlatform:          submission.SourcePlatform,
		SourceURL:               submission.SourceURL,
		SourceID:                submission.SourceID,
		DurationSeconds:         submission.DurationSeconds,
		DurationVerified:        submission.DurationVerified,
		CreatorName:             submission.CreatorName,
		CreatorID:               submission.CreatorID,
		BroadcasterName:         submission.BroadcasterName,
		BroadcasterID:           submission.BroadcasterID,
		GameID:                  submission.GameID,
		GameName:                submission.GameName,
		ThumbnailURL:            submission.ThumbnailURL,
		Duration:                submission.Duration,
		ViewCount:               submission.ViewCount,
	}
}

// NewSubmissionHandler creates a new SubmissionHandler
func NewSubmissionHandler(submissionService submissionServiceAPI, uploadValidator *services.UploadValidator, clipStorage storage.ClipStorage) *SubmissionHandler {
	return &SubmissionHandler{
		submissionService: submissionService,
		uploadValidator:   uploadValidator,
		clipStorage:       clipStorage,
	}
}

// SubmitClip handles clip submission
// POST /clips/submit
func (h *SubmissionHandler) SubmitClip(c *gin.Context) {
	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	userID := userIDVal.(uuid.UUID)

	var req services.SubmitClipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Get IP address and device fingerprint for abuse detection
	ip := c.ClientIP()
	deviceFingerprint := c.GetHeader("User-Agent") // Simple fingerprint using user agent

	submission, err := h.submissionService.SubmitClip(c.Request.Context(), userID, &req, ip, deviceFingerprint)
	if err != nil {
		var moderationErr *services.CreatorModerationError
		if errors.As(err, &moderationErr) {
			c.JSON(http.StatusForbidden, gin.H{"error": moderationErr.Message})
			return
		}
		// Check if it's a rate limit error
		if rateLimitErr, ok := err.(*services.RateLimitError); ok {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       rateLimitErr.Error,
				"limit":       rateLimitErr.Limit,
				"window":      rateLimitErr.Window,
				"retry_after": rateLimitErr.RetryAfter,
			})
			return
		}

		// Check if it's a validation error
		if valErr, ok := err.(*services.ValidationError); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   valErr.Message,
				"field":   valErr.Field,
				"success": false,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to submit clip",
			"success": false,
		})
		return
	}

	status := http.StatusCreated
	message := "Clip submitted for review"
	if submission.Status == "approved" {
		message = "Clip submitted and auto-approved!"
	}

	c.JSON(status, gin.H{
		"success":    true,
		"message":    message,
		"submission": toSubmissionResponse(submission),
	})
}

// SubmitUpload handles hosted clip uploads.
// POST /submissions/upload
func (h *SubmissionHandler) SubmitUpload(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	if h.submissionService == nil || h.uploadValidator == nil || h.clipStorage == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Hosted uploads are not configured"})
		return
	}

	maxUploadBytes := h.uploadValidator.MaxUploadBytes()
	if maxUploadBytes > 0 {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBytes+uploadMultipartOverheadBytes)
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		if isUploadTooLargeError(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": fmt.Sprintf("uploaded file exceeds maximum allowed size of %d bytes", maxUploadBytes)})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read uploaded file"})
		return
	}

	if maxUploadBytes > 0 && fileHeader.Size > maxUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": fmt.Sprintf("uploaded file exceeds maximum allowed size of %d bytes", maxUploadBytes)})
		return
	}

	customTitle := parseOptionalUploadField(c.PostForm("custom_title"))
	submissionReason := parseOptionalUploadField(c.PostForm("submission_reason"))
	isNSFW := false
	if rawNSFW := strings.TrimSpace(c.PostForm("is_nsfw")); rawNSFW != "" {
		parsed, parseErr := strconv.ParseBool(rawNSFW)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "is_nsfw must be a boolean"})
			return
		}
		isNSFW = parsed
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open uploaded file"})
		return
	}
	defer file.Close()

	mimeType, err := detectUploadMimeType(file, fileHeader.Filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to inspect uploaded file"})
		return
	}

	tempFile, err := os.CreateTemp("", "clip-upload-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare upload"})
		return
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if copied, err := io.Copy(tempFile, io.LimitReader(file, maxUploadBytes+1)); err != nil {
		tempFile.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to stage uploaded file"})
		return
	} else if maxUploadBytes > 0 && copied > maxUploadBytes {
		tempFile.Close()
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": fmt.Sprintf("uploaded file exceeds maximum allowed size of %d bytes", maxUploadBytes)})
		return
	}
	if err := tempFile.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to stage uploaded file"})
		return
	}

	validationResult, err := h.uploadValidator.Validate(c.Request.Context(), mimeType, fileHeader.Size, tempPath)
	if err != nil {
		if valErr, ok := err.(*services.ValidationError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": valErr.Message, "field": valErr.Field, "success": false})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to validate upload"})
		return
	}

	uploadFile, err := os.Open(tempPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reopen uploaded file"})
		return
	}
	defer uploadFile.Close()

	submissionID := uuid.New()
	keyExt, err := uploadExtensionForMimeType(validationResult.MimeType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	storageKey := services.BuildUploadStorageKey(userID, submissionID, keyExt)

	storageResult, err := h.clipStorage.PutObject(c.Request.Context(), storage.PutObjectInput{
		Key:         storageKey,
		Body:        uploadFile,
		Size:        validationResult.FileSizeBytes,
		ContentType: validationResult.MimeType,
		Metadata: map[string]string{
			"user_id":           userID.String(),
			"submission_id":     submissionID.String(),
			"original_filename": fileHeader.Filename,
			"mime_type":         validationResult.MimeType,
			"duration_seconds":  strconv.FormatInt(validationResult.DurationSeconds, 10),
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store uploaded clip"})
		return
	}

	submission, err := h.submissionService.SubmitUpload(c.Request.Context(), userID, &services.SubmitUploadRequest{
		SubmissionID:     submissionID,
		CustomTitle:      customTitle,
		IsNSFW:           isNSFW,
		SubmissionReason: submissionReason,
		OriginalFilename: fileHeader.Filename,
		MimeType:         validationResult.MimeType,
		FileSizeBytes:    validationResult.FileSizeBytes,
		DurationSeconds:  validationResult.DurationSeconds,
		DurationVerified: validationResult.DurationVerified,
		StorageProvider:  storageResult.Provider,
		StorageBucket:    storageResult.Bucket,
		StorageKey:       storageResult.Key,
	}, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		var moderationErr *services.CreatorModerationError
		if errors.As(err, &moderationErr) {
			c.JSON(http.StatusForbidden, gin.H{"error": moderationErr.Message})
			return
		}
		_ = h.clipStorage.DeleteObject(c.Request.Context(), storageResult.Key)
		if rateLimitErr, ok := err.(*services.RateLimitError); ok {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": rateLimitErr.Error, "limit": rateLimitErr.Limit, "window": rateLimitErr.Window, "retry_after": rateLimitErr.RetryAfter})
			return
		}
		if valErr, ok := err.(*services.ValidationError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": valErr.Message, "field": valErr.Field, "success": false})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit upload"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":    true,
		"message":    "Upload submitted for review",
		"submission": toSubmissionResponse(submission),
	})
}

// GetUserSubmissions lists submissions for the authenticated user
// GET /submissions
func (h *SubmissionHandler) GetUserSubmissions(c *gin.Context) {
	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	userID := userIDVal.(uuid.UUID)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	submissions, total, err := h.submissionService.GetUserSubmissions(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve submissions",
		})
		return
	}

	totalPages := (total + limit - 1) / limit
	responseSubmissions := make([]submissionResponse, 0, len(submissions))
	for _, submission := range submissions {
		if submission == nil {
			continue
		}
		responseSubmissions = append(responseSubmissions, toSubmissionResponse(submission))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responseSubmissions,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func parseOptionalUploadField(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func isUploadTooLargeError(err error) bool {
	if err == nil {
		return false
	}

	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return true
	}

	if errors.Is(err, multipart.ErrMessageTooLarge) {
		return true
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "request body too large") || strings.Contains(msg, "multipart: message too large")
}

func detectUploadMimeType(file multipart.File, filename string) (string, error) {
	if file == nil {
		return "", fmt.Errorf("file is required")
	}

	if seeker, ok := file.(io.Seeker); ok {
		buf := make([]byte, 512)
		n, readErr := file.Read(buf)
		if readErr != nil && readErr != io.EOF {
			return "", readErr
		}
		mimeType := http.DetectContentType(buf[:n])
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return "", err
		}
		if mimeType != "" && mimeType != "application/octet-stream" {
			return mimeType, nil
		}
	}

	if mimeType := strings.TrimSpace(mime.TypeByExtension(strings.ToLower(filepath.Ext(filename)))); mimeType != "" {
		return mimeType, nil
	}

	return "application/octet-stream", nil
}

func uploadExtensionForMimeType(mimeType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "video/mp4":
		return ".mp4", nil
	case "video/webm":
		return ".webm", nil
	case "video/quicktime":
		return ".mov", nil
	default:
		return "", fmt.Errorf("unsupported upload MIME type: %s", mimeType)
	}
}

// GetSubmissionStats returns submission statistics for the authenticated user
// GET /submissions/stats
func (h *SubmissionHandler) GetSubmissionStats(c *gin.Context) {
	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	userID := userIDVal.(uuid.UUID)

	stats, err := h.submissionService.GetSubmissionStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve stats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetClipMetadata fetches clip metadata from Twitch API
// GET /submissions/metadata?url={twitchClipUrl}
func (h *SubmissionHandler) GetClipMetadata(c *gin.Context) {
	clipURL := c.Query("url")
	if clipURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "URL parameter is required",
			"field":   "url",
		})
		return
	}

	metadata, err := h.submissionService.GetClipMetadata(c.Request.Context(), clipURL)
	if err != nil {
		// Check if it's a validation error
		if valErr, ok := err.(*services.ValidationError); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   valErr.Message,
				"field":   valErr.Field,
			})
			return
		}

		// Check for Twitch API errors (502 Bad Gateway)
		if _, ok := err.(*services.TwitchAPIError); ok {
			c.JSON(http.StatusBadGateway, gin.H{
				"success": false,
				"error":   "Unable to fetch clip metadata from Twitch. Please try again later.",
			})
			return
		}

		// Check for Twitch API not configured error
		if strings.Contains(err.Error(), "not configured") {
			c.JSON(http.StatusBadGateway, gin.H{
				"success": false,
				"error":   "Unable to fetch clip metadata from Twitch. Please try again later.",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch clip metadata",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metadata,
	})
}

// CheckClipStatus checks if a clip exists and whether it can be claimed
// GET /submissions/check/:clip_id
// Note: This endpoint is public to allow users to check clip status before attempting to claim.
// Sensitive fields are filtered from the response.
func (h *SubmissionHandler) CheckClipStatus(c *gin.Context) {
	clipID := c.Param("clip_id")
	if clipID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Clip ID is required",
		})
		return
	}

	result, err := h.submissionService.CheckClipExistence(c.Request.Context(), clipID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to check clip status",
		})
		return
	}

	response := gin.H{
		"success":        true,
		"exists":         result.Exists,
		"can_be_claimed": result.CanBeClaimed,
	}

	// If clip exists, return minimal public information only
	if result.Exists && result.Clip != nil {
		response["clip"] = gin.H{
			"id":               result.Clip.ID,
			"title":            result.Clip.Title,
			"broadcaster_name": result.Clip.BroadcasterName,
			"game_name":        result.Clip.GameName,
			"view_count":       result.Clip.ViewCount,
			"created_at":       result.Clip.CreatedAt,
			// Exclude sensitive fields: is_removed, removed_reason, submitted_by_user_id, etc.
		}
	}

	c.JSON(http.StatusOK, response)
}

// ListPendingSubmissions lists pending submissions for moderation (admin/moderator only)
// GET /admin/submissions
// Supports filters: is_nsfw, broadcaster, creator, tags (comma-separated), start_date (RFC3339), end_date (RFC3339)
func (h *SubmissionHandler) ListPendingSubmissions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Parse filters
	filters := repository.SubmissionFilters{}

	if isNSFWStr := c.Query("is_nsfw"); isNSFWStr != "" {
		isNSFW := isNSFWStr == "true"
		filters.IsNSFW = &isNSFW
	}

	if broadcaster := c.Query("broadcaster"); broadcaster != "" {
		filters.BroadcasterName = &broadcaster
	}

	if creator := c.Query("creator"); creator != "" {
		filters.CreatorName = &creator
	}

	if tagsStr := c.Query("tags"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
		filters.Tags = tags
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid start_date format (use RFC3339)",
			})
			return
		}
		filters.StartDate = &startDate
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid end_date format (use RFC3339)",
			})
			return
		}
		filters.EndDate = &endDate
	}

	submissions, total, err := h.submissionService.GetPendingSubmissionsWithFilters(c.Request.Context(), filters, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve pending submissions",
		})
		return
	}

	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    submissions,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// ApproveSubmission approves a pending submission (admin/moderator only)
// POST /admin/submissions/:id/approve
func (h *SubmissionHandler) ApproveSubmission(c *gin.Context) {
	// Get submission ID from URL
	submissionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid submission ID",
		})
		return
	}

	// Get reviewer ID from context
	reviewerIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	reviewerID := reviewerIDVal.(uuid.UUID)

	if err := h.submissionService.ApproveSubmission(c.Request.Context(), submissionID, reviewerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to approve submission: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Submission approved",
	})
}

// RejectSubmission rejects a pending submission (admin/moderator only)
// POST /admin/submissions/:id/reject
func (h *SubmissionHandler) RejectSubmission(c *gin.Context) {
	// Get submission ID from URL
	submissionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid submission ID",
		})
		return
	}

	// Get reviewer ID from context
	reviewerIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	reviewerID := reviewerIDVal.(uuid.UUID)

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Rejection reason is required",
		})
		return
	}

	if err := h.submissionService.RejectSubmission(c.Request.Context(), submissionID, reviewerID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to reject submission: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Submission rejected",
	})
}

// BulkApproveSubmissions approves multiple submissions (admin/moderator only)
// POST /admin/submissions/bulk-approve
func (h *SubmissionHandler) BulkApproveSubmissions(c *gin.Context) {
	// Get reviewer ID from context
	reviewerIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	reviewerID := reviewerIDVal.(uuid.UUID)

	var req struct {
		SubmissionIDs []string `json:"submission_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Submission IDs are required",
		})
		return
	}

	if len(req.SubmissionIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "At least one submission ID is required",
		})
		return
	}

	// Parse UUIDs
	submissionIDs := make([]uuid.UUID, 0, len(req.SubmissionIDs))
	for _, idStr := range req.SubmissionIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid submission ID: " + idStr,
			})
			return
		}
		submissionIDs = append(submissionIDs, id)
	}

	if err := h.submissionService.BulkApproveSubmissions(c.Request.Context(), submissionIDs, reviewerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to bulk approve submissions: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Submissions approved",
		"count":   len(submissionIDs),
	})
}

// BulkRejectSubmissions rejects multiple submissions (admin/moderator only)
// POST /admin/submissions/bulk-reject
func (h *SubmissionHandler) BulkRejectSubmissions(c *gin.Context) {
	// Get reviewer ID from context
	reviewerIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	reviewerID := reviewerIDVal.(uuid.UUID)

	var req struct {
		SubmissionIDs []string `json:"submission_ids" binding:"required"`
		Reason        string   `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Submission IDs and reason are required",
		})
		return
	}

	if len(req.SubmissionIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "At least one submission ID is required",
		})
		return
	}

	// Parse UUIDs
	submissionIDs := make([]uuid.UUID, 0, len(req.SubmissionIDs))
	for _, idStr := range req.SubmissionIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid submission ID: " + idStr,
			})
			return
		}
		submissionIDs = append(submissionIDs, id)
	}

	if err := h.submissionService.BulkRejectSubmissions(c.Request.Context(), submissionIDs, reviewerID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to bulk reject submissions: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Submissions rejected",
		"count":   len(submissionIDs),
	})
}

// GetRejectionReasonTemplates returns available rejection reason templates
// GET /admin/submissions/rejection-reasons
func (h *SubmissionHandler) GetRejectionReasonTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    models.GetRejectionReasonTemplates(),
	})
}
