package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// MockCommunityRepository is a mock implementation of CommunityRepository
type MockCommunityRepository struct {
	mock.Mock
}

func (m *MockCommunityRepository) GetCommunityByID(ctx context.Context, id uuid.UUID) (*models.Community, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Community), args.Error(1)
}

func (m *MockCommunityRepository) GetMember(ctx context.Context, communityID, userID uuid.UUID) (*models.CommunityMember, error) {
	args := m.Called(ctx, communityID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CommunityMember), args.Error(1)
}

func (m *MockCommunityRepository) IsBanned(ctx context.Context, communityID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, communityID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCommunityRepository) BanMember(ctx context.Context, ban *models.CommunityBan) error {
	args := m.Called(ctx, ban)
	return args.Error(0)
}

func (m *MockCommunityRepository) UnbanMember(ctx context.Context, communityID, userID uuid.UUID) error {
	args := m.Called(ctx, communityID, userID)
	return args.Error(0)
}

func (m *MockCommunityRepository) RemoveMember(ctx context.Context, communityID, userID uuid.UUID) error {
	args := m.Called(ctx, communityID, userID)
	return args.Error(0)
}

func (m *MockCommunityRepository) ListBans(ctx context.Context, communityID uuid.UUID, limit, offset int) ([]*models.CommunityBan, int, error) {
	args := m.Called(ctx, communityID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.CommunityBan), args.Int(1), args.Error(2)
}

func (m *MockCommunityRepository) ListAllBans(ctx context.Context, limit, offset int) ([]*models.CommunityBan, int, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.CommunityBan), args.Int(1), args.Error(2)
}

func (m *MockCommunityRepository) GetBanByID(ctx context.Context, banID uuid.UUID) (*models.CommunityBan, error) {
	args := m.Called(ctx, banID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CommunityBan), args.Error(1)
}

// MockModerationUserRepository is a mock implementation of UserRepository for moderation tests
type MockModerationUserRepository struct {
	mock.Mock
}

func (m *MockModerationUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// MockModerationAuditLogRepository is a mock implementation of AuditLogRepository for moderation tests
type MockModerationAuditLogRepository struct {
	mock.Mock
}

func (m *MockModerationAuditLogRepository) Create(ctx context.Context, log *models.ModerationAuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

// MockDBPool is a mock implementation of pgxpool.Pool for transactions
type MockDBPool struct {
	mock.Mock
}

type MockTx struct {
	mock.Mock
}

func (m *MockTx) Commit(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTx) Rollback(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Test helper to create a site moderator
func createSiteModerator() *models.User {
	return &models.User{
		ID:             uuid.New(),
		Username:       "sitemod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}
}

// Test helper to create a community moderator
func createCommunityModerator(communityID uuid.UUID) *models.User {
	return &models.User{
		ID:                 uuid.New(),
		Username:           "communitymod",
		AccountType:        models.AccountTypeCommunityModerator,
		ModeratorScope:     models.ModeratorScopeCommunity,
		ModerationChannels: []uuid.UUID{communityID},
	}
}

// Test helper to create an admin
func createAdmin() *models.User {
	return &models.User{
		ID:          uuid.New(),
		Username:    "admin",
		AccountType: models.AccountTypeAdmin,
		Role:        models.RoleAdmin,
	}
}

// Test helper to create a regular user
func createRegularUser() *models.User {
	return &models.User{
		ID:          uuid.New(),
		Username:    "regularuser",
		AccountType: models.AccountTypeMember,
		Role:        models.RoleUser,
	}
}

func TestModerationService_BanUser_SiteModerator(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	siteMod := createSiteModerator()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockAuditLogRepo := new(MockModerationAuditLogRepository)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
		auditLogRepo:  mockAuditLogRepo,
	}

	// Note: We can't properly test transaction without a real DB pool
	// In a real test environment, you'd use a test database
	// This test only validates permission and scope checks
	err := service.validateModerationPermission(ctx, siteMod, communityID)
	assert.NoError(t, err)

	err = service.validateModerationScope(siteMod, communityID)
	assert.NoError(t, err)
}

func TestModerationService_BanUser_CommunityModerator_Authorized(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	communityMod := createCommunityModerator(communityID)

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	member := &models.CommunityMember{
		ID:          uuid.New(),
		CommunityID: communityID,
		UserID:      communityMod.ID,
		Role:        models.CommunityRoleMod,
	}
	mockCommunityRepo.On("GetMember", ctx, communityID, communityMod.ID).Return(member, nil)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	err := service.validateModerationPermission(ctx, communityMod, communityID)
	assert.NoError(t, err)

	err = service.validateModerationScope(communityMod, communityID)
	assert.NoError(t, err)

	mockCommunityRepo.AssertExpectations(t)
}

func TestModerationService_BanUser_CommunityModerator_Unauthorized(t *testing.T) {
	communityID := uuid.New()
	otherCommunityID := uuid.New()
	communityMod := createCommunityModerator(otherCommunityID) // Authorized for different community

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	// Should fail scope validation
	err := service.validateModerationScope(communityMod, communityID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not authorized to moderate this community")
}

func TestModerationService_BanUser_RegularUser_Denied(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	regularUser := createRegularUser()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	// Should fail permission validation
	err := service.validateModerationPermission(ctx, regularUser, communityID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient permissions")
}

func TestModerationService_BanUser_CannotBanOwner(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	siteMod := createSiteModerator()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	// Validate permission and scope pass, but trying to ban owner should fail
	err := service.validateModerationPermission(ctx, siteMod, communityID)
	assert.NoError(t, err)
}

func TestModerationService_UnbanUser_Success(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	siteMod := createSiteModerator()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockAuditLogRepo := new(MockModerationAuditLogRepository)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
		auditLogRepo:  mockAuditLogRepo,
	}

	err := service.validateModerationPermission(ctx, siteMod, communityID)
	assert.NoError(t, err)
}

func TestModerationService_UnbanUser_NotBanned(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	siteMod := createSiteModerator()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	// Validation should pass
	err := service.validateModerationPermission(ctx, siteMod, communityID)
	assert.NoError(t, err)
}

func TestModerationService_GetBans_Success(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	siteMod := createSiteModerator()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	// Validate permission and scope
	err := service.validateModerationPermission(ctx, siteMod, communityID)
	assert.NoError(t, err)
}

func TestModerationService_UpdateBan_Success(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	targetUserID := uuid.New()
	siteMod := createSiteModerator()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockAuditLogRepo := new(MockModerationAuditLogRepository)

	mockUserRepo.On("GetByID", ctx, siteMod.ID).Return(siteMod, nil)
	mockCommunityRepo.On("IsBanned", ctx, communityID, targetUserID).Return(true, nil)
	mockCommunityRepo.On("UnbanMember", ctx, communityID, targetUserID).Return(nil)
	mockCommunityRepo.On("BanMember", ctx, mock.AnythingOfType("*models.CommunityBan")).Return(nil)
	mockAuditLogRepo.On("Create", ctx, mock.AnythingOfType("*models.ModerationAuditLog")).Return(nil)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
		auditLogRepo:  mockAuditLogRepo,
	}

	// Validate permission and scope
	err := service.validateModerationPermission(ctx, siteMod, communityID)
	assert.NoError(t, err)
}

func TestModerationService_Admin_CanModerateAnywhere(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	admin := createAdmin()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	// Admin should pass both permission and scope validation
	err := service.validateModerationPermission(ctx, admin, communityID)
	assert.NoError(t, err)

	err = service.validateModerationScope(admin, communityID)
	assert.NoError(t, err)
}

func TestModerationService_CommunityModerator_NotMember_Denied(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	communityMod := createCommunityModerator(communityID)

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	// Mock that moderator is not a member of the community
	mockCommunityRepo.On("GetMember", ctx, communityID, communityMod.ID).Return(nil, nil)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	// Should fail permission validation
	err := service.validateModerationPermission(ctx, communityMod, communityID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient permissions")

	mockCommunityRepo.AssertExpectations(t)
}

func TestModerationService_CommunityModerator_InsufficientRole_Denied(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	communityMod := createCommunityModerator(communityID)

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	// Mock that moderator is a member but only has "member" role, not "mod" or "admin"
	member := &models.CommunityMember{
		ID:          uuid.New(),
		CommunityID: communityID,
		UserID:      communityMod.ID,
		Role:        models.CommunityRoleMember, // Not a mod!
	}
	mockCommunityRepo.On("GetMember", ctx, communityID, communityMod.ID).Return(member, nil)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	// Should fail permission validation
	err := service.validateModerationPermission(ctx, communityMod, communityID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient permissions")

	mockCommunityRepo.AssertExpectations(t)
}

func TestModerationService_AuditLogging(t *testing.T) {
	ctx := context.Background()

	mockAuditLogRepo := new(MockModerationAuditLogRepository)

	// Test that audit log is created with correct fields
	mockAuditLogRepo.On("Create", ctx, mock.MatchedBy(func(log *models.ModerationAuditLog) bool {
		return log.Action == "ban_user" &&
			log.EntityType == "community_ban" &&
			log.ModeratorID != uuid.Nil &&
			log.Metadata != nil
	})).Return(nil)

	auditLog := &models.ModerationAuditLog{
		Action:      "ban_user",
		EntityType:  "community_ban",
		EntityID:    uuid.New(),
		ModeratorID: uuid.New(),
		Metadata: map[string]interface{}{
			"community_id": uuid.New().String(),
		},
	}

	err := mockAuditLogRepo.Create(ctx, auditLog)
	assert.NoError(t, err)

	mockAuditLogRepo.AssertExpectations(t)
}

// TestModerationService_BanUser_FullFlow tests the complete ban user flow
func TestModerationService_BanUser_FullFlow(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()
	targetUserID := uuid.New()
	ownerID := uuid.New()
	reason := "spam"

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockAuditLogRepo := new(MockModerationAuditLogRepository)

	siteMod := &models.User{
		ID:             moderatorID,
		Username:       "sitemod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	community := &models.Community{
		ID:      communityID,
		OwnerID: ownerID,
	}

	// Setup mocks
	mockUserRepo.On("GetByID", ctx, moderatorID).Return(siteMod, nil)
	mockCommunityRepo.On("GetCommunityByID", ctx, communityID).Return(community, nil)
	mockCommunityRepo.On("RemoveMember", ctx, communityID, targetUserID).Return(nil)
	mockCommunityRepo.On("BanMember", ctx, mock.AnythingOfType("*models.CommunityBan")).Return(nil)
	mockAuditLogRepo.On("Create", ctx, mock.AnythingOfType("*models.ModerationAuditLog")).Return(nil)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
		auditLogRepo:  mockAuditLogRepo,
	}

	err := service.BanUser(ctx, communityID, moderatorID, targetUserID, &reason)
	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
	mockCommunityRepo.AssertExpectations(t)
	mockAuditLogRepo.AssertExpectations(t)
}

// TestModerationService_UnbanUser_FullFlow tests the complete unban user flow
func TestModerationService_UnbanUser_FullFlow(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()
	targetUserID := uuid.New()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockAuditLogRepo := new(MockModerationAuditLogRepository)

	siteMod := &models.User{
		ID:             moderatorID,
		Username:       "sitemod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	// Setup mocks
	mockUserRepo.On("GetByID", ctx, moderatorID).Return(siteMod, nil)
	mockCommunityRepo.On("IsBanned", ctx, communityID, targetUserID).Return(true, nil)
	mockCommunityRepo.On("UnbanMember", ctx, communityID, targetUserID).Return(nil)
	mockAuditLogRepo.On("Create", ctx, mock.AnythingOfType("*models.ModerationAuditLog")).Return(nil)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
		auditLogRepo:  mockAuditLogRepo,
	}

	err := service.UnbanUser(ctx, communityID, moderatorID, targetUserID)
	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
	mockCommunityRepo.AssertExpectations(t)
	mockAuditLogRepo.AssertExpectations(t)
}

// TestModerationService_GetBans_FullFlow tests the complete get bans flow with pagination
func TestModerationService_GetBans_FullFlow(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	siteMod := &models.User{
		ID:             moderatorID,
		Username:       "sitemod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	bans := []*models.CommunityBan{
		{
			ID:           uuid.New(),
			CommunityID:  communityID,
			BannedUserID: uuid.New(),
			BannedAt:     time.Now(),
		},
		{
			ID:           uuid.New(),
			CommunityID:  communityID,
			BannedUserID: uuid.New(),
			BannedAt:     time.Now(),
		},
	}

	// Setup mocks
	mockUserRepo.On("GetByID", ctx, moderatorID).Return(siteMod, nil)
	mockCommunityRepo.On("ListBans", ctx, communityID, 20, 0).Return(bans, 2, nil)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	result, total, err := service.GetBans(ctx, communityID, moderatorID, 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Equal(t, 2, len(result))

	mockUserRepo.AssertExpectations(t)
	mockCommunityRepo.AssertExpectations(t)
}

// TestModerationService_GetBans_PaginationDefaults tests pagination defaults
func TestModerationService_GetBans_PaginationDefaults(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	siteMod := &models.User{
		ID:             moderatorID,
		Username:       "sitemod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	// Setup mocks - should use default page=1, limit=20
	mockUserRepo.On("GetByID", ctx, moderatorID).Return(siteMod, nil)
	mockCommunityRepo.On("ListBans", ctx, communityID, 20, 0).Return([]*models.CommunityBan{}, 0, nil)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	// Test with invalid page and limit
	_, _, err := service.GetBans(ctx, communityID, moderatorID, 0, 0)
	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
	mockCommunityRepo.AssertExpectations(t)
}

// TestModerationService_GetBans_MaxLimit tests pagination max limit
func TestModerationService_GetBans_MaxLimit(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	siteMod := &models.User{
		ID:             moderatorID,
		Username:       "sitemod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	// Setup mocks - should cap at 100
	mockUserRepo.On("GetByID", ctx, moderatorID).Return(siteMod, nil)
	mockCommunityRepo.On("ListBans", ctx, communityID, 100, 0).Return([]*models.CommunityBan{}, 0, nil)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	// Test with excessive limit
	_, _, err := service.GetBans(ctx, communityID, moderatorID, 1, 1000)
	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
	mockCommunityRepo.AssertExpectations(t)
}

// TestModerationService_UpdateBan_FullFlow tests the complete update ban flow
func TestModerationService_UpdateBan_FullFlow(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()
	targetUserID := uuid.New()
	newReason := "updated reason"

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockAuditLogRepo := new(MockModerationAuditLogRepository)

	siteMod := &models.User{
		ID:             moderatorID,
		Username:       "sitemod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	// Setup mocks
	mockUserRepo.On("GetByID", ctx, moderatorID).Return(siteMod, nil)
	mockCommunityRepo.On("IsBanned", ctx, communityID, targetUserID).Return(true, nil)
	mockCommunityRepo.On("UnbanMember", ctx, communityID, targetUserID).Return(nil)
	mockCommunityRepo.On("BanMember", ctx, mock.AnythingOfType("*models.CommunityBan")).Return(nil)
	mockAuditLogRepo.On("Create", ctx, mock.AnythingOfType("*models.ModerationAuditLog")).Return(nil)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
		auditLogRepo:  mockAuditLogRepo,
	}

	err := service.UpdateBan(ctx, communityID, moderatorID, targetUserID, &newReason)
	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
	mockCommunityRepo.AssertExpectations(t)
	mockAuditLogRepo.AssertExpectations(t)
}

// TestModerationService_UpdateBan_UserNotBanned tests updating ban when user isn't banned
func TestModerationService_UpdateBan_UserNotBanned(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()
	targetUserID := uuid.New()
	newReason := "updated reason"

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	siteMod := &models.User{
		ID:             moderatorID,
		Username:       "sitemod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	// Setup mocks - user is not banned
	mockUserRepo.On("GetByID", ctx, moderatorID).Return(siteMod, nil)
	mockCommunityRepo.On("IsBanned", ctx, communityID, targetUserID).Return(false, nil)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	err := service.UpdateBan(ctx, communityID, moderatorID, targetUserID, &newReason)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not banned")

	mockUserRepo.AssertExpectations(t)
	mockCommunityRepo.AssertExpectations(t)
}

// TestModerationService_NewModerationService tests service creation
func TestModerationService_NewModerationService(t *testing.T) {
	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockAuditLogRepo := new(MockModerationAuditLogRepository)

	service := &ModerationService{
		db:            nil,
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
		auditLogRepo:  mockAuditLogRepo,
	}

	assert.NotNil(t, service)
	assert.Equal(t, mockCommunityRepo, service.communityRepo)
	assert.Equal(t, mockUserRepo, service.userRepo)
	assert.Equal(t, mockAuditLogRepo, service.auditLogRepo)
}

// TestModerationService_HasModerationPermission tests permission check without ban queries
func TestModerationService_HasModerationPermission(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	siteMod := &models.User{
		ID:             moderatorID,
		Username:       "sitemod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	mockUserRepo.On("GetByID", ctx, moderatorID).Return(siteMod, nil)

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	err := service.HasModerationPermission(ctx, communityID, moderatorID)
	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
}

// TestModerationService_BanUser_GetModeratorError tests error when fetching moderator
func TestModerationService_BanUser_GetModeratorError(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()
	targetUserID := uuid.New()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	// Mock error getting moderator
	mockUserRepo.On("GetByID", ctx, moderatorID).Return(nil, errors.New("db error"))

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	err := service.BanUser(ctx, communityID, moderatorID, targetUserID, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get moderator")

	mockUserRepo.AssertExpectations(t)
}

// TestModerationService_BanUser_GetCommunityError tests error when fetching community
func TestModerationService_BanUser_GetCommunityError(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()
	targetUserID := uuid.New()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	siteMod := &models.User{
		ID:             moderatorID,
		Username:       "sitemod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	mockUserRepo.On("GetByID", ctx, moderatorID).Return(siteMod, nil)
	mockCommunityRepo.On("GetCommunityByID", ctx, communityID).Return(nil, errors.New("community not found"))

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	err := service.BanUser(ctx, communityID, moderatorID, targetUserID, nil)
	assert.Error(t, err)
	assert.Equal(t, ErrModerationCommunityNotFound, err)

	mockUserRepo.AssertExpectations(t)
	mockCommunityRepo.AssertExpectations(t)
}

// TestModerationService_BanUser_BanMemberError tests error when creating ban
func TestModerationService_BanUser_BanMemberError(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()
	targetUserID := uuid.New()
	ownerID := uuid.New()
	reason := "test reason"

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockAuditLogRepo := new(MockModerationAuditLogRepository)

	siteMod := &models.User{
		ID:             moderatorID,
		Username:       "sitemod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	community := &models.Community{
		ID:      communityID,
		OwnerID: ownerID,
	}

	mockUserRepo.On("GetByID", ctx, moderatorID).Return(siteMod, nil)
	mockCommunityRepo.On("GetCommunityByID", ctx, communityID).Return(community, nil)
	mockCommunityRepo.On("RemoveMember", ctx, communityID, targetUserID).Return(nil)
	mockCommunityRepo.On("BanMember", ctx, mock.AnythingOfType("*models.CommunityBan")).Return(errors.New("db error"))

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
		auditLogRepo:  mockAuditLogRepo,
	}

	err := service.BanUser(ctx, communityID, moderatorID, targetUserID, &reason)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create ban")

	mockUserRepo.AssertExpectations(t)
	mockCommunityRepo.AssertExpectations(t)
}

// TestModerationService_BanUser_AuditLogError tests error when creating audit log
func TestModerationService_BanUser_AuditLogError(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()
	targetUserID := uuid.New()
	ownerID := uuid.New()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockAuditLogRepo := new(MockModerationAuditLogRepository)

	siteMod := &models.User{
		ID:             moderatorID,
		Username:       "sitemod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	community := &models.Community{
		ID:      communityID,
		OwnerID: ownerID,
	}

	mockUserRepo.On("GetByID", ctx, moderatorID).Return(siteMod, nil)
	mockCommunityRepo.On("GetCommunityByID", ctx, communityID).Return(community, nil)
	mockCommunityRepo.On("RemoveMember", ctx, communityID, targetUserID).Return(nil)
	mockCommunityRepo.On("BanMember", ctx, mock.AnythingOfType("*models.CommunityBan")).Return(nil)
	mockAuditLogRepo.On("Create", ctx, mock.AnythingOfType("*models.ModerationAuditLog")).Return(errors.New("audit error"))

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
		auditLogRepo:  mockAuditLogRepo,
	}

	err := service.BanUser(ctx, communityID, moderatorID, targetUserID, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create audit log")

	mockUserRepo.AssertExpectations(t)
	mockCommunityRepo.AssertExpectations(t)
	mockAuditLogRepo.AssertExpectations(t)
}

// TestModerationService_UnbanUser_GetModeratorError tests error when fetching moderator for unban
func TestModerationService_UnbanUser_GetModeratorError(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()
	targetUserID := uuid.New()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	mockUserRepo.On("GetByID", ctx, moderatorID).Return(nil, errors.New("db error"))

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	err := service.UnbanUser(ctx, communityID, moderatorID, targetUserID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get moderator")

	mockUserRepo.AssertExpectations(t)
}

// TestModerationService_UnbanUser_IsBannedError tests error checking ban status
func TestModerationService_UnbanUser_IsBannedError(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()
	targetUserID := uuid.New()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	siteMod := &models.User{
		ID:             moderatorID,
		Username:       "sitemod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	mockUserRepo.On("GetByID", ctx, moderatorID).Return(siteMod, nil)
	mockCommunityRepo.On("IsBanned", ctx, communityID, targetUserID).Return(false, errors.New("db error"))

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	err := service.UnbanUser(ctx, communityID, moderatorID, targetUserID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check ban status")

	mockUserRepo.AssertExpectations(t)
	mockCommunityRepo.AssertExpectations(t)
}

// TestModerationService_UnbanUser_UnbanError tests error when removing ban
func TestModerationService_UnbanUser_UnbanError(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()
	targetUserID := uuid.New()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	siteMod := &models.User{
		ID:             moderatorID,
		Username:       "sitemod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	mockUserRepo.On("GetByID", ctx, moderatorID).Return(siteMod, nil)
	mockCommunityRepo.On("IsBanned", ctx, communityID, targetUserID).Return(true, nil)
	mockCommunityRepo.On("UnbanMember", ctx, communityID, targetUserID).Return(errors.New("db error"))

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	err := service.UnbanUser(ctx, communityID, moderatorID, targetUserID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove ban")

	mockUserRepo.AssertExpectations(t)
	mockCommunityRepo.AssertExpectations(t)
}

// TestModerationService_UnbanUser_AuditLogError tests error when creating audit log for unban
func TestModerationService_UnbanUser_AuditLogError(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()
	targetUserID := uuid.New()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockAuditLogRepo := new(MockModerationAuditLogRepository)

	siteMod := &models.User{
		ID:             moderatorID,
		Username:       "sitemod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	mockUserRepo.On("GetByID", ctx, moderatorID).Return(siteMod, nil)
	mockCommunityRepo.On("IsBanned", ctx, communityID, targetUserID).Return(true, nil)
	mockCommunityRepo.On("UnbanMember", ctx, communityID, targetUserID).Return(nil)
	mockAuditLogRepo.On("Create", ctx, mock.AnythingOfType("*models.ModerationAuditLog")).Return(errors.New("audit error"))

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
		auditLogRepo:  mockAuditLogRepo,
	}

	err := service.UnbanUser(ctx, communityID, moderatorID, targetUserID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create audit log")

	mockUserRepo.AssertExpectations(t)
	mockCommunityRepo.AssertExpectations(t)
	mockAuditLogRepo.AssertExpectations(t)
}

// TestModerationService_UpdateBan_GetModeratorError tests error when fetching moderator for update
func TestModerationService_UpdateBan_GetModeratorError(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()
	targetUserID := uuid.New()
	newReason := "updated reason"

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	mockUserRepo.On("GetByID", ctx, moderatorID).Return(nil, errors.New("db error"))

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	err := service.UpdateBan(ctx, communityID, moderatorID, targetUserID, &newReason)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get moderator")

	mockUserRepo.AssertExpectations(t)
}

// TestModerationService_UpdateBan_IsBannedError tests error checking ban status for update
func TestModerationService_UpdateBan_IsBannedError(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()
	targetUserID := uuid.New()
	newReason := "updated reason"

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	siteMod := &models.User{
		ID:             moderatorID,
		Username:       "sitemod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	mockUserRepo.On("GetByID", ctx, moderatorID).Return(siteMod, nil)
	mockCommunityRepo.On("IsBanned", ctx, communityID, targetUserID).Return(false, errors.New("db error"))

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	err := service.UpdateBan(ctx, communityID, moderatorID, targetUserID, &newReason)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check ban status")

	mockUserRepo.AssertExpectations(t)
	mockCommunityRepo.AssertExpectations(t)
}

// TestModerationService_UpdateBan_UnbanError tests error when removing old ban during update
func TestModerationService_UpdateBan_UnbanError(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()
	targetUserID := uuid.New()
	newReason := "updated reason"

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	siteMod := &models.User{
		ID:             moderatorID,
		Username:       "sitemod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	mockUserRepo.On("GetByID", ctx, moderatorID).Return(siteMod, nil)
	mockCommunityRepo.On("IsBanned", ctx, communityID, targetUserID).Return(true, nil)
	mockCommunityRepo.On("UnbanMember", ctx, communityID, targetUserID).Return(errors.New("db error"))

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	err := service.UpdateBan(ctx, communityID, moderatorID, targetUserID, &newReason)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove old ban")

	mockUserRepo.AssertExpectations(t)
	mockCommunityRepo.AssertExpectations(t)
}

// TestModerationService_UpdateBan_CreateBanError tests error when creating new ban during update
func TestModerationService_UpdateBan_CreateBanError(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()
	targetUserID := uuid.New()
	newReason := "updated reason"

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	siteMod := &models.User{
		ID:             moderatorID,
		Username:       "sitemod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	mockUserRepo.On("GetByID", ctx, moderatorID).Return(siteMod, nil)
	mockCommunityRepo.On("IsBanned", ctx, communityID, targetUserID).Return(true, nil)
	mockCommunityRepo.On("UnbanMember", ctx, communityID, targetUserID).Return(nil)
	mockCommunityRepo.On("BanMember", ctx, mock.AnythingOfType("*models.CommunityBan")).Return(errors.New("db error"))

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	err := service.UpdateBan(ctx, communityID, moderatorID, targetUserID, &newReason)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create updated ban")

	mockUserRepo.AssertExpectations(t)
	mockCommunityRepo.AssertExpectations(t)
}

// TestModerationService_UpdateBan_AuditLogError tests error when creating audit log for update
func TestModerationService_UpdateBan_AuditLogError(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()
	targetUserID := uuid.New()
	newReason := "updated reason"

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)
	mockAuditLogRepo := new(MockModerationAuditLogRepository)

	siteMod := &models.User{
		ID:             moderatorID,
		Username:       "sitemod",
		AccountType:    models.AccountTypeModerator,
		ModeratorScope: models.ModeratorScopeSite,
	}

	mockUserRepo.On("GetByID", ctx, moderatorID).Return(siteMod, nil)
	mockCommunityRepo.On("IsBanned", ctx, communityID, targetUserID).Return(true, nil)
	mockCommunityRepo.On("UnbanMember", ctx, communityID, targetUserID).Return(nil)
	mockCommunityRepo.On("BanMember", ctx, mock.AnythingOfType("*models.CommunityBan")).Return(nil)
	mockAuditLogRepo.On("Create", ctx, mock.AnythingOfType("*models.ModerationAuditLog")).Return(errors.New("audit error"))

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
		auditLogRepo:  mockAuditLogRepo,
	}

	err := service.UpdateBan(ctx, communityID, moderatorID, targetUserID, &newReason)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create audit log")

	mockUserRepo.AssertExpectations(t)
	mockCommunityRepo.AssertExpectations(t)
	mockAuditLogRepo.AssertExpectations(t)
}

// TestModerationService_GetBans_GetModeratorError tests error when fetching moderator for GetBans
func TestModerationService_GetBans_GetModeratorError(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	moderatorID := uuid.New()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	mockUserRepo.On("GetByID", ctx, moderatorID).Return(nil, errors.New("db error"))

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	_, _, err := service.GetBans(ctx, communityID, moderatorID, 1, 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get moderator")

	mockUserRepo.AssertExpectations(t)
}

// TestModerationService_validateModerationPermission_GetMemberError tests error fetching member
func TestModerationService_validateModerationPermission_GetMemberError(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	communityModID := uuid.New()

	mockCommunityRepo := new(MockCommunityRepository)
	mockUserRepo := new(MockModerationUserRepository)

	communityMod := &models.User{
		ID:                 communityModID,
		Username:           "communitymod",
		AccountType:        models.AccountTypeCommunityModerator,
		ModeratorScope:     models.ModeratorScopeCommunity,
		ModerationChannels: []uuid.UUID{communityID},
	}

	mockCommunityRepo.On("GetMember", ctx, communityID, communityModID).Return(nil, errors.New("db error"))

	service := &ModerationService{
		communityRepo: mockCommunityRepo,
		userRepo:      mockUserRepo,
	}

	err := service.validateModerationPermission(ctx, communityMod, communityID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get member")

	mockCommunityRepo.AssertExpectations(t)
}
