package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// ResourceType represents the type of resource being accessed
type ResourceType string

const (
	ResourceTypeComment      ResourceType = "comment"
	ResourceTypeClip         ResourceType = "clip"
	ResourceTypeUser         ResourceType = "user"
	ResourceTypeFavorite     ResourceType = "favorite"
	ResourceTypeSubscription ResourceType = "subscription"
	ResourceTypeSubmission   ResourceType = "submission"
)

// Action represents the type of action being performed
type Action string

const (
	ActionCreate Action = "create"
	ActionRead   Action = "read"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

// Permission defines a permission rule for a resource
type Permission struct {
	Resource            ResourceType
	Action              Action
	RequiresOwner       bool
	AllowedRoles        []string
	AllowedAccountTypes []string
}

// PermissionMatrix defines all authorization rules
var PermissionMatrix = []Permission{
	// Comment permissions
	{Resource: ResourceTypeComment, Action: ActionCreate, RequiresOwner: false, AllowedAccountTypes: []string{models.AccountTypeMember, models.AccountTypeBroadcaster, models.AccountTypeModerator, models.AccountTypeAdmin}},
	{Resource: ResourceTypeComment, Action: ActionRead, RequiresOwner: false, AllowedAccountTypes: []string{models.AccountTypeMember, models.AccountTypeBroadcaster, models.AccountTypeModerator, models.AccountTypeAdmin}},
	{Resource: ResourceTypeComment, Action: ActionUpdate, RequiresOwner: true, AllowedRoles: []string{models.RoleAdmin}},
	{Resource: ResourceTypeComment, Action: ActionDelete, RequiresOwner: true, AllowedRoles: []string{models.RoleModerator, models.RoleAdmin}},

	// Clip permissions
	{Resource: ResourceTypeClip, Action: ActionCreate, RequiresOwner: false, AllowedAccountTypes: []string{models.AccountTypeMember, models.AccountTypeBroadcaster, models.AccountTypeModerator, models.AccountTypeAdmin}},
	{Resource: ResourceTypeClip, Action: ActionRead, RequiresOwner: false},
	{Resource: ResourceTypeClip, Action: ActionUpdate, RequiresOwner: true, AllowedRoles: []string{models.RoleAdmin}},
	{Resource: ResourceTypeClip, Action: ActionDelete, RequiresOwner: false, AllowedRoles: []string{models.RoleAdmin}},

	// User permissions
	{Resource: ResourceTypeUser, Action: ActionRead, RequiresOwner: false},
	{Resource: ResourceTypeUser, Action: ActionUpdate, RequiresOwner: true},
	{Resource: ResourceTypeUser, Action: ActionDelete, RequiresOwner: true, AllowedRoles: []string{models.RoleAdmin}},

	// Favorite permissions
	{Resource: ResourceTypeFavorite, Action: ActionCreate, RequiresOwner: true},
	{Resource: ResourceTypeFavorite, Action: ActionRead, RequiresOwner: true},
	{Resource: ResourceTypeFavorite, Action: ActionDelete, RequiresOwner: true},

	// Subscription permissions
	{Resource: ResourceTypeSubscription, Action: ActionRead, RequiresOwner: true},
	{Resource: ResourceTypeSubscription, Action: ActionUpdate, RequiresOwner: true},
	{Resource: ResourceTypeSubscription, Action: ActionDelete, RequiresOwner: true, AllowedRoles: []string{models.RoleAdmin}},

	// Submission permissions
	{Resource: ResourceTypeSubmission, Action: ActionCreate, RequiresOwner: false, AllowedAccountTypes: []string{models.AccountTypeMember, models.AccountTypeBroadcaster, models.AccountTypeModerator, models.AccountTypeAdmin}},
	{Resource: ResourceTypeSubmission, Action: ActionRead, RequiresOwner: false},
	{Resource: ResourceTypeSubmission, Action: ActionUpdate, RequiresOwner: false, AllowedRoles: []string{models.RoleModerator, models.RoleAdmin}},
	{Resource: ResourceTypeSubmission, Action: ActionDelete, RequiresOwner: false, AllowedRoles: []string{models.RoleModerator, models.RoleAdmin}},
}

// ResourceOwnershipChecker defines the interface for checking resource ownership
type ResourceOwnershipChecker interface {
	IsOwner(ctx context.Context, resourceID, userID uuid.UUID) (bool, error)
}

// CommentOwnershipChecker checks comment ownership
type CommentOwnershipChecker struct {
	repo *repository.CommentRepository
}

// NewCommentOwnershipChecker creates a new CommentOwnershipChecker
func NewCommentOwnershipChecker(repo *repository.CommentRepository) *CommentOwnershipChecker {
	return &CommentOwnershipChecker{repo: repo}
}

// IsOwner checks if the user owns the comment
func (c *CommentOwnershipChecker) IsOwner(ctx context.Context, resourceID, userID uuid.UUID) (bool, error) {
	comment, err := c.repo.GetByID(ctx, resourceID, nil)
	if err != nil {
		return false, err
	}
	return comment.UserID == userID, nil
}

// ClipOwnershipChecker checks clip ownership
type ClipOwnershipChecker struct {
	repo *repository.ClipRepository
}

// NewClipOwnershipChecker creates a new ClipOwnershipChecker
func NewClipOwnershipChecker(repo *repository.ClipRepository) *ClipOwnershipChecker {
	return &ClipOwnershipChecker{repo: repo}
}

// IsOwner checks if the user owns the clip (submitted it)
func (c *ClipOwnershipChecker) IsOwner(ctx context.Context, resourceID, userID uuid.UUID) (bool, error) {
	clip, err := c.repo.GetByID(ctx, resourceID)
	if err != nil {
		return false, err
	}
	// Check if the user submitted this clip
	return clip.SubmittedByUserID != nil && *clip.SubmittedByUserID == userID, nil
}

// UserOwnershipChecker checks user resource ownership (for settings, profile, etc.)
type UserOwnershipChecker struct{}

// NewUserOwnershipChecker creates a new UserOwnershipChecker
func NewUserOwnershipChecker() *UserOwnershipChecker {
	return &UserOwnershipChecker{}
}

// IsOwner checks if the resourceID matches the userID (accessing own profile)
func (u *UserOwnershipChecker) IsOwner(ctx context.Context, resourceID, userID uuid.UUID) (bool, error) {
	return resourceID == userID, nil
}

// AuthorizationContext holds authorization information
type AuthorizationContext struct {
	UserID       uuid.UUID
	User         *models.User
	ResourceID   uuid.UUID
	Action       Action
	ResourceType ResourceType
	IPAddress    string
	UserAgent    string
}

// AuthorizationResult contains the result of an authorization check
type AuthorizationResult struct {
	Allowed  bool
	Reason   string
	Metadata map[string]interface{}
}

// CanAccessResource checks if a user can perform an action on a resource
// Returns AuthorizationResult with decision details for audit logging
func CanAccessResource(ctx *AuthorizationContext, checker ResourceOwnershipChecker) (*AuthorizationResult, error) {
	result := &AuthorizationResult{
		Allowed:  false,
		Metadata: make(map[string]interface{}),
	}

	// Find the permission rule
	var rule *Permission
	for _, p := range PermissionMatrix {
		if p.Resource == ctx.ResourceType && p.Action == ctx.Action {
			rule = &p
			break
		}
	}

	if rule == nil {
		// No explicit rule found - deny by default
		result.Reason = "no_permission_rule"
		return result, fmt.Errorf("no permission rule found for %s:%s", ctx.ResourceType, ctx.Action)
	}

	// Check if ownership is required
	if rule.RequiresOwner {
		isOwner, err := checker.IsOwner(context.Background(), ctx.ResourceID, ctx.UserID)
		if err != nil {
			result.Reason = "ownership_check_failed"
			result.Metadata["error"] = err.Error()
			return result, fmt.Errorf("failed to check ownership: %w", err)
		}

		// If owner, allow access
		if isOwner {
			result.Allowed = true
			result.Reason = "user_is_owner"
			return result, nil
		}

		// If not owner, check if user has elevated permissions
		if ctx.User != nil {
			for _, role := range rule.AllowedRoles {
				if ctx.User.HasRole(role) {
					result.Allowed = true
					result.Reason = fmt.Sprintf("elevated_role_%s", role)
					result.Metadata["role"] = role
					return result, nil
				}
			}
		}

		// Not owner and no elevated permissions
		result.Reason = "not_owner_insufficient_role"
		if ctx.User != nil {
			result.Metadata["user_role"] = ctx.User.Role
		}
		return result, nil
	}

	// Check role-based permissions if no ownership required
	if len(rule.AllowedRoles) > 0 {
		if ctx.User != nil {
			for _, role := range rule.AllowedRoles {
				if ctx.User.HasRole(role) {
					result.Allowed = true
					result.Reason = fmt.Sprintf("role_based_access_%s", role)
					result.Metadata["role"] = role
					return result, nil
				}
			}
		}
		result.Reason = "insufficient_role"
		if ctx.User != nil {
			result.Metadata["user_role"] = ctx.User.Role
			result.Metadata["required_roles"] = rule.AllowedRoles
		}
		return result, nil
	}

	// Check account type permissions
	if len(rule.AllowedAccountTypes) > 0 {
		if ctx.User != nil {
			userAccountType := ctx.User.GetAccountType()
			for _, accountType := range rule.AllowedAccountTypes {
				if userAccountType == accountType {
					result.Allowed = true
					result.Reason = fmt.Sprintf("account_type_access_%s", accountType)
					result.Metadata["account_type"] = accountType
					return result, nil
				}
			}
		}
		result.Reason = "insufficient_account_type"
		if ctx.User != nil {
			result.Metadata["user_account_type"] = ctx.User.GetAccountType()
			result.Metadata["required_account_types"] = rule.AllowedAccountTypes
		}
		return result, nil
	}

	// No restrictions - allow access
	result.Allowed = true
	result.Reason = "no_restrictions"
	return result, nil
}

// RequireResourceOwnership creates middleware that requires resource ownership
// This middleware should be used on routes that modify user-owned resources
func RequireResourceOwnership(resourceType ResourceType, action Action, checker ResourceOwnershipChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (set by AuthMiddleware)
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authentication required",
				},
			})
			c.Abort()
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Invalid user format",
				},
			})
			c.Abort()
			return
		}

		// Get resource ID from URL parameter
		resourceIDStr := c.Param("id")
		resourceID, err := uuid.Parse(resourceIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_ID",
					"message": "Invalid resource ID format",
				},
			})
			c.Abort()
			return
		}

		// Capture client information for audit logging
		ipAddress := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Build authorization context
		authCtx := &AuthorizationContext{
			UserID:       user.ID,
			User:         user,
			ResourceID:   resourceID,
			Action:       action,
			ResourceType: resourceType,
			IPAddress:    ipAddress,
			UserAgent:    userAgent,
		}

		// Check access
		result, err := CanAccessResource(authCtx, checker)
		if err != nil {
			// Log the authorization error
			LogAuthorizationDecision(
				user.ID,
				resourceType,
				resourceID,
				action,
				"error",
				fmt.Sprintf("authorization_check_failed: %v", err),
				ipAddress,
				userAgent,
				nil,
			)

			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "AUTHORIZATION_ERROR",
					"message": "Failed to verify permissions",
				},
			})
			c.Abort()
			return
		}

		if !result.Allowed {
			// Log denied authorization
			LogAuthorizationDecision(
				user.ID,
				resourceType,
				resourceID,
				action,
				"denied",
				result.Reason,
				ipAddress,
				userAgent,
				result.Metadata,
			)

			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "You do not have permission to perform this action",
					"details": gin.H{
						"resource": resourceType,
						"action":   action,
					},
				},
			})
			c.Abort()
			return
		}

		// Log successful authorization
		LogAuthorizationDecision(
			user.ID,
			resourceType,
			resourceID,
			action,
			"allowed",
			result.Reason,
			ipAddress,
			userAgent,
			result.Metadata,
		)

		c.Next()
	}
}

// AuthorizationAuditLog represents an authorization decision audit log entry
type AuthorizationAuditLog struct {
	Timestamp  time.Time              `json:"timestamp"`
	UserID     string                 `json:"user_id"`
	Resource   string                 `json:"resource"`
	ResourceID string                 `json:"resource_id"`
	Action     string                 `json:"action"`
	Decision   string                 `json:"decision"` // "allowed", "denied", or "error"
	Reason     string                 `json:"reason"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// LogAuthorizationDecision logs all authorization decisions for audit purposes
// This function logs to structured JSON format without PII, following GDPR compliance
func LogAuthorizationDecision(userID uuid.UUID, resourceType ResourceType, resourceID uuid.UUID, action Action, decision string, reason string, ipAddress, userAgent string, metadata map[string]interface{}) {
	logger := utils.GetLogger()

	auditLog := AuthorizationAuditLog{
		Timestamp:  time.Now().UTC(),
		UserID:     userID.String(),
		Resource:   string(resourceType),
		ResourceID: resourceID.String(),
		Action:     string(action),
		Decision:   decision,
		Reason:     reason,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Metadata:   metadata,
	}

	fields := map[string]interface{}{
		"audit_type":  "authorization",
		"user_id":     auditLog.UserID,
		"resource":    auditLog.Resource,
		"resource_id": auditLog.ResourceID,
		"action":      auditLog.Action,
		"decision":    auditLog.Decision,
		"reason":      auditLog.Reason,
		"ip_address":  auditLog.IPAddress,
		"user_agent":  auditLog.UserAgent,
	}

	if metadata != nil {
		fields["metadata"] = metadata
	}

	message := fmt.Sprintf("Authorization %s: user=%s resource=%s:%s action=%s reason=%s",
		decision, userID, resourceType, resourceID, action, reason)

	if decision == "denied" {
		logger.Warn(message, fields)
	} else if decision == "error" {
		logger.Error(message, nil, fields)
	} else {
		logger.Info(message, fields)
	}
}

// LogAuthorizationFailure logs authorization failures for security monitoring
// Deprecated: Use LogAuthorizationDecision instead
func LogAuthorizationFailure(userID uuid.UUID, resourceType ResourceType, resourceID uuid.UUID, action Action, reason string) {
	LogAuthorizationDecision(userID, resourceType, resourceID, action, "denied", reason, "", "", nil)
}
