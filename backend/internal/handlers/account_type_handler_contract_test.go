package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type accountTypeServiceStub struct {
	info        *models.AccountTypeResponse
	stats       map[string]int
	conversions []models.AccountTypeConversion
	total       int
	err         error
}

func (s *accountTypeServiceStub) GetUserAccountType(context.Context, uuid.UUID) (*models.AccountTypeResponse, error) {
	return s.info, s.err
}
func (s *accountTypeServiceStub) ConvertToBroadcaster(context.Context, uuid.UUID, *string, bool) error {
	return s.err
}
func (s *accountTypeServiceStub) ConvertToModerator(context.Context, uuid.UUID, uuid.UUID, *string) error {
	return s.err
}
func (s *accountTypeServiceStub) GetConversionHistory(context.Context, uuid.UUID, int, int) ([]models.AccountTypeConversion, int, error) {
	return s.conversions, s.total, s.err
}
func (s *accountTypeServiceStub) GetRecentConversions(context.Context, int, int) ([]models.AccountTypeConversion, int, error) {
	return s.conversions, s.total, s.err
}
func (s *accountTypeServiceStub) GetAccountTypeStats(context.Context) (map[string]int, error) {
	return s.stats, s.err
}

func TestAccountTypeAdminResponseContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	conversionID := uuid.New()
	service := &accountTypeServiceStub{
		info: &models.AccountTypeResponse{AccountType: models.AccountTypeModerator, Permissions: []string{"moderate_content"}},
		stats: map[string]int{
			models.AccountTypeMember: 10, models.AccountTypeBroadcaster: 3,
			models.AccountTypeModerator: 2, models.AccountTypeAdmin: 1,
		},
		conversions: []models.AccountTypeConversion{{
			ID: conversionID, UserID: userID, OldType: models.AccountTypeMember,
			NewType: models.AccountTypeModerator, ConvertedAt: time.Now(), CreatedAt: time.Now(),
		}},
		total: 1,
	}
	handler := NewAccountTypeHandler(service, nil)

	t.Run("stats", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/account-types/stats", nil)
		handler.GetAccountTypeStats(ctx)
		require.Equal(t, http.StatusOK, recorder.Code)
		var body struct {
			Success bool           `json:"success"`
			Data    map[string]int `json:"data"`
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
		require.True(t, body.Success)
		require.Equal(t, service.stats, body.Data)
	})

	t.Run("recent conversions", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/account-types/conversions?limit=25&offset=5", nil)
		handler.GetRecentConversions(ctx)
		require.Equal(t, http.StatusOK, recorder.Code)
		var body struct {
			Success bool `json:"success"`
			Data    struct {
				Conversions []models.AccountTypeConversion `json:"conversions"`
				Total       int                            `json:"total"`
				Limit       int                            `json:"limit"`
				Offset      int                            `json:"offset"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
		require.True(t, body.Success)
		require.Len(t, body.Data.Conversions, 1)
		require.Equal(t, conversionID, body.Data.Conversions[0].ID)
		require.Equal(t, 25, body.Data.Limit)
		require.Equal(t, 5, body.Data.Offset)
	})

	t.Run("convert to moderator", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Params = gin.Params{{Key: "id", Value: userID.String()}}
		ctx.Set("user", &models.User{ID: uuid.New(), Role: "admin"})
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/account-types/users/"+userID.String()+"/convert-to-moderator", http.NoBody)
		ctx.Request.Header.Set("Content-Type", "application/json")
		handler.ConvertToModerator(ctx)
		require.Equal(t, http.StatusBadRequest, recorder.Code, "an empty required JSON body must be rejected")

		recorder = httptest.NewRecorder()
		ctx, _ = gin.CreateTestContext(recorder)
		ctx.Params = gin.Params{{Key: "id", Value: userID.String()}}
		ctx.Set("user", &models.User{ID: uuid.New(), Role: "admin"})
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/account-types/users/"+userID.String()+"/convert-to-moderator", strings.NewReader(`{}`))
		ctx.Request.Header.Set("Content-Type", "application/json")
		handler.ConvertToModerator(ctx)
		require.Equal(t, http.StatusOK, recorder.Code)
		var body struct {
			Success bool                        `json:"success"`
			Message string                      `json:"message"`
			Data    *models.AccountTypeResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
		require.True(t, body.Success)
		require.Equal(t, models.AccountTypeModerator, body.Data.AccountType)
	})
}
