---
title: "Moderation System Developer Guide"
summary: "Comprehensive guide to implementing moderation features in Clipper"
tags: ["backend", "moderation", "development", "guide", "onboarding"]
area: "backend"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2025-02-03
aliases: ["dev-guide", "moderation-dev", "developer-guide"]
---

# Moderation System Developer Guide

> **Complete guide for implementing moderation features in the Clipper platform**

This comprehensive guide covers everything developers need to know to implement new moderation features, add permissions, create audit trails, and integrate with the frontend. After reading this guide, you should be able to implement complete moderation workflows end-to-end.

## Table of Contents

- [System Architecture](#system-architecture)
- [Adding New Permissions](#adding-new-permissions)
- [Adding New Moderation Actions](#adding-new-moderation-actions)
- [Service Layer Patterns](#service-layer-patterns)
- [Database Schema & Migrations](#database-schema--migrations)
- [API Endpoint Patterns](#api-endpoint-patterns)
- [Frontend Integration](#frontend-integration)
- [Testing Strategies](#testing-strategies)
- [Logging & Audit Trails](#logging--audit-trails)
- [Debugging Guide](#debugging-guide)
- [Quick Reference](#quick-reference)

---

## System Architecture

The moderation system follows a clean layered architecture that separates concerns and enables testability.

### Architecture Layers

```
┌─────────────────────────────────────────────────┐
│           Frontend (React/TypeScript)            │
│  - Components with permission-based rendering    │
│  - API client for moderation endpoints          │
│  - React hooks for state management             │
└───────────────────┬─────────────────────────────┘
                    │ HTTP/REST API
┌───────────────────▼─────────────────────────────┐
│             API Handler Layer (Gin)              │
│  - Request validation & parsing                  │
│  - Authentication & authorization middleware     │
│  - Error handling & response formatting          │
└───────────────────┬─────────────────────────────┘
                    │
┌───────────────────▼─────────────────────────────┐
│              Service Layer (Go)                  │
│  - Business logic & orchestration                │
│  - Permission & scope validation                 │
│  - Transaction management                        │
│  - Audit log creation                            │
└───────────────────┬─────────────────────────────┘
                    │
┌───────────────────▼─────────────────────────────┐
│           Repository Layer (Go)                  │
│  - Database queries (CRUD operations)            │
│  - Data access abstraction                       │
│  - Query optimization                            │
└───────────────────┬─────────────────────────────┘
                    │
┌───────────────────▼─────────────────────────────┐
│           Data Layer (PostgreSQL)                │
│  - Tables: users, community_bans, audit_logs    │
│  - Indexes, constraints, triggers                │
│  - Transaction isolation                         │
└─────────────────────────────────────────────────┘
```

### Key Components

#### 1. Models (`internal/models/`)
Define data structures and permission logic:
- `roles.go` - Permission constants, account types, permission checks
- `user.go` - User model with permission methods
- `community.go` - Community and ban models

#### 2. Services (`internal/services/`)
Contain business logic:
- `moderation_service.go` - Core ban/unban operations
- `moderation_event_service.go` - Event queue processing
- `audit_log_service.go` - Audit trail management

#### 3. Handlers (`internal/handlers/`)
Handle HTTP requests:
- `moderation_handler.go` - Moderation API endpoints
- Middleware for authentication and authorization

#### 4. Repositories (`internal/repository/`)
Manage data access:
- `community_repository.go` - Community and ban queries
- `user_repository.go` - User queries
- `audit_log_repository.go` - Audit log queries

### Data Flow Example: Ban User

```
1. POST /api/v1/moderation/communities/:id/ban
   ↓
2. Handler validates JWT token, extracts user_id
   ↓
3. Handler calls ModerationService.BanUser()
   ↓
4. Service validates moderator permissions & scope
   ↓
5. Service checks if target is community owner (cannot ban)
   ↓
6. Repository removes member from community
   ↓
7. Repository creates ban record in community_bans
   ↓
8. Service creates audit log entry
   ↓
9. Handler returns success response
```

---

## Adding New Permissions

Permissions control what actions users can perform. Follow these steps to add a new permission to the system.

### Step 1: Define Permission Constant

Add the permission constant to `backend/internal/models/roles.go`:

```go
// File: backend/internal/models/roles.go

const (
    // Existing permissions...
    
    // Your new permission
    PermissionManageReports = "manage:reports"  // New!
)
```

**Naming Convention:**
- Format: `{action}:{resource}`
- Actions: `create`, `view`, `manage`, `moderate`, `delete`
- Resources: `submission`, `comment`, `users`, `reports`, `analytics`

### Step 2: Add to Account Type Permissions

Update the `accountTypePermissions` map to grant the permission to appropriate account types:

```go
// File: backend/internal/models/roles.go

var accountTypePermissions = map[string][]string{
    AccountTypeModerator: {
        // All broadcaster permissions
        PermissionCreateSubmission,
        PermissionCreateComment,
        // ... other permissions
        
        // Add your new permission
        PermissionManageReports,  // New!
    },
    AccountTypeAdmin: {
        // All permissions including new one
        PermissionManageReports,  // New!
        // ... other permissions
    },
}
```

**Permission Hierarchy:**
- `member` - Base permissions only
- `broadcaster` - Inherits member + broadcaster permissions
- `moderator` - Inherits broadcaster + moderation permissions
- `community_moderator` - Limited, channel-scoped permissions
- `admin` - All permissions

### Step 3: Create Permission Check Method (Optional)

For commonly used permissions, add a convenience method to the User model:

```go
// File: backend/internal/models/roles.go (or user.go)

// CanManageReports checks if user can manage reports
func (u *User) CanManageReports() bool {
    return u.Can(PermissionManageReports)
}
```

### Step 4: Add Middleware Protection

Create or use middleware to protect routes requiring the permission:

```go
// File: backend/internal/handlers/middleware.go (or your route file)

// RequireManageReports middleware
func RequireManageReports() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := c.Get("user_id")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        
        // Get user from context or database
        user, err := getUserByID(c.Request.Context(), userID.(uuid.UUID))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load user"})
            c.Abort()
            return
        }
        
        if !user.Can(PermissionManageReports) {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "You do not have permission to manage reports",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### Step 5: Apply Middleware to Routes

Apply the middleware to relevant routes:

```go
// File: backend/cmd/server/main.go or route configuration

api := router.Group("/api/v1")
{
    // Protected routes
    reports := api.Group("/reports")
    reports.Use(AuthRequired())  // Verify JWT token
    reports.Use(RequireManageReports())  // Check permission
    {
        reports.GET("", reportHandler.ListReports)
        reports.POST("/:id/resolve", reportHandler.ResolveReport)
        reports.DELETE("/:id", reportHandler.DeleteReport)
    }
}
```

### Step 6: Test Permission Checks

Write tests to verify permission behavior:

```go
// File: backend/internal/models/roles_test.go

func TestCanManageReports(t *testing.T) {
    tests := []struct {
        name        string
        accountType string
        expected    bool
    }{
        {"member cannot manage reports", AccountTypeMember, false},
        {"broadcaster cannot manage reports", AccountTypeBroadcaster, false},
        {"moderator can manage reports", AccountTypeModerator, true},
        {"admin can manage reports", AccountTypeAdmin, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            user := &User{AccountType: tt.accountType}
            result := user.Can(PermissionManageReports)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Complete Example: Adding "Manage Ban Templates" Permission

```go
// 1. Define constant
const PermissionManageBanTemplates = "manage:ban_templates"

// 2. Add to accountTypePermissions
AccountTypeModerator: {
    // ... existing permissions
    PermissionManageBanTemplates,
},

// 3. Create check method
func (u *User) CanManageBanTemplates() bool {
    return u.Can(PermissionManageBanTemplates)
}

// 4. Create middleware
func RequireManageBanTemplates() gin.HandlerFunc {
    return func(c *gin.Context) {
        user := mustGetUser(c)
        if !user.CanManageBanTemplates() {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "Permission denied: manage:ban_templates required",
            })
            c.Abort()
            return
        }
        c.Next()
    }
}

// 5. Apply to routes
banTemplates := api.Group("/ban-templates")
banTemplates.Use(AuthRequired(), RequireManageBanTemplates())
{
    banTemplates.GET("", handler.ListTemplates)
    banTemplates.POST("", handler.CreateTemplate)
    banTemplates.PUT("/:id", handler.UpdateTemplate)
    banTemplates.DELETE("/:id", handler.DeleteTemplate)
}
```


---

## Adding New Moderation Actions

This section shows how to add a complete moderation action from service to handler to routes.

### Step 1: Define Service Method

Add business logic to the moderation service:

```go
// File: backend/internal/services/moderation_service.go

// WarnUser sends a warning to a user without banning them
func (s *ModerationService) WarnUser(
    ctx context.Context,
    communityID, moderatorID, targetUserID uuid.UUID,
    warningMessage string,
) error {
    // 1. Get moderator user
    moderator, err := s.userRepo.GetByID(ctx, moderatorID)
    if err != nil {
        return fmt.Errorf("failed to get moderator: %w", err)
    }

    // 2. Validate scope
    if err := s.validateModerationScope(moderator, communityID); err != nil {
        return err
    }

    // 3. Validate permission
    if err := s.validateModerationPermission(ctx, moderator, communityID); err != nil {
        return err
    }

    // 4. Check if target user exists
    targetUser, err := s.userRepo.GetByID(ctx, targetUserID)
    if err != nil {
        return ErrModerationUserNotFound
    }

    // 5. Create warning record (add to warnings table)
    warning := &models.UserWarning{
        ID:          uuid.New(),
        CommunityID: communityID,
        UserID:      targetUserID,
        ModeratorID: moderatorID,
        Message:     warningMessage,
        CreatedAt:   time.Now(),
    }

    if err := s.warningRepo.Create(ctx, warning); err != nil {
        return fmt.Errorf("failed to create warning: %w", err)
    }

    // 6. Send notification to user (optional)
    // s.notificationService.SendWarningNotification(ctx, targetUser, warningMessage)

    // 7. Create audit log
    metadata := map[string]interface{}{
        "community_id":    communityID.String(),
        "target_user_id":  targetUserID.String(),
        "warning_message": warningMessage,
        "moderator_scope": moderator.ModeratorScope,
    }

    auditLog := &models.ModerationAuditLog{
        Action:      "warn_user",
        EntityType:  "user_warning",
        EntityID:    warning.ID,
        ModeratorID: moderatorID,
        Metadata:    metadata,
    }

    if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
        // Log error but don't fail the operation
        return fmt.Errorf("failed to create audit log: %w", err)
    }

    return nil
}
```

**Service Layer Best Practices:**
1. Always validate permissions and scope first
2. Use sentinel errors for common failure cases
3. Create audit logs for all moderation actions
4. Use transactions for multi-step operations
5. Return descriptive errors with context

### Step 2: Add Handler Endpoint

Create HTTP handler for the action:

```go
// File: backend/internal/handlers/moderation_handler.go

// WarnUserRequest represents the request body for warning a user
type WarnUserRequest struct {
    TargetUserID string `json:"target_user_id" binding:"required,uuid"`
    Message      string `json:"message" binding:"required,min=10,max=500"`
}

// WarnUser issues a warning to a community member
// POST /api/v1/moderation/communities/:community_id/warn
func (h *ModerationHandler) WarnUser(c *gin.Context) {
    ctx := c.Request.Context()

    // 1. Parse community ID from URL
    communityID, err := uuid.Parse(c.Param("community_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid community ID",
        })
        return
    }

    // 2. Get moderator ID from authenticated context
    moderatorIDVal, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error": "Unauthorized",
        })
        return
    }
    moderatorID := moderatorIDVal.(uuid.UUID)

    // 3. Parse and validate request body
    var req WarnUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid request",
            "details": err.Error(),
        })
        return
    }

    // 4. Parse target user ID
    targetUserID, err := uuid.Parse(req.TargetUserID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid target user ID",
        })
        return
    }

    // 5. Call service method
    err = h.moderationService.WarnUser(
        ctx,
        communityID,
        moderatorID,
        targetUserID,
        req.Message,
    )

    // 6. Handle errors
    if err != nil {
        switch {
        case errors.Is(err, services.ErrModerationPermissionDenied):
            c.JSON(http.StatusForbidden, gin.H{
                "error": "You do not have permission to warn users",
            })
        case errors.Is(err, services.ErrModerationNotAuthorized):
            c.JSON(http.StatusForbidden, gin.H{
                "error": "You are not authorized to moderate this community",
            })
        case errors.Is(err, services.ErrModerationUserNotFound):
            c.JSON(http.StatusNotFound, gin.H{
                "error": "User not found",
            })
        default:
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Failed to warn user",
            })
        }
        return
    }

    // 7. Return success response
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "User warning issued successfully",
    })
}
```

**Handler Best Practices:**
1. Validate all input parameters
2. Use binding tags for automatic validation
3. Handle all sentinel errors explicitly
4. Return descriptive error messages
5. Use appropriate HTTP status codes
6. Keep handlers thin - delegate to services

### Step 3: Register Route

Add the route to your router configuration:

```go
// File: backend/cmd/server/main.go or routes.go

func setupModerationRoutes(router *gin.Engine, handlers *Handlers) {
    api := router.Group("/api/v1")
    {
        moderation := api.Group("/moderation")
        moderation.Use(middleware.AuthRequired())
        {
            communities := moderation.Group("/communities/:community_id")
            {
                // Existing routes
                communities.POST("/ban", handlers.Moderation.BanUser)
                communities.DELETE("/ban/:user_id", handlers.Moderation.UnbanUser)
                
                // New route
                communities.POST("/warn", handlers.Moderation.WarnUser)
            }
        }
    }
}
```

### Step 4: Write Tests

#### Service Tests

```go
// File: backend/internal/services/moderation_service_test.go

func TestModerationService_WarnUser(t *testing.T) {
    // Setup
    db, cleanup := testutil.SetupTestDB(t)
    defer cleanup()

    userRepo := repository.NewUserRepository(db)
    communityRepo := repository.NewCommunityRepository(db)
    auditLogRepo := repository.NewAuditLogRepository(db)
    warningRepo := repository.NewWarningRepository(db)
    
    service := services.NewModerationService(
        db, communityRepo, userRepo, auditLogRepo, warningRepo,
    )

    // Create test data
    community := testutil.CreateTestCommunity(t, db)
    moderator := testutil.CreateTestModerator(t, db, community.ID)
    targetUser := testutil.CreateTestUser(t, db)

    t.Run("successful warning", func(t *testing.T) {
        err := service.WarnUser(
            context.Background(),
            community.ID,
            moderator.ID,
            targetUser.ID,
            "Please follow community guidelines",
        )
        
        assert.NoError(t, err)
        
        // Verify warning was created
        warnings, err := warningRepo.GetByUserID(context.Background(), targetUser.ID)
        assert.NoError(t, err)
        assert.Len(t, warnings, 1)
        assert.Equal(t, "Please follow community guidelines", warnings[0].Message)
        
        // Verify audit log was created
        logs, err := auditLogRepo.GetByAction(context.Background(), "warn_user", 10)
        assert.NoError(t, err)
        assert.Len(t, logs, 1)
    })

    t.Run("permission denied for non-moderator", func(t *testing.T) {
        regularUser := testutil.CreateTestUser(t, db)
        
        err := service.WarnUser(
            context.Background(),
            community.ID,
            regularUser.ID,
            targetUser.ID,
            "Test warning",
        )
        
        assert.ErrorIs(t, err, services.ErrModerationPermissionDenied)
    })

    t.Run("unauthorized for wrong community", func(t *testing.T) {
        otherCommunity := testutil.CreateTestCommunity(t, db)
        communityMod := testutil.CreateTestCommunityModerator(t, db, otherCommunity.ID)
        
        err := service.WarnUser(
            context.Background(),
            community.ID,  // Different community
            communityMod.ID,
            targetUser.ID,
            "Test warning",
        )
        
        assert.ErrorIs(t, err, services.ErrModerationNotAuthorized)
    })
}
```

#### Handler Tests

```go
// File: backend/internal/handlers/moderation_handler_test.go

func TestModerationHandler_WarnUser(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    
    t.Run("successful warning", func(t *testing.T) {
        // Create mocks
        mockService := new(MockModerationService)
        handler := NewModerationHandler(mockService, nil, nil, nil, nil, nil, nil, nil)

        // Setup expectations
        communityID := uuid.New()
        moderatorID := uuid.New()
        targetUserID := uuid.New()
        
        mockService.On("WarnUser",
            mock.Anything,
            communityID,
            moderatorID,
            targetUserID,
            "Test warning message",
        ).Return(nil)

        // Create request
        reqBody := `{"target_user_id":"` + targetUserID.String() + `","message":"Test warning message"}`
        req := httptest.NewRequest("POST", "/api/v1/moderation/communities/"+communityID.String()+"/warn", strings.NewReader(reqBody))
        req.Header.Set("Content-Type", "application/json")
        
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = req
        c.Set("user_id", moderatorID)
        c.Params = gin.Params{{Key: "community_id", Value: communityID.String()}}

        // Execute
        handler.WarnUser(c)

        // Assert
        assert.Equal(t, http.StatusOK, w.Code)
        
        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        assert.True(t, response["success"].(bool))
        
        mockService.AssertExpectations(t)
    })

    t.Run("permission denied", func(t *testing.T) {
        mockService := new(MockModerationService)
        handler := NewModerationHandler(mockService, nil, nil, nil, nil, nil, nil, nil)

        communityID := uuid.New()
        moderatorID := uuid.New()
        targetUserID := uuid.New()
        
        mockService.On("WarnUser",
            mock.Anything,
            mock.Anything,
            mock.Anything,
            mock.Anything,
            mock.Anything,
        ).Return(services.ErrModerationPermissionDenied)

        reqBody := `{"target_user_id":"` + targetUserID.String() + `","message":"Test warning"}`
        req := httptest.NewRequest("POST", "/api/v1/moderation/communities/"+communityID.String()+"/warn", strings.NewReader(reqBody))
        req.Header.Set("Content-Type", "application/json")
        
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = req
        c.Set("user_id", moderatorID)
        c.Params = gin.Params{{Key: "community_id", Value: communityID.String()}}

        handler.WarnUser(c)

        assert.Equal(t, http.StatusForbidden, w.Code)
        mockService.AssertExpectations(t)
    })
}
```


---

## Service Layer Patterns

The service layer contains business logic and orchestrates operations across multiple repositories.

### Service Structure

```go
// File: backend/internal/services/example_service.go

type ExampleService struct {
    db            *pgxpool.Pool           // For transactions
    exampleRepo   ExampleRepository       // Primary repository
    userRepo      UserRepository          // Related repository
    auditLogRepo  AuditLogRepository      // Audit logging
    cache         *CacheService           // Optional caching
    logger        *zap.Logger             // Structured logging
}

func NewExampleService(
    db *pgxpool.Pool,
    exampleRepo ExampleRepository,
    userRepo UserRepository,
    auditLogRepo AuditLogRepository,
) *ExampleService {
    return &ExampleService{
        db:           db,
        exampleRepo:  exampleRepo,
        userRepo:     userRepo,
        auditLogRepo: auditLogRepo,
        logger:       zap.L().Named("example_service"),
    }
}
```

### Error Handling with Sentinel Errors

Define sentinel errors at the package level for common failure cases:

```go
// File: backend/internal/services/moderation_service.go

// Sentinel errors for moderation operations
var (
    ErrModerationPermissionDenied  = errors.New("insufficient permissions: user does not have moderation privileges")
    ErrModerationNotAuthorized     = errors.New("moderator is not authorized to moderate this community")
    ErrModerationCommunityNotFound = errors.New("community not found")
    ErrModerationUserNotFound      = errors.New("user not found")
    ErrModerationNotBanned         = errors.New("user is not banned from this community")
    ErrModerationCannotBanOwner    = errors.New("cannot ban the community owner")
)
```

Use them in service methods:

```go
func (s *ModerationService) BanUser(ctx context.Context, ...) error {
    // Validation
    if err := s.validateModerationPermission(ctx, moderator, communityID); err != nil {
        return err  // Returns ErrModerationPermissionDenied
    }
    
    // Check community owner
    if community.OwnerID == targetUserID {
        return ErrModerationCannotBanOwner
    }
    
    // Database operation
    if err := s.communityRepo.BanMember(ctx, ban); err != nil {
        return fmt.Errorf("failed to create ban: %w", err)
    }
    
    return nil
}
```

Check errors in handlers:

```go
err := h.moderationService.BanUser(ctx, ...)
if err != nil {
    switch {
    case errors.Is(err, services.ErrModerationPermissionDenied):
        c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
    case errors.Is(err, services.ErrModerationCannotBanOwner):
        c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot ban community owner"})
    default:
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Operation failed"})
    }
    return
}
```

### Transaction Patterns

Use transactions for operations that modify multiple tables:

```go
func (s *ExampleService) ComplexOperation(ctx context.Context, ...) error {
    // Begin transaction
    tx, err := s.db.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)  // Rollback if commit not called

    // Operation 1
    if err := s.exampleRepo.UpdateWithTx(ctx, tx, data); err != nil {
        return fmt.Errorf("failed to update: %w", err)
    }

    // Operation 2
    if err := s.relatedRepo.CreateWithTx(ctx, tx, relatedData); err != nil {
        return fmt.Errorf("failed to create related: %w", err)
    }

    // Operation 3: Create audit log
    auditLog := &models.AuditLog{...}
    if err := s.auditLogRepo.CreateWithTx(ctx, tx, auditLog); err != nil {
        return fmt.Errorf("failed to create audit log: %w", err)
    }

    // Commit transaction
    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}
```

### Validation Patterns

Extract validation logic into helper methods:

```go
// File: backend/internal/services/moderation_service.go

// validateModerationPermission checks if a user has permission to moderate
func (s *ModerationService) validateModerationPermission(
    ctx context.Context,
    moderator *models.User,
    communityID uuid.UUID,
) error {
    // Site moderators can moderate anywhere
    if moderator.AccountType == models.AccountTypeModerator && 
       moderator.ModeratorScope == models.ModeratorScopeSite {
        return nil
    }

    // Admins can moderate anywhere
    if moderator.AccountType == models.AccountTypeAdmin || moderator.Role == models.RoleAdmin {
        return nil
    }

    // Community moderators need community-specific authorization
    if moderator.AccountType == models.AccountTypeCommunityModerator {
        member, err := s.communityRepo.GetMember(ctx, communityID, moderator.ID)
        if err != nil {
            return fmt.Errorf("failed to get member: %w", err)
        }
        if member == nil {
            return ErrModerationPermissionDenied
        }
        if member.Role != models.CommunityRoleMod && member.Role != models.CommunityRoleAdmin {
            return ErrModerationPermissionDenied
        }
        return nil
    }

    return ErrModerationPermissionDenied
}

// validateModerationScope checks if a community moderator is authorized for specific community
func (s *ModerationService) validateModerationScope(
    moderator *models.User,
    communityID uuid.UUID,
) error {
    // Site moderators and admins have no scope restrictions
    if moderator.AccountType == models.AccountTypeModerator && 
       moderator.ModeratorScope == models.ModeratorScopeSite {
        return nil
    }
    if moderator.AccountType == models.AccountTypeAdmin || moderator.Role == models.RoleAdmin {
        return nil
    }

    // Community moderators must have the community in their moderation channels
    if moderator.AccountType == models.AccountTypeCommunityModerator {
        if moderator.ModeratorScope != models.ModeratorScopeCommunity {
            return ErrModerationPermissionDenied
        }

        // Check if this community is in their authorized scope
        for _, authorizedCommunityID := range moderator.ModerationChannels {
            if authorizedCommunityID == communityID {
                return nil
            }
        }
        return ErrModerationNotAuthorized
    }

    return nil
}
```

### Caching Patterns

Use caching for frequently accessed, slowly changing data:

```go
func (s *ExampleService) GetPopularItem(ctx context.Context, id uuid.UUID) (*models.Item, error) {
    cacheKey := fmt.Sprintf("item:%s", id.String())
    
    // Try cache first
    var item models.Item
    err := s.cache.Get(ctx, cacheKey, &item)
    if err == nil {
        s.logger.Debug("cache hit", zap.String("key", cacheKey))
        return &item, nil
    }
    
    // Cache miss - query database
    item, err = s.exampleRepo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get item: %w", err)
    }
    
    // Store in cache (5 minute TTL)
    if err := s.cache.Set(ctx, cacheKey, item, 5*time.Minute); err != nil {
        s.logger.Warn("failed to cache item", zap.Error(err))
        // Don't fail the request if caching fails
    }
    
    return item, nil
}

// Invalidate cache when item is updated
func (s *ExampleService) UpdateItem(ctx context.Context, id uuid.UUID, data *models.Item) error {
    if err := s.exampleRepo.Update(ctx, id, data); err != nil {
        return err
    }
    
    // Invalidate cache
    cacheKey := fmt.Sprintf("item:%s", id.String())
    if err := s.cache.Delete(ctx, cacheKey); err != nil {
        s.logger.Warn("failed to invalidate cache", zap.Error(err))
    }
    
    return nil
}
```

### Real Example from Codebase

Here's the actual `BanUser` method from `moderation_service.go`:

```go
// BanUser bans a user from a community with permission and scope validation
func (s *ModerationService) BanUser(
    ctx context.Context,
    communityID, moderatorID, targetUserID uuid.UUID,
    reason *string,
) error {
    // Get moderator user
    moderator, err := s.userRepo.GetByID(ctx, moderatorID)
    if err != nil {
        return fmt.Errorf("failed to get moderator: %w", err)
    }

    // Validate scope first for better error messages
    if err := s.validateModerationScope(moderator, communityID); err != nil {
        return err
    }

    // Validate permission
    if err := s.validateModerationPermission(ctx, moderator, communityID); err != nil {
        return err
    }

    // Check if target user is the community owner
    community, err := s.communityRepo.GetCommunityByID(ctx, communityID)
    if err != nil {
        return ErrModerationCommunityNotFound
    }
    if community.OwnerID == targetUserID {
        return ErrModerationCannotBanOwner
    }

    // Remove user from community if they are a member
    if err := s.communityRepo.RemoveMember(ctx, communityID, targetUserID); err != nil {
        // Log non-critical errors but continue
        if err.Error() != "member not found" && err.Error() != "no rows affected" {
            // Could log here but continue as this is not critical
        }
    }

    // Create ban record
    ban := &models.CommunityBan{
        ID:             uuid.New(),
        CommunityID:    communityID,
        BannedUserID:   targetUserID,
        BannedByUserID: &moderatorID,
        Reason:         reason,
        BannedAt:       time.Now(),
    }

    if err := s.communityRepo.BanMember(ctx, ban); err != nil {
        return fmt.Errorf("failed to create ban: %w", err)
    }

    // Log audit entry
    metadata := map[string]interface{}{
        "community_id":    communityID.String(),
        "banned_user_id":  targetUserID.String(),
        "moderator_scope": moderator.ModeratorScope,
    }
    if reason != nil {
        metadata["reason"] = *reason
    }

    auditLog := &models.ModerationAuditLog{
        Action:      "ban_user",
        EntityType:  "community_ban",
        EntityID:    ban.ID,
        ModeratorID: moderatorID,
        Reason:      reason,
        Metadata:    metadata,
    }

    if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
        return fmt.Errorf("failed to create audit log: %w", err)
    }

    return nil
}
```


---

## Database Schema & Migrations

### Schema Overview

#### moderation_audit_logs Table

Tracks all moderation actions for compliance and debugging.

```sql
CREATE TABLE moderation_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action VARCHAR(50) NOT NULL,          -- approve, reject, ban_user, unban_user
    entity_type VARCHAR(50) NOT NULL,     -- clip_submission, clip, comment, user, community_ban
    entity_id UUID NOT NULL,              -- ID of the affected entity
    moderator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    reason TEXT,                          -- Optional reason for the action
    metadata JSONB,                       -- Additional context (filters, bulk actions, etc.)
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for efficient queries
CREATE INDEX idx_audit_logs_moderator ON moderation_audit_logs(moderator_id);
CREATE INDEX idx_audit_logs_entity ON moderation_audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_created ON moderation_audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_action ON moderation_audit_logs(action);
```

**Key Fields:**
- `action` - What was done (ban_user, approve, reject, etc.)
- `entity_type` - Type of thing affected (user, clip, comment, etc.)
- `entity_id` - ID of the specific entity
- `moderator_id` - Who performed the action
- `metadata` - JSON field for additional context

#### community_bans Table

Stores user bans from communities.

```sql
CREATE TABLE community_bans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    banned_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    banned_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    reason TEXT,
    banned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(community_id, banned_user_id)  -- One ban per user per community
);

-- Indexes
CREATE INDEX idx_community_bans_community_id ON community_bans(community_id);
CREATE INDEX idx_community_bans_banned_user_id ON community_bans(banned_user_id);
```

**Key Constraints:**
- `UNIQUE(community_id, banned_user_id)` - Prevents duplicate bans
- `ON DELETE CASCADE` - Auto-delete bans if community is deleted
- `ON DELETE SET NULL` - Preserve ban if moderator account is deleted

### Creating Migrations

Migrations use golang-migrate with up/down SQL files.

#### Step 1: Create Migration Files

```bash
# Generate migration files with sequential number
cd backend
migrate create -ext sql -dir migrations -seq add_user_warnings
```

This creates two files:
- `000XXX_add_user_warnings.up.sql` - Applies changes
- `000XXX_add_user_warnings.down.sql` - Reverts changes

#### Step 2: Write Up Migration

```sql
-- File: backend/migrations/000XXX_add_user_warnings.up.sql

-- User warnings table
CREATE TABLE user_warnings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    moderator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    message TEXT NOT NULL CHECK (LENGTH(message) >= 10 AND LENGTH(message) <= 500),
    severity VARCHAR(20) NOT NULL DEFAULT 'low' CHECK (severity IN ('low', 'medium', 'high')),
    acknowledged BOOLEAN DEFAULT FALSE,
    acknowledged_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP  -- Optional expiration
);

-- Indexes for efficient queries
CREATE INDEX idx_user_warnings_user_id ON user_warnings(user_id, created_at DESC);
CREATE INDEX idx_user_warnings_community_id ON user_warnings(community_id);
CREATE INDEX idx_user_warnings_moderator_id ON user_warnings(moderator_id);
CREATE INDEX idx_user_warnings_acknowledged ON user_warnings(acknowledged) WHERE acknowledged = false;
CREATE INDEX idx_user_warnings_severity ON user_warnings(severity) WHERE severity IN ('medium', 'high');

-- Add comment for documentation
COMMENT ON TABLE user_warnings IS 'Stores moderation warnings issued to users';
COMMENT ON COLUMN user_warnings.severity IS 'Warning severity: low, medium, high';
COMMENT ON COLUMN user_warnings.expires_at IS 'Optional expiration time for the warning';
```

#### Step 3: Write Down Migration

```sql
-- File: backend/migrations/000XXX_add_user_warnings.down.sql

-- Drop indexes first
DROP INDEX IF EXISTS idx_user_warnings_severity;
DROP INDEX IF EXISTS idx_user_warnings_acknowledged;
DROP INDEX IF EXISTS idx_user_warnings_moderator_id;
DROP INDEX IF EXISTS idx_user_warnings_community_id;
DROP INDEX IF EXISTS idx_user_warnings_user_id;

-- Drop table
DROP TABLE IF EXISTS user_warnings;
```

#### Step 4: Apply Migration

```bash
# Apply migration
make migrate-up

# Or manually:
migrate -path backend/migrations \
        -database "postgresql://user:pass@localhost:5432/clpr?sslmode=disable" \
        up

# Verify
psql -d clpr -c "\d user_warnings"
```

#### Step 5: Test Rollback

```bash
# Rollback last migration
make migrate-down

# Or manually:
migrate -path backend/migrations \
        -database "postgresql://user:pass@localhost:5432/clpr?sslmode=disable" \
        down 1
```

### Migration Best Practices

1. **Always write both up and down migrations**
   - Test rollback before committing
   - Ensure down migration fully reverts changes

2. **Use constraints for data integrity**
   ```sql
   -- Good: Enforces data quality at DB level
   CHECK (LENGTH(message) >= 10 AND LENGTH(message) <= 500)
   CHECK (severity IN ('low', 'medium', 'high'))
   ```

3. **Add indexes for query performance**
   ```sql
   -- Index commonly queried columns
   CREATE INDEX idx_table_user_id ON table(user_id);
   
   -- Partial index for filtered queries
   CREATE INDEX idx_table_active ON table(status) WHERE status = 'active';
   
   -- Composite index for multi-column queries
   CREATE INDEX idx_table_user_created ON table(user_id, created_at DESC);
   ```

4. **Use appropriate foreign key actions**
   ```sql
   -- CASCADE: Delete children when parent is deleted
   REFERENCES communities(id) ON DELETE CASCADE
   
   -- RESTRICT: Prevent deletion if children exist
   REFERENCES users(id) ON DELETE RESTRICT
   
   -- SET NULL: Set to NULL when parent is deleted
   REFERENCES users(id) ON DELETE SET NULL
   ```

5. **Add comments for documentation**
   ```sql
   COMMENT ON TABLE user_warnings IS 'Stores moderation warnings';
   COMMENT ON COLUMN user_warnings.severity IS 'low, medium, high';
   ```

6. **Use transactions for complex migrations**
   ```sql
   BEGIN;
   
   -- Multiple operations
   ALTER TABLE users ADD COLUMN new_field VARCHAR(50);
   UPDATE users SET new_field = 'default_value';
   ALTER TABLE users ALTER COLUMN new_field SET NOT NULL;
   
   COMMIT;
   ```

### Real Example: Moderation Queue System

Here's the actual migration for the moderation queue:

```sql
-- File: backend/migrations/000049_add_moderation_queue_system.up.sql

-- Moderation Queue Table
CREATE TABLE moderation_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_type VARCHAR(20) NOT NULL,
    content_id UUID NOT NULL,
    reason VARCHAR(50) NOT NULL,
    priority INT DEFAULT 50,
    status VARCHAR(20) DEFAULT 'pending',
    assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,
    reported_by UUID[] DEFAULT '{}',
    report_count INT DEFAULT 0,
    auto_flagged BOOLEAN DEFAULT FALSE,
    confidence_score DECIMAL(3,2),
    created_at TIMESTAMP DEFAULT NOW(),
    reviewed_at TIMESTAMP,
    reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT moderation_queue_valid_content_type CHECK (content_type IN ('comment', 'clip', 'user', 'submission')),
    CONSTRAINT moderation_queue_valid_status CHECK (status IN ('pending', 'approved', 'rejected', 'escalated')),
    CONSTRAINT moderation_queue_valid_priority CHECK (priority >= 0 AND priority <= 100),
    CONSTRAINT moderation_queue_valid_confidence CHECK (confidence_score IS NULL OR (confidence_score >= 0 AND confidence_score <= 1))
);

-- Indexes for efficient queue queries
CREATE INDEX idx_modqueue_status_priority ON moderation_queue(status, priority DESC, created_at);
CREATE INDEX idx_modqueue_content ON moderation_queue(content_type, content_id);
CREATE INDEX idx_modqueue_assigned_to ON moderation_queue(assigned_to) WHERE status = 'pending';
CREATE INDEX idx_modqueue_auto_flagged ON moderation_queue(auto_flagged) WHERE status = 'pending';

-- Prevent duplicate entries
CREATE UNIQUE INDEX uq_modqueue_content_pending ON moderation_queue(content_type, content_id) 
WHERE status = 'pending';

-- Trigger to auto-update reviewed_at
CREATE OR REPLACE FUNCTION update_moderation_queue_reviewed()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status != 'pending' AND OLD.status = 'pending' THEN
        NEW.reviewed_at = NOW();
        IF NEW.reviewed_by IS NULL THEN
            RAISE EXCEPTION 'reviewed_by must be set when changing status from pending';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_moderation_queue_reviewed
    BEFORE UPDATE ON moderation_queue
    FOR EACH ROW
    EXECUTE FUNCTION update_moderation_queue_reviewed();
```


---

## API Endpoint Patterns

### RESTful Design

Follow REST conventions for moderation endpoints:

```
GET    /api/v1/moderation/communities/:id/bans       - List bans
POST   /api/v1/moderation/communities/:id/ban        - Create ban
DELETE /api/v1/moderation/communities/:id/bans/:user_id  - Remove ban
GET    /api/v1/moderation/communities/:id/bans/:user_id  - Get ban details
PATCH  /api/v1/moderation/communities/:id/bans/:user_id  - Update ban

GET    /api/v1/moderation/audit-logs                 - List audit logs
GET    /api/v1/moderation/audit-logs/:id             - Get specific log
GET    /api/v1/moderation/audit-logs/export          - Export logs (CSV)
```

### Handler Template

Standard handler structure:

```go
// HandlerName handles [description]
// METHOD /api/v1/path
func (h *Handler) HandlerName(c *gin.Context) {
    ctx := c.Request.Context()

    // 1. Parse and validate path parameters
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid ID format",
        })
        return
    }

    // 2. Get authenticated user from context
    userIDVal, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error": "Unauthorized",
        })
        return
    }
    userID := userIDVal.(uuid.UUID)

    // 3. Parse and validate query parameters (for GET)
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
    if limit < 1 || limit > 100 {
        limit = 50
    }

    // 4. Parse and validate request body (for POST/PUT/PATCH)
    var req RequestStruct
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid request body",
            "details": err.Error(),
        })
        return
    }

    // 5. Call service method
    result, err := h.service.MethodName(ctx, id, userID, req)
    
    // 6. Handle errors with appropriate status codes
    if err != nil {
        switch {
        case errors.Is(err, services.ErrNotFound):
            c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
        case errors.Is(err, services.ErrPermissionDenied):
            c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
        case errors.Is(err, services.ErrValidation):
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Operation failed"})
        }
        return
    }

    // 7. Return success response
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    result,
    })
}
```

### Request Validation

Use binding tags for automatic validation:

```go
type CreateBanRequest struct {
    TargetUserID string  `json:"target_user_id" binding:"required,uuid"`
    Reason       *string `json:"reason" binding:"omitempty,min=10,max=500"`
    Duration     *int    `json:"duration" binding:"omitempty,min=1,max=87600"` // Max 10 years in hours
}

type UpdateBanRequest struct {
    Reason *string `json:"reason" binding:"required,min=10,max=500"`
}

type ListBansRequest struct {
    Page   int    `form:"page" binding:"omitempty,min=1"`
    Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
    Status string `form:"status" binding:"omitempty,oneof=active expired"`
}
```

**Common Validation Tags:**
- `required` - Field must be present
- `omitempty` - Skip validation if empty
- `min=X`, `max=X` - Length/value limits
- `uuid` - Must be valid UUID
- `email` - Must be valid email
- `oneof=a b c` - Must be one of specified values
- `gte=X`, `lte=X` - Greater/less than or equal

### Error Response Format

Standard error response:

```go
// Single error
c.JSON(http.StatusBadRequest, gin.H{
    "error": "Invalid request",
})

// Error with details
c.JSON(http.StatusBadRequest, gin.H{
    "error":   "Validation failed",
    "details": validationErrors,
})

// Error with code for client handling
c.JSON(http.StatusForbidden, gin.H{
    "error": "Insufficient permissions",
    "code":  "PERMISSION_DENIED",
})
```

### Success Response Format

Standard success response:

```go
// Simple success
c.JSON(http.StatusOK, gin.H{
    "success": true,
})

// Success with data
c.JSON(http.StatusOK, gin.H{
    "success": true,
    "data":    result,
})

// Success with metadata
c.JSON(http.StatusOK, gin.H{
    "success": true,
    "data":    items,
    "meta": gin.H{
        "total": totalCount,
        "page":  page,
        "limit": limit,
    },
})

// Created resource (201)
c.JSON(http.StatusCreated, gin.H{
    "success": true,
    "data":    newResource,
    "message": "Resource created successfully",
})
```

### Middleware Stack

Typical middleware for protected endpoints:

```go
// Route configuration
moderation := api.Group("/moderation")
moderation.Use(
    middleware.AuthRequired(),              // Verify JWT token
    middleware.RateLimiter("moderation"),   // Rate limiting
    middleware.RequestLogger(),             // Log requests
)
{
    // Public moderation info (authenticated users)
    moderation.GET("/stats", handler.GetStats)
    
    // Moderator-only endpoints
    moderatorRoutes := moderation.Group("")
    moderatorRoutes.Use(middleware.RequireModeratorRole())
    {
        moderatorRoutes.GET("/queue", handler.GetQueue)
        moderatorRoutes.POST("/:id/approve", handler.Approve)
        moderatorRoutes.POST("/:id/reject", handler.Reject)
    }
    
    // Admin-only endpoints
    adminRoutes := moderation.Group("/admin")
    adminRoutes.Use(middleware.RequireAdminRole())
    {
        adminRoutes.POST("/users/:id/ban", handler.BanUser)
        adminRoutes.DELETE("/users/:id/ban", handler.UnbanUser)
    }
}
```

### Real Handler Example

Actual `BanUser` handler from the codebase:

```go
// BanUser bans a user from a community
// POST /api/v1/moderation/communities/:community_id/ban
func (h *ModerationHandler) BanUser(c *gin.Context) {
    ctx := c.Request.Context()

    // Parse community ID
    communityID, err := uuid.Parse(c.Param("community_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid community ID",
        })
        return
    }

    // Get moderator ID from authenticated context
    moderatorIDVal, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error": "Unauthorized",
        })
        return
    }
    moderatorID := moderatorIDVal.(uuid.UUID)

    // Parse request body
    var req struct {
        TargetUserID string  `json:"target_user_id" binding:"required,uuid"`
        Reason       *string `json:"reason" binding:"omitempty,min=10,max=500"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid request",
            "details": err.Error(),
        })
        return
    }

    targetUserID, err := uuid.Parse(req.TargetUserID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid target user ID",
        })
        return
    }

    // Call service
    err = h.moderationService.BanUser(ctx, communityID, moderatorID, targetUserID, req.Reason)
    if err != nil {
        switch {
        case errors.Is(err, services.ErrModerationPermissionDenied):
            c.JSON(http.StatusForbidden, gin.H{
                "error": "You do not have permission to ban users",
            })
        case errors.Is(err, services.ErrModerationNotAuthorized):
            c.JSON(http.StatusForbidden, gin.H{
                "error": "You are not authorized to moderate this community",
            })
        case errors.Is(err, services.ErrModerationCannotBanOwner):
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Cannot ban the community owner",
            })
        default:
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Failed to ban user",
            })
        }
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "User banned successfully",
    })
}
```


---

## Frontend Integration

### API Client Patterns

Create typed API client functions:

```typescript
// File: frontend/src/lib/moderation-api.ts

import { apiClient } from './api';

export interface BanRequest {
    target_user_id: string;
    reason?: string;
    duration?: number;
}

export interface Ban {
    id: string;
    community_id: string;
    banned_user_id: string;
    banned_by_user_id?: string;
    reason?: string;
    banned_at: string;
    user?: {
        id: string;
        username: string;
        display_name?: string;
    };
}

export interface BanListResponse {
    success: boolean;
    data: Ban[];
    meta: {
        total: number;
        page: number;
        limit: number;
    };
}

/**
 * Ban a user from a community
 */
export async function banUser(
    communityId: string,
    targetUserId: string,
    reason?: string
): Promise<{ success: boolean; message: string }> {
    const response = await apiClient.post(
        `/moderation/communities/${communityId}/ban`,
        {
            target_user_id: targetUserId,
            reason,
        }
    );
    return response.data;
}

/**
 * Unban a user from a community
 */
export async function unbanUser(
    communityId: string,
    userId: string
): Promise<{ success: boolean; message: string }> {
    const response = await apiClient.delete(
        `/moderation/communities/${communityId}/bans/${userId}`
    );
    return response.data;
}

/**
 * Get list of bans for a community
 */
export async function getBans(
    communityId: string,
    page: number = 1,
    limit: number = 20
): Promise<BanListResponse> {
    const response = await apiClient.get<BanListResponse>(
        `/moderation/communities/${communityId}/bans`,
        {
            params: { page, limit },
        }
    );
    return response.data;
}
```

### React Hook Patterns

Create custom hooks for moderation operations:

```typescript
// File: frontend/src/hooks/useModeration.ts

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { banUser, unbanUser, getBans } from '@/lib/moderation-api';
import { useToast } from '@/hooks/useToast';

/**
 * Hook for banning a user
 */
export function useBanUser(communityId: string) {
    const queryClient = useQueryClient();
    const { toast } = useToast();

    return useMutation({
        mutationFn: ({ userId, reason }: { userId: string; reason?: string }) =>
            banUser(communityId, userId, reason),
        onSuccess: () => {
            // Invalidate bans list to refetch
            queryClient.invalidateQueries(['bans', communityId]);
            toast({
                title: 'User banned',
                description: 'The user has been banned successfully',
                variant: 'success',
            });
        },
        onError: (error: any) => {
            toast({
                title: 'Failed to ban user',
                description: error.response?.data?.error || 'An error occurred',
                variant: 'destructive',
            });
        },
    });
}

/**
 * Hook for unbanning a user
 */
export function useUnbanUser(communityId: string) {
    const queryClient = useQueryClient();
    const { toast } = useToast();

    return useMutation({
        mutationFn: (userId: string) => unbanUser(communityId, userId),
        onSuccess: () => {
            queryClient.invalidateQueries(['bans', communityId]);
            toast({
                title: 'User unbanned',
                description: 'The ban has been removed successfully',
                variant: 'success',
            });
        },
        onError: (error: any) => {
            toast({
                title: 'Failed to unban user',
                description: error.response?.data?.error || 'An error occurred',
                variant: 'destructive',
            });
        },
    });
}

/**
 * Hook for fetching bans list
 */
export function useBans(communityId: string, page: number = 1, limit: number = 20) {
    return useQuery({
        queryKey: ['bans', communityId, page, limit],
        queryFn: () => getBans(communityId, page, limit),
        keepPreviousData: true,
        staleTime: 30000, // 30 seconds
    });
}
```

### Component Patterns

#### Ban User Dialog Component

```typescript
// File: frontend/src/components/moderation/BanUserDialog.tsx

import { useState } from 'react';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { useBanUser } from '@/hooks/useModeration';

interface BanUserDialogProps {
    communityId: string;
    userId: string;
    username: string;
    open: boolean;
    onOpenChange: (open: boolean) => void;
}

export function BanUserDialog({
    communityId,
    userId,
    username,
    open,
    onOpenChange,
}: BanUserDialogProps) {
    const [reason, setReason] = useState('');
    const banUser = useBanUser(communityId);

    const handleBan = async () => {
        await banUser.mutateAsync({
            userId,
            reason: reason || undefined,
        });
        onOpenChange(false);
        setReason('');
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>Ban User</DialogTitle>
                    <DialogDescription>
                        Are you sure you want to ban {username}? This action can be reversed later.
                    </DialogDescription>
                </DialogHeader>

                <div className="space-y-4">
                    <div>
                        <label className="text-sm font-medium">
                            Reason (optional)
                        </label>
                        <Textarea
                            value={reason}
                            onChange={(e) => setReason(e.target.value)}
                            placeholder="Explain why this user is being banned..."
                            className="mt-1"
                            rows={4}
                            maxLength={500}
                        />
                        <p className="text-xs text-muted-foreground mt-1">
                            {reason.length}/500 characters
                        </p>
                    </div>
                </div>

                <DialogFooter>
                    <Button
                        variant="outline"
                        onClick={() => onOpenChange(false)}
                        disabled={banUser.isLoading}
                    >
                        Cancel
                    </Button>
                    <Button
                        variant="destructive"
                        onClick={handleBan}
                        disabled={banUser.isLoading}
                    >
                        {banUser.isLoading ? 'Banning...' : 'Ban User'}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
```

#### Bans List Component

```typescript
// File: frontend/src/components/moderation/BansList.tsx

import { useState } from 'react';
import { useBans, useUnbanUser } from '@/hooks/useModeration';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { formatDistanceToNow } from 'date-fns';

interface BansListProps {
    communityId: string;
}

export function BansList({ communityId }: BansListProps) {
    const [page, setPage] = useState(1);
    const { data, isLoading, error } = useBans(communityId, page, 20);
    const unbanUser = useUnbanUser(communityId);

    if (isLoading) {
        return <div>Loading bans...</div>;
    }

    if (error) {
        return <div>Error loading bans: {error.message}</div>;
    }

    if (!data || data.data.length === 0) {
        return (
            <Card className="p-6 text-center text-muted-foreground">
                No banned users
            </Card>
        );
    }

    return (
        <div className="space-y-4">
            {data.data.map((ban) => (
                <Card key={ban.id} className="p-4">
                    <div className="flex items-center justify-between">
                        <div>
                            <h4 className="font-medium">
                                {ban.user?.display_name || ban.user?.username || 'Unknown User'}
                            </h4>
                            <p className="text-sm text-muted-foreground">
                                Banned {formatDistanceToNow(new Date(ban.banned_at), { addSuffix: true })}
                            </p>
                            {ban.reason && (
                                <p className="text-sm mt-2">
                                    <span className="font-medium">Reason:</span> {ban.reason}
                                </p>
                            )}
                        </div>
                        <Button
                            variant="outline"
                            onClick={() => unbanUser.mutate(ban.banned_user_id)}
                            disabled={unbanUser.isLoading}
                        >
                            Unban
                        </Button>
                    </div>
                </Card>
            ))}

            {data.meta.total > 20 && (
                <div className="flex justify-between items-center">
                    <Button
                        variant="outline"
                        onClick={() => setPage(page - 1)}
                        disabled={page === 1}
                    >
                        Previous
                    </Button>
                    <span className="text-sm text-muted-foreground">
                        Page {page} of {Math.ceil(data.meta.total / 20)}
                    </span>
                    <Button
                        variant="outline"
                        onClick={() => setPage(page + 1)}
                        disabled={page * 20 >= data.meta.total}
                    >
                        Next
                    </Button>
                </div>
            )}
        </div>
    );
}
```

### Permission-Based UI

Show/hide UI elements based on permissions:

```typescript
// File: frontend/src/components/moderation/ModerationActions.tsx

import { useUser } from '@/hooks/useUser';
import { Button } from '@/components/ui/button';
import { BanUserDialog } from './BanUserDialog';
import { useState } from 'react';

interface ModerationActionsProps {
    communityId: string;
    targetUserId: string;
    targetUsername: string;
}

export function ModerationActions({
    communityId,
    targetUserId,
    targetUsername,
}: ModerationActionsProps) {
    const { user } = useUser();
    const [showBanDialog, setShowBanDialog] = useState(false);

    // Check if user has moderation permissions
    const canModerate = user?.permissions?.includes('moderate:users') || 
                       user?.permissions?.includes('community:moderate');

    // Don't render anything if user can't moderate
    if (!canModerate) {
        return null;
    }

    return (
        <>
            <div className="flex gap-2">
                <Button
                    variant="outline"
                    onClick={() => {/* Warn user logic */}}
                >
                    Warn User
                </Button>
                <Button
                    variant="destructive"
                    onClick={() => setShowBanDialog(true)}
                >
                    Ban User
                </Button>
            </div>

            <BanUserDialog
                communityId={communityId}
                userId={targetUserId}
                username={targetUsername}
                open={showBanDialog}
                onOpenChange={setShowBanDialog}
            />
        </>
    );
}
```


---

## Testing Strategies

### Unit Testing Services

Test service methods with mocked dependencies:

```go
// File: backend/internal/services/moderation_service_test.go

package services_test

import (
    "context"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "git.subcult.tv/subculture-collective/clpr/internal/models"
    "git.subcult.tv/subculture-collective/clpr/internal/services"
)

// Mock repositories
type MockCommunityRepo struct {
    mock.Mock
}

func (m *MockCommunityRepo) GetCommunityByID(ctx context.Context, id uuid.UUID) (*models.Community, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.Community), args.Error(1)
}

func (m *MockCommunityRepo) BanMember(ctx context.Context, ban *models.CommunityBan) error {
    args := m.Called(ctx, ban)
    return args.Error(0)
}

type MockUserRepo struct {
    mock.Mock
}

func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.User), args.Error(1)
}

type MockAuditLogRepo struct {
    mock.Mock
}

func (m *MockAuditLogRepo) Create(ctx context.Context, log *models.ModerationAuditLog) error {
    args := m.Called(ctx, log)
    return args.Error(0)
}

func TestModerationService_BanUser(t *testing.T) {
    t.Run("successful ban by site moderator", func(t *testing.T) {
        // Setup mocks
        communityRepo := new(MockCommunityRepo)
        userRepo := new(MockUserRepo)
        auditLogRepo := new(MockAuditLogRepo)

        service := services.NewModerationService(
            nil, // db pool not needed for unit test
            communityRepo,
            userRepo,
            auditLogRepo,
        )

        // Test data
        communityID := uuid.New()
        moderatorID := uuid.New()
        targetUserID := uuid.New()
        reason := "Spam"

        moderator := &models.User{
            ID:             moderatorID,
            AccountType:    models.AccountTypeModerator,
            ModeratorScope: models.ModeratorScopeSite,
        }

        community := &models.Community{
            ID:      communityID,
            OwnerID: uuid.New(), // Different from target
        }

        // Setup expectations
        userRepo.On("GetByID", mock.Anything, moderatorID).Return(moderator, nil)
        communityRepo.On("GetCommunityByID", mock.Anything, communityID).Return(community, nil)
        communityRepo.On("RemoveMember", mock.Anything, communityID, targetUserID).Return(nil)
        communityRepo.On("BanMember", mock.Anything, mock.AnythingOfType("*models.CommunityBan")).Return(nil)
        auditLogRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.ModerationAuditLog")).Return(nil)

        // Execute
        err := service.BanUser(context.Background(), communityID, moderatorID, targetUserID, &reason)

        // Assert
        assert.NoError(t, err)
        communityRepo.AssertExpectations(t)
        userRepo.AssertExpectations(t)
        auditLogRepo.AssertExpectations(t)
    })

    t.Run("permission denied for regular user", func(t *testing.T) {
        communityRepo := new(MockCommunityRepo)
        userRepo := new(MockUserRepo)
        auditLogRepo := new(MockAuditLogRepo)

        service := services.NewModerationService(nil, communityRepo, userRepo, auditLogRepo)

        moderatorID := uuid.New()
        regularUser := &models.User{
            ID:          moderatorID,
            AccountType: models.AccountTypeMember, // Not a moderator
        }

        userRepo.On("GetByID", mock.Anything, moderatorID).Return(regularUser, nil)

        err := service.BanUser(context.Background(), uuid.New(), moderatorID, uuid.New(), nil)

        assert.ErrorIs(t, err, services.ErrModerationPermissionDenied)
        userRepo.AssertExpectations(t)
    })

    t.Run("cannot ban community owner", func(t *testing.T) {
        communityRepo := new(MockCommunityRepo)
        userRepo := new(MockUserRepo)
        auditLogRepo := new(MockAuditLogRepo)

        service := services.NewModerationService(nil, communityRepo, userRepo, auditLogRepo)

        communityID := uuid.New()
        moderatorID := uuid.New()
        ownerID := uuid.New()

        moderator := &models.User{
            ID:             moderatorID,
            AccountType:    models.AccountTypeModerator,
            ModeratorScope: models.ModeratorScopeSite,
        }

        community := &models.Community{
            ID:      communityID,
            OwnerID: ownerID, // Same as target
        }

        userRepo.On("GetByID", mock.Anything, moderatorID).Return(moderator, nil)
        communityRepo.On("GetCommunityByID", mock.Anything, communityID).Return(community, nil)

        err := service.BanUser(context.Background(), communityID, moderatorID, ownerID, nil)

        assert.ErrorIs(t, err, services.ErrModerationCannotBanOwner)
    })
}
```

### Integration Testing with Real Database

Test with actual PostgreSQL database:

```go
// File: backend/internal/services/moderation_service_integration_test.go

// +build integration

package services_test

import (
    "context"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "git.subcult.tv/subculture-collective/clpr/internal/models"
    "git.subcult.tv/subculture-collective/clpr/internal/repository"
    "git.subcult.tv/subculture-collective/clpr/internal/services"
    "git.subcult.tv/subculture-collective/clpr/internal/testutil"
)

func TestModerationService_Integration(t *testing.T) {
    // Setup test database
    db, cleanup := testutil.SetupTestDB(t)
    defer cleanup()

    // Create repositories
    communityRepo := repository.NewCommunityRepository(db)
    userRepo := repository.NewUserRepository(db)
    auditLogRepo := repository.NewAuditLogRepository(db)

    // Create service
    service := services.NewModerationService(db, communityRepo, userRepo, auditLogRepo)

    t.Run("full ban workflow", func(t *testing.T) {
        ctx := context.Background()

        // Create test community
        community := &models.Community{
            ID:          uuid.New(),
            Name:        "Test Community",
            Slug:        "test-community",
            OwnerID:     uuid.New(),
            Description: "Test",
        }
        err := communityRepo.Create(ctx, community)
        require.NoError(t, err)

        // Create moderator
        moderator := &models.User{
            ID:             uuid.New(),
            Username:       "mod_user",
            Email:          "mod@example.com",
            AccountType:    models.AccountTypeModerator,
            ModeratorScope: models.ModeratorScopeSite,
        }
        err = userRepo.Create(ctx, moderator)
        require.NoError(t, err)

        // Create target user
        targetUser := &models.User{
            ID:       uuid.New(),
            Username: "bad_user",
            Email:    "bad@example.com",
        }
        err = userRepo.Create(ctx, targetUser)
        require.NoError(t, err)

        // Ban user
        reason := "Violating community rules"
        err = service.BanUser(ctx, community.ID, moderator.ID, targetUser.ID, &reason)
        assert.NoError(t, err)

        // Verify ban exists
        isBanned, err := communityRepo.IsBanned(ctx, community.ID, targetUser.ID)
        assert.NoError(t, err)
        assert.True(t, isBanned)

        // Verify audit log was created
        logs, err := auditLogRepo.GetByModerator(ctx, moderator.ID, 10)
        assert.NoError(t, err)
        assert.Len(t, logs, 1)
        assert.Equal(t, "ban_user", logs[0].Action)

        // Unban user
        err = service.UnbanUser(ctx, community.ID, moderator.ID, targetUser.ID)
        assert.NoError(t, err)

        // Verify ban is removed
        isBanned, err = communityRepo.IsBanned(ctx, community.ID, targetUser.ID)
        assert.NoError(t, err)
        assert.False(t, isBanned)
    })
}
```

### Handler Testing

Test HTTP handlers with mock services:

```go
// File: backend/internal/handlers/moderation_handler_test.go

package handlers_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "git.subcult.tv/subculture-collective/clpr/internal/handlers"
    "git.subcult.tv/subculture-collective/clpr/internal/services"
)

func TestBanUserHandler(t *testing.T) {
    gin.SetMode(gin.TestMode)

    t.Run("successful ban", func(t *testing.T) {
        // Setup
        mockService := new(MockModerationService)
        handler := handlers.NewModerationHandler(mockService, nil, nil, nil, nil, nil, nil, nil)

        communityID := uuid.New()
        moderatorID := uuid.New()
        targetUserID := uuid.New()

        mockService.On("BanUser",
            mock.Anything,
            communityID,
            moderatorID,
            targetUserID,
            mock.Anything,
        ).Return(nil)

        // Create request
        reqBody := map[string]interface{}{
            "target_user_id": targetUserID.String(),
            "reason":         "Spam",
        }
        body, _ := json.Marshal(reqBody)
        req := httptest.NewRequest("POST", "/api/v1/moderation/communities/"+communityID.String()+"/ban", bytes.NewReader(body))
        req.Header.Set("Content-Type", "application/json")

        // Create context
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = req
        c.Set("user_id", moderatorID)
        c.Params = gin.Params{{Key: "community_id", Value: communityID.String()}}

        // Execute
        handler.BanUser(c)

        // Assert
        assert.Equal(t, http.StatusOK, w.Code)

        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        assert.True(t, response["success"].(bool))

        mockService.AssertExpectations(t)
    })
}
```

### E2E Testing with Playwright

Test complete workflows from frontend to backend:

```typescript
// File: frontend/e2e/tests/moderation.spec.ts

import { test, expect } from '@playwright/test';

test.describe('Moderation', () => {
    test('moderator can ban and unban user', async ({ page }) => {
        // Login as moderator
        await page.goto('/login');
        await page.fill('[name=email]', 'moderator@example.com');
        await page.fill('[name=password]', 'password');
        await page.click('button[type=submit]');

        // Navigate to community
        await page.goto('/communities/test-community');
        await expect(page).toHaveURL(/\/communities\/test-community/);

        // Open user menu
        await page.click('[data-testid=user-menu]');
        await page.click('[data-testid=ban-user]');

        // Fill ban dialog
        await expect(page.locator('[data-testid=ban-dialog]')).toBeVisible();
        await page.fill('[name=reason]', 'Violating community guidelines');
        await page.click('[data-testid=confirm-ban]');

        // Verify success
        await expect(page.locator('[data-testid=toast-success]')).toBeVisible();
        await expect(page.locator('[data-testid=toast-success]')).toContainText('User banned');

        // Verify user appears in bans list
        await page.goto('/communities/test-community/moderation/bans');
        await expect(page.locator('[data-testid=ban-item]')).toContainText('bad_user');

        // Unban user
        await page.click('[data-testid=unban-button]');
        await expect(page.locator('[data-testid=toast-success]')).toContainText('User unbanned');

        // Verify user removed from bans list
        await expect(page.locator('[data-testid=ban-item]')).not.toContainText('bad_user');
    });

    test('regular user cannot access moderation tools', async ({ page }) => {
        // Login as regular user
        await page.goto('/login');
        await page.fill('[name=email]', 'user@example.com');
        await page.fill('[name=password]', 'password');
        await page.click('button[type=submit]');

        // Navigate to community
        await page.goto('/communities/test-community');

        // Verify moderation tools are not visible
        await expect(page.locator('[data-testid=ban-user]')).not.toBeVisible();

        // Try to access moderation page directly
        await page.goto('/communities/test-community/moderation/bans');
        await expect(page).toHaveURL(/\/403/); // Redirected to forbidden page
    });
});
```

### Load Testing with k6

Test performance under load:

```javascript
// File: tests/load/moderation.js

import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '30s', target: 20 },  // Ramp up to 20 users
        { duration: '1m', target: 20 },   // Stay at 20 users
        { duration: '30s', target: 0 },   // Ramp down
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% of requests under 500ms
        http_req_failed: ['rate<0.01'],   // Less than 1% failures
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TOKEN = __ENV.AUTH_TOKEN; // Get from env

export default function () {
    const headers = {
        'Authorization': `Bearer ${TOKEN}`,
        'Content-Type': 'application/json',
    };

    // Get bans list
    const bansRes = http.get(
        `${BASE_URL}/api/v1/moderation/communities/test-community/bans`,
        { headers }
    );
    check(bansRes, {
        'bans list status 200': (r) => r.status === 200,
        'bans list has data': (r) => JSON.parse(r.body).data !== undefined,
    });

    sleep(1);

    // Get audit logs
    const auditRes = http.get(
        `${BASE_URL}/api/v1/moderation/audit-logs?limit=50`,
        { headers }
    );
    check(auditRes, {
        'audit logs status 200': (r) => r.status === 200,
        'audit logs response time OK': (r) => r.timings.duration < 300,
    });

    sleep(1);
}
```

Run load test:

```bash
k6 run tests/load/moderation.js \
    --env BASE_URL=https://api.clpr.tv \
    --env AUTH_TOKEN=your_token_here
```


---

## Logging & Audit Trails

### Audit Log Structure

Every moderation action creates an audit log entry:

```go
type ModerationAuditLog struct {
    ID          uuid.UUID              `json:"id"`
    Action      string                 `json:"action"`       // ban_user, unban_user, approve, reject
    EntityType  string                 `json:"entity_type"`  // community_ban, clip, comment
    EntityID    uuid.UUID              `json:"entity_id"`    // ID of affected entity
    ModeratorID uuid.UUID              `json:"moderator_id"` // Who performed the action
    Reason      *string                `json:"reason"`       // Optional reason
    Metadata    map[string]interface{} `json:"metadata"`     // Additional context
    CreatedAt   time.Time              `json:"created_at"`
}
```

### Creating Audit Logs

Always create audit logs for moderation actions:

```go
func (s *ModerationService) BanUser(...) error {
    // ... perform ban operation ...

    // Create audit log with rich metadata
    metadata := map[string]interface{}{
        "community_id":    communityID.String(),
        "banned_user_id":  targetUserID.String(),
        "moderator_scope": moderator.ModeratorScope,
        "moderator_type":  moderator.AccountType,
    }
    if reason != nil {
        metadata["reason"] = *reason
    }

    auditLog := &models.ModerationAuditLog{
        Action:      "ban_user",
        EntityType:  "community_ban",
        EntityID:    ban.ID,
        ModeratorID: moderatorID,
        Reason:      reason,
        Metadata:    metadata,
    }

    if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
        // Log error but don't fail the operation
        s.logger.Error("failed to create audit log",
            zap.Error(err),
            zap.String("action", "ban_user"),
            zap.String("moderator_id", moderatorID.String()),
        )
        // Still return success since the main operation succeeded
    }

    return nil
}
```

### Structured Logging with Zap

Use structured logging for better debugging:

```go
import "go.uber.org/zap"

type ModerationService struct {
    // ... other fields ...
    logger *zap.Logger
}

func NewModerationService(...) *ModerationService {
    return &ModerationService{
        // ... other initialization ...
        logger: zap.L().Named("moderation_service"),
    }
}

func (s *ModerationService) BanUser(...) error {
    s.logger.Info("ban user initiated",
        zap.String("community_id", communityID.String()),
        zap.String("moderator_id", moderatorID.String()),
        zap.String("target_user_id", targetUserID.String()),
    )

    // Validation
    if err := s.validateModerationPermission(ctx, moderator, communityID); err != nil {
        s.logger.Warn("permission denied",
            zap.Error(err),
            zap.String("moderator_id", moderatorID.String()),
            zap.String("moderator_type", moderator.AccountType),
        )
        return err
    }

    // Execute ban
    if err := s.communityRepo.BanMember(ctx, ban); err != nil {
        s.logger.Error("failed to create ban",
            zap.Error(err),
            zap.String("community_id", communityID.String()),
            zap.String("target_user_id", targetUserID.String()),
        )
        return fmt.Errorf("failed to create ban: %w", err)
    }

    s.logger.Info("user banned successfully",
        zap.String("ban_id", ban.ID.String()),
        zap.String("community_id", communityID.String()),
        zap.String("target_user_id", targetUserID.String()),
    )

    return nil
}
```

### Log Levels

Use appropriate log levels:

```go
// DEBUG: Detailed information for debugging
s.logger.Debug("checking cache",
    zap.String("key", cacheKey),
)

// INFO: General informational messages
s.logger.Info("operation completed",
    zap.String("operation", "ban_user"),
    zap.Duration("duration", time.Since(start)),
)

// WARN: Warning messages for recoverable issues
s.logger.Warn("cache miss, querying database",
    zap.String("key", cacheKey),
)

// ERROR: Error messages for failures
s.logger.Error("database query failed",
    zap.Error(err),
    zap.String("query", "SELECT * FROM bans"),
)

// FATAL: Critical errors that require immediate attention
s.logger.Fatal("failed to connect to database",
    zap.Error(err),
)
```

### Querying Audit Logs

Create methods to query audit logs:

```go
// Get audit logs by moderator
func (r *AuditLogRepository) GetByModerator(
    ctx context.Context,
    moderatorID uuid.UUID,
    limit int,
) ([]*models.ModerationAuditLog, error) {
    query := `
        SELECT id, action, entity_type, entity_id, moderator_id, reason, metadata, created_at
        FROM moderation_audit_logs
        WHERE moderator_id = $1
        ORDER BY created_at DESC
        LIMIT $2
    `
    
    rows, err := r.db.Query(ctx, query, moderatorID, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var logs []*models.ModerationAuditLog
    for rows.Next() {
        var log models.ModerationAuditLog
        err := rows.Scan(
            &log.ID, &log.Action, &log.EntityType, &log.EntityID,
            &log.ModeratorID, &log.Reason, &log.Metadata, &log.CreatedAt,
        )
        if err != nil {
            return nil, err
        }
        logs = append(logs, &log)
    }

    return logs, nil
}

// Get audit logs by action
func (r *AuditLogRepository) GetByAction(
    ctx context.Context,
    action string,
    limit int,
) ([]*models.ModerationAuditLog, error) {
    query := `
        SELECT id, action, entity_type, entity_id, moderator_id, reason, metadata, created_at
        FROM moderation_audit_logs
        WHERE action = $1
        ORDER BY created_at DESC
        LIMIT $2
    `
    
    rows, err := r.db.Query(ctx, query, action, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var logs []*models.ModerationAuditLog
    for rows.Next() {
        var log models.ModerationAuditLog
        err := rows.Scan(
            &log.ID, &log.Action, &log.EntityType, &log.EntityID,
            &log.ModeratorID, &log.Reason, &log.Metadata, &log.CreatedAt,
        )
        if err != nil {
            return nil, err
        }
        logs = append(logs, &log)
    }

    return logs, nil
}

// Get audit logs for specific entity
func (r *AuditLogRepository) GetByEntity(
    ctx context.Context,
    entityType string,
    entityID uuid.UUID,
) ([]*models.ModerationAuditLog, error) {
    query := `
        SELECT id, action, entity_type, entity_id, moderator_id, reason, metadata, created_at
        FROM moderation_audit_logs
        WHERE entity_type = $1 AND entity_id = $2
        ORDER BY created_at DESC
    `
    
    rows, err := r.db.Query(ctx, query, entityType, entityID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var logs []*models.ModerationAuditLog
    for rows.Next() {
        var log models.ModerationAuditLog
        err := rows.Scan(
            &log.ID, &log.Action, &log.EntityType, &log.EntityID,
            &log.ModeratorID, &log.Reason, &log.Metadata, &log.CreatedAt,
        )
        if err != nil {
            return nil, err
        }
        logs = append(logs, &log)
    }

    return logs, nil
}
```

---

## Debugging Guide

### Common Issues

#### 1. Permission Denied Errors

**Symptom:** User gets 403 Forbidden when trying to perform moderation action

**Debugging Steps:**

```bash
# Check user's account type and permissions
psql -d clpr -c "SELECT id, username, account_type, moderator_scope FROM users WHERE id = 'user_id';"

# Check if user is in moderation_channels (for community moderators)
psql -d clpr -c "SELECT moderation_channels FROM users WHERE id = 'user_id';"

# Check community membership and role
psql -d clpr -c "SELECT role FROM community_members WHERE user_id = 'user_id' AND community_id = 'community_id';"
```

**Common Causes:**
- User has `account_type = 'member'` instead of `moderator` or `community_moderator`
- Community moderator trying to moderate a community not in their `moderation_channels`
- User has correct account type but wrong `moderator_scope`

**Fix:**
```sql
-- Upgrade user to site moderator
UPDATE users 
SET account_type = 'moderator', moderator_scope = 'site' 
WHERE id = 'user_id';

-- Add community to moderator's channels
UPDATE users 
SET moderation_channels = array_append(moderation_channels, 'community_id'::uuid)
WHERE id = 'user_id';
```

#### 2. Audit Logs Not Created

**Symptom:** Moderation action succeeds but no audit log appears

**Debugging Steps:**

```bash
# Check if audit logs exist for the moderator
psql -d clpr -c "SELECT * FROM moderation_audit_logs WHERE moderator_id = 'moderator_id' ORDER BY created_at DESC LIMIT 5;"

# Check application logs for errors
tail -f /var/log/clpr/app.log | grep "audit log"

# Check if audit log repository is properly initialized
# Look for initialization errors in startup logs
```

**Common Causes:**
- Audit log creation failing silently (service continues despite error)
- Database connection issues
- Missing audit_log_repo in service initialization

**Fix:**
```go
// Ensure audit log errors are logged
if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
    s.logger.Error("failed to create audit log",
        zap.Error(err),
        zap.String("action", auditLog.Action),
    )
    // Consider whether to fail the operation or continue
}
```

#### 3. Scope Validation Failures

**Symptom:** Community moderator cannot moderate their assigned community

**Debugging Steps:**

```bash
# Check moderator configuration
psql -d clpr -c "
    SELECT 
        u.id,
        u.username,
        u.account_type,
        u.moderator_scope,
        u.moderation_channels
    FROM users u
    WHERE u.id = 'moderator_id';
"

# Verify community exists
psql -d clpr -c "SELECT id, name FROM communities WHERE id = 'community_id';"

# Check community membership
psql -d clpr -c "
    SELECT role 
    FROM community_members 
    WHERE user_id = 'moderator_id' AND community_id = 'community_id';
"
```

**Common Causes:**
- Community ID not in `moderation_channels` array
- `moderator_scope` set to 'site' when it should be 'community'
- User not a member of the community with 'mod' or 'admin' role

**Fix:**
```sql
-- Set correct scope and add channel
UPDATE users 
SET 
    account_type = 'community_moderator',
    moderator_scope = 'community',
    moderation_channels = ARRAY['community_id'::uuid]
WHERE id = 'moderator_id';

-- Ensure user is community moderator
INSERT INTO community_members (community_id, user_id, role)
VALUES ('community_id'::uuid, 'moderator_id'::uuid, 'mod')
ON CONFLICT (community_id, user_id) DO UPDATE SET role = 'mod';
```

### Debugging Tools

#### 1. Check Permissions in Go

```go
// Test permission check in Go
user := &models.User{
    AccountType:        models.AccountTypeCommunityModerator,
    ModeratorScope:     models.ModeratorScopeCommunity,
    ModerationChannels: []uuid.UUID{communityID},
}

canModerate := user.Can(models.PermissionModerateUsers)
fmt.Printf("Can moderate users: %v\n", canModerate)
```

#### 2. SQL Debugging Queries

```sql
-- Get all moderators and their scopes
SELECT 
    id,
    username,
    account_type,
    moderator_scope,
    array_length(moderation_channels, 1) as channel_count
FROM users
WHERE account_type IN ('moderator', 'community_moderator', 'admin');

-- Get all bans with moderator info
SELECT 
    cb.id,
    cb.community_id,
    cb.banned_user_id,
    cb.reason,
    cb.banned_at,
    u.username as banned_by
FROM community_bans cb
LEFT JOIN users u ON cb.banned_by_user_id = u.id
ORDER BY cb.banned_at DESC
LIMIT 10;

-- Get audit log summary
SELECT 
    action,
    COUNT(*) as count,
    MAX(created_at) as last_action
FROM moderation_audit_logs
GROUP BY action
ORDER BY count DESC;
```

#### 3. cURL Testing

```bash
# Get JWT token
TOKEN=$(curl -X POST https://api.clpr.tv/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"mod@example.com","password":"password"}' \
  | jq -r '.access_token')

# Test ban endpoint
curl -X POST https://api.clpr.tv/api/v1/moderation/communities/{community_id}/ban \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "user_id",
    "reason": "Test ban"
  }' | jq

# Get bans list
curl https://api.clpr.tv/api/v1/moderation/communities/{community_id}/bans \
  -H "Authorization: Bearer $TOKEN" | jq
```

### Testing in Development

#### 1. Create Test Moderator

```sql
-- Create site moderator
INSERT INTO users (id, username, email, account_type, moderator_scope)
VALUES (
    gen_random_uuid(),
    'test_mod',
    'testmod@example.com',
    'moderator',
    'site'
);

-- Create community moderator for specific community
INSERT INTO users (id, username, email, account_type, moderator_scope, moderation_channels)
VALUES (
    gen_random_uuid(),
    'community_mod',
    'communitymod@example.com',
    'community_moderator',
    'community',
    ARRAY['community_id_here'::uuid]
);
```

#### 2. Seed Test Data

```sql
-- Create test community
INSERT INTO communities (id, name, slug, owner_id, description)
VALUES (
    gen_random_uuid(),
    'Test Community',
    'test-community',
    'owner_user_id'::uuid,
    'Community for testing'
);

-- Create test ban
INSERT INTO community_bans (community_id, banned_user_id, banned_by_user_id, reason)
VALUES (
    'community_id'::uuid,
    'target_user_id'::uuid,
    'moderator_id'::uuid,
    'Test ban reason'
);
```


---

## Quick Reference

### Checklist: Adding New Moderation Action

- [ ] Define service method in `moderation_service.go`
  - [ ] Add permission validation
  - [ ] Add scope validation
  - [ ] Create audit log entry
  - [ ] Handle errors with sentinel errors
- [ ] Create handler in `moderation_handler.go`
  - [ ] Parse and validate inputs
  - [ ] Call service method
  - [ ] Handle all sentinel errors
  - [ ] Return appropriate HTTP status codes
- [ ] Register route in router configuration
  - [ ] Apply authentication middleware
  - [ ] Apply authorization middleware
  - [ ] Set appropriate HTTP method (GET/POST/PUT/DELETE)
- [ ] Write tests
  - [ ] Unit tests for service
  - [ ] Unit tests for handler
  - [ ] Integration test (optional)
  - [ ] E2E test (optional)
- [ ] Update documentation
  - [ ] API documentation
  - [ ] OpenAPI/Swagger spec

### Checklist: Adding New Permission

- [ ] Add permission constant to `models/roles.go`
- [ ] Add to `accountTypePermissions` map for relevant account types
- [ ] Create permission check method (optional)
- [ ] Create middleware for permission check
- [ ] Apply middleware to protected routes
- [ ] Write permission tests
- [ ] Update frontend to check permission

### Permission Constants Reference

```go
// Member permissions
PermissionCreateSubmission
PermissionCreateComment
PermissionCreateVote
PermissionCreateFollow

// Broadcaster permissions
PermissionViewBroadcasterAnalytics
PermissionClaimBroadcasterProfile

// Moderator permissions
PermissionModerateContent
PermissionModerateUsers
PermissionCreateDiscoveryLists
PermissionManageUsers

// Community Moderator permissions
PermissionCommunityModerate
PermissionViewChannelAnalytics
PermissionManageModerators

// Admin permissions
PermissionManageSystem
PermissionViewAnalyticsDashboard
PermissionModerateOverride
```

### Account Types Reference

```go
AccountTypeMember             // Regular user
AccountTypeBroadcaster        // Content creator
AccountTypeModerator          // Site-wide moderator
AccountTypeCommunityModerator // Channel-scoped moderator
AccountTypeAdmin              // Full admin access
```

### Moderator Scopes

```go
ModeratorScopeSite      // Can moderate any community
ModeratorScopeCommunity // Can only moderate assigned communities
```

### Common Sentinel Errors

```go
ErrModerationPermissionDenied   // User lacks permission
ErrModerationNotAuthorized      // Not authorized for this community
ErrModerationCommunityNotFound  // Community doesn't exist
ErrModerationUserNotFound       // User doesn't exist
ErrModerationNotBanned          // User is not banned
ErrModerationCannotBanOwner     // Cannot ban community owner
```

### HTTP Status Codes

```
200 OK                  - Successful GET/PUT/PATCH
201 Created             - Successful POST (resource created)
204 No Content          - Successful DELETE
400 Bad Request         - Invalid input/validation error
401 Unauthorized        - Not authenticated (no/invalid token)
403 Forbidden           - Not authorized (lacks permission)
404 Not Found           - Resource doesn't exist
409 Conflict            - Resource already exists
429 Too Many Requests   - Rate limit exceeded
500 Internal Server     - Server error
```

### Database Tables

**moderation_audit_logs**
- Tracks all moderation actions
- Indexes: moderator_id, entity_type+entity_id, created_at, action

**community_bans**
- Stores user bans from communities
- Unique constraint: (community_id, banned_user_id)
- Indexes: community_id, banned_user_id

**moderation_queue**
- Queue for content requiring review
- Unique constraint: (content_type, content_id) WHERE status='pending'
- Indexes: status+priority, content_type+content_id, assigned_to

### Useful Commands

```bash
# Run backend tests
cd backend && go test ./...

# Run tests with coverage
cd backend && go test -coverprofile=coverage.out ./...

# View coverage report
cd backend && go tool cover -html=coverage.out

# Run integration tests
make test-integration

# Create new migration
cd backend && migrate create -ext sql -dir migrations -seq migration_name

# Apply migrations
make migrate-up

# Rollback last migration
make migrate-down

# Run frontend tests
cd frontend && npm test

# Run E2E tests
cd frontend && npm run test:e2e

# Run load tests
k6 run tests/load/moderation.js

# Format Go code
cd backend && go fmt ./...

# Lint Go code
cd backend && golangci-lint run

# Check for security issues
cd backend && gosec ./...
```

### Environment Variables

```bash
# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/clpr

# JWT
JWT_SECRET=your_jwt_secret_here
JWT_EXPIRATION=3600

# Server
PORT=8080
GIN_MODE=release

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Redis (caching)
REDIS_URL=redis://localhost:6379

# Twitch (for ban sync)
TWITCH_CLIENT_ID=your_client_id
TWITCH_CLIENT_SECRET=your_client_secret
```

### Code Snippets

#### Create Service Method Template

```go
func (s *ServiceName) MethodName(
    ctx context.Context,
    param1 Type1,
    param2 Type2,
) (ReturnType, error) {
    // 1. Get user
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    // 2. Validate permissions
    if !user.Can(PermissionName) {
        return nil, ErrPermissionDenied
    }

    // 3. Validate scope
    if err := s.validateScope(user, resourceID); err != nil {
        return nil, err
    }

    // 4. Perform operation
    result, err := s.repo.DoSomething(ctx, param1, param2)
    if err != nil {
        return nil, fmt.Errorf("operation failed: %w", err)
    }

    // 5. Create audit log
    auditLog := &models.AuditLog{
        Action:     "action_name",
        EntityType: "entity_type",
        EntityID:   result.ID,
        UserID:     userID,
        Metadata:   map[string]interface{}{"key": "value"},
    }
    if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
        s.logger.Error("failed to create audit log", zap.Error(err))
    }

    return result, nil
}
```

#### Create Handler Template

```go
func (h *HandlerName) MethodName(c *gin.Context) {
    ctx := c.Request.Context()

    // 1. Parse path params
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
        return
    }

    // 2. Get authenticated user
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    // 3. Parse request body
    var req RequestStruct
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid request",
            "details": err.Error(),
        })
        return
    }

    // 4. Call service
    result, err := h.service.MethodName(ctx, id, userID.(uuid.UUID), req)
    if err != nil {
        switch {
        case errors.Is(err, services.ErrPermissionDenied):
            c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
        case errors.Is(err, services.ErrNotFound):
            c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Operation failed"})
        }
        return
    }

    // 5. Return success
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    result,
    })
}
```

#### Create React Hook Template

```typescript
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';

export function useActionName(resourceId: string) {
    const queryClient = useQueryClient();
    const { toast } = useToast();

    return useMutation({
        mutationFn: async (data: RequestData) => {
            const response = await apiClient.post(`/endpoint/${resourceId}`, data);
            return response.data;
        },
        onSuccess: () => {
            queryClient.invalidateQueries(['query-key', resourceId]);
            toast({
                title: 'Success',
                description: 'Operation completed successfully',
                variant: 'success',
            });
        },
        onError: (error: any) => {
            toast({
                title: 'Error',
                description: error.response?.data?.error || 'Operation failed',
                variant: 'destructive',
            });
        },
    });
}
```

### Related Documentation

- **[Moderation API](./moderation-api.md)** - Complete API reference
- **[Permission Model](./permission-model.md)** - Permission system details
- **[Testing Guide](./testing.md)** - Testing best practices
- **[Architecture](./architecture.md)** - System architecture overview
- **[Database Schema](./database.md)** - Complete database schema
- **[Authentication](./authentication.md)** - Auth implementation
- **[Authorization Framework](./authorization-framework.md)** - Authorization details

### Getting Help

- **Internal Wiki**: https://wiki.clpr.internal/moderation
- **Team Chat**: #moderation-dev on Slack
- **Code Reviews**: Tag @moderation-team
- **Issues**: Label with `moderation` and `needs-review`

### Contributing

When implementing moderation features:

1. **Follow existing patterns** - Use this guide as a reference
2. **Write tests** - Aim for >80% coverage
3. **Document your code** - Add comments for complex logic
4. **Create audit logs** - All moderation actions must be logged
5. **Handle errors properly** - Use sentinel errors, return descriptive messages
6. **Update documentation** - Keep API docs and guides current
7. **Request code review** - Get approval from moderation team

---

## Summary

You now have a comprehensive understanding of the Clipper moderation system. Key takeaways:

1. **Architecture** - Clean layered design: Frontend → Handler → Service → Repository → Database
2. **Permissions** - Hierarchical system with account types and granular permissions
3. **Audit Trails** - Every moderation action is logged for compliance
4. **Testing** - Multiple layers: unit, integration, E2E, and load tests
5. **Patterns** - Consistent code patterns for services, handlers, and frontend

Start implementing moderation features by following the checklists and templates in this guide. When in doubt, look at existing code in `moderation_service.go` and `moderation_handler.go` as reference implementations.

**Happy coding! 🚀**

---

*Last Updated: 2025-02-03*  
*Version: 1.0*  
*Maintainer: Team Core*
