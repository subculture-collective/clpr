package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

func TestRequirePermission_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequirePermission(models.PermissionCreateSubmission))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestRequirePermission_HasPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Middleware to set user
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:          uuid.New(),
			Username:    "broadcaster",
			Role:        models.RoleUser,
			AccountType: models.AccountTypeBroadcaster,
		}
		c.Set("user", user)
		c.Next()
	})

	router.Use(RequirePermission(models.PermissionViewBroadcasterAnalytics))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequirePermission_LacksPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Middleware to set user (member cannot moderate)
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:          uuid.New(),
			Username:    "member",
			Role:        models.RoleUser,
			AccountType: models.AccountTypeMember,
		}
		c.Set("user", user)
		c.Next()
	})

	router.Use(RequirePermission(models.PermissionModerateContent))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRequirePermission_AdminHasAllPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Middleware to set admin user
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:          uuid.New(),
			Username:    "admin",
			Role:        models.RoleAdmin,
			AccountType: models.AccountTypeMember, // Even with member account type, admin role grants all
		}
		c.Set("user", user)
		c.Next()
	})

	router.Use(RequirePermission(models.PermissionManageSystem))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequireAnyPermission_HasOnePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Middleware to set broadcaster user
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:          uuid.New(),
			Username:    "broadcaster",
			Role:        models.RoleUser,
			AccountType: models.AccountTypeBroadcaster,
		}
		c.Set("user", user)
		c.Next()
	})

	router.Use(RequireAnyPermission(
		models.PermissionModerateContent,
		models.PermissionViewBroadcasterAnalytics,
	))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequireAnyPermission_LacksAllPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Middleware to set member user
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:          uuid.New(),
			Username:    "member",
			Role:        models.RoleUser,
			AccountType: models.AccountTypeMember,
		}
		c.Set("user", user)
		c.Next()
	})

	router.Use(RequireAnyPermission(
		models.PermissionModerateContent,
		models.PermissionViewBroadcasterAnalytics,
	))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRequireAccountType_HasAccountType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Middleware to set broadcaster user
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:          uuid.New(),
			Username:    "broadcaster",
			Role:        models.RoleUser,
			AccountType: models.AccountTypeBroadcaster,
		}
		c.Set("user", user)
		c.Next()
	})

	router.Use(RequireAccountType(models.AccountTypeBroadcaster, models.AccountTypeModerator))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequireAccountType_LacksAccountType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Middleware to set member user
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:          uuid.New(),
			Username:    "member",
			Role:        models.RoleUser,
			AccountType: models.AccountTypeMember,
		}
		c.Set("user", user)
		c.Next()
	})

	router.Use(RequireAccountType(models.AccountTypeBroadcaster, models.AccountTypeModerator))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRequireAccountType_DefaultsToMember(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Middleware to set user without account type set (should default to member)
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:          uuid.New(),
			Username:    "member",
			Role:        models.RoleUser,
			AccountType: "", // Empty, should default to member
		}
		c.Set("user", user)
		c.Next()
	})

	router.Use(RequireAccountType(models.AccountTypeMember))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestRequirePermission_CommunityModAccessOwnChannel tests that community moderators can access their own channels
func TestRequirePermission_CommunityModAccessOwnChannel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	channelID := uuid.New()

	// Middleware to set community moderator user
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:                 uuid.New(),
			Username:           "community_mod",
			Role:               models.RoleUser,
			AccountType:        models.AccountTypeCommunityModerator,
			ModeratorScope:     models.ModeratorScopeCommunity,
			ModerationChannels: []uuid.UUID{channelID},
		}
		c.Set("user", user)
		c.Next()
	})

	router.Use(RequirePermission(models.PermissionCommunityModerate))
	router.GET("/test/:channel_id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test/"+channelID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestRequirePermission_CommunityModDeniedOtherChannel tests that community moderators are denied access to channels they don't moderate
func TestRequirePermission_CommunityModDeniedOtherChannel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	ownChannelID := uuid.New()
	otherChannelID := uuid.New()

	// Middleware to set community moderator user
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:                 uuid.New(),
			Username:           "community_mod",
			Role:               models.RoleUser,
			AccountType:        models.AccountTypeCommunityModerator,
			ModeratorScope:     models.ModeratorScopeCommunity,
			ModerationChannels: []uuid.UUID{ownChannelID}, // Only has access to ownChannelID
		}
		c.Set("user", user)
		c.Next()
	})

	router.Use(RequirePermission(models.PermissionCommunityModerate))
	router.GET("/test/:channel_id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Try to access a different channel
	req, _ := http.NewRequest("GET", "/test/"+otherChannelID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

// TestRequirePermission_SiteModeratorAccessAnyChannel tests that site moderators can access any channel
func TestRequirePermission_SiteModeratorAccessAnyChannel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	channelID := uuid.New()

	// Middleware to set site moderator user
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:                 uuid.New(),
			Username:           "site_mod",
			Role:               models.RoleUser,
			AccountType:        models.AccountTypeModerator,
			ModeratorScope:     models.ModeratorScopeSite,
			ModerationChannels: []uuid.UUID{}, // Site mods have no channel restrictions
		}
		c.Set("user", user)
		c.Next()
	})

	router.Use(RequirePermission(models.PermissionModerateContent))
	router.GET("/test/:channel_id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test/"+channelID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestRequirePermission_AdminAccessAnyChannel tests that admins can access any channel
func TestRequirePermission_AdminAccessAnyChannel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	channelID := uuid.New()

	// Middleware to set admin user
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:          uuid.New(),
			Username:    "admin",
			Role:        models.RoleAdmin,
			AccountType: models.AccountTypeAdmin,
		}
		c.Set("user", user)
		c.Next()
	})

	router.Use(RequirePermission(models.PermissionManageSystem))
	router.GET("/test/:channel_id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test/"+channelID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestRequirePermission_CommunityModNoChannelInRequest tests that community moderators can access endpoints without channel_id
func TestRequirePermission_CommunityModNoChannelInRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	channelID := uuid.New()

	// Middleware to set community moderator user
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:                 uuid.New(),
			Username:           "community_mod",
			Role:               models.RoleUser,
			AccountType:        models.AccountTypeCommunityModerator,
			ModeratorScope:     models.ModeratorScopeCommunity,
			ModerationChannels: []uuid.UUID{channelID},
		}
		c.Set("user", user)
		c.Next()
	})

	router.Use(RequirePermission(models.PermissionCommunityModerate))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Request without channel_id - should pass (no scope check needed)
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestRequirePermission_CommunityModAccessOwnChannelViaQuery tests channel_id extraction from query parameter
func TestRequirePermission_CommunityModAccessOwnChannelViaQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	channelID := uuid.New()

	// Middleware to set community moderator user
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:                 uuid.New(),
			Username:           "community_mod",
			Role:               models.RoleUser,
			AccountType:        models.AccountTypeCommunityModerator,
			ModeratorScope:     models.ModeratorScopeCommunity,
			ModerationChannels: []uuid.UUID{channelID},
		}
		c.Set("user", user)
		c.Next()
	})

	router.Use(RequirePermission(models.PermissionCommunityModerate))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Request with channel_id as query parameter - should pass
	req, _ := http.NewRequest("GET", "/test?channel_id="+channelID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestRequirePermission_CommunityModDeniedOtherChannelViaQuery tests channel_id extraction from query parameter
func TestRequirePermission_CommunityModDeniedOtherChannelViaQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	ownChannelID := uuid.New()
	otherChannelID := uuid.New()

	// Middleware to set community moderator user
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:                 uuid.New(),
			Username:           "community_mod",
			Role:               models.RoleUser,
			AccountType:        models.AccountTypeCommunityModerator,
			ModeratorScope:     models.ModeratorScopeCommunity,
			ModerationChannels: []uuid.UUID{ownChannelID},
		}
		c.Set("user", user)
		c.Next()
	})

	router.Use(RequirePermission(models.PermissionCommunityModerate))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Request with different channel_id as query parameter - should be denied
	req, _ := http.NewRequest("GET", "/test?channel_id="+otherChannelID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

// TestRequirePermission_CommunityModAccessOwnChannelViaContext tests channel_id extraction from context
func TestRequirePermission_CommunityModAccessOwnChannelViaContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	channelID := uuid.New()

	// Middleware to set community moderator user and channel_id in context
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:                 uuid.New(),
			Username:           "community_mod",
			Role:               models.RoleUser,
			AccountType:        models.AccountTypeCommunityModerator,
			ModeratorScope:     models.ModeratorScopeCommunity,
			ModerationChannels: []uuid.UUID{channelID},
		}
		c.Set("user", user)
		c.Set("channel_id", channelID) // Set channel_id in context
		c.Next()
	})

	router.Use(RequirePermission(models.PermissionCommunityModerate))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Request without channel_id in path/query - should use context and pass
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestRequirePermission_CommunityModDeniedOtherChannelViaContext tests channel_id extraction from context
func TestRequirePermission_CommunityModDeniedOtherChannelViaContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	ownChannelID := uuid.New()
	otherChannelID := uuid.New()

	// Middleware to set community moderator user and different channel_id in context
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:                 uuid.New(),
			Username:           "community_mod",
			Role:               models.RoleUser,
			AccountType:        models.AccountTypeCommunityModerator,
			ModeratorScope:     models.ModeratorScopeCommunity,
			ModerationChannels: []uuid.UUID{ownChannelID},
		}
		c.Set("user", user)
		c.Set("channel_id", otherChannelID) // Set different channel_id in context
		c.Next()
	})

	router.Use(RequirePermission(models.PermissionCommunityModerate))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Request without channel_id in path/query - should use context and be denied
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}
