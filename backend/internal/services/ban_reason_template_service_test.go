package services

import (
	"context"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCanManageBanTemplates(t *testing.T) {
	channel := "12345"
	otherChannel := "67890"
	empty := ""

	broadcaster := &models.User{ID: uuid.New(), Role: models.RoleUser, TwitchID: &channel}
	viewer := &models.User{ID: uuid.New(), Role: models.RoleUser, TwitchID: &otherChannel}
	noTwitch := &models.User{ID: uuid.New(), Role: models.RoleUser}
	platformMod := &models.User{ID: uuid.New(), Role: models.RoleModerator}
	admin := &models.User{ID: uuid.New(), Role: models.RoleAdmin}

	assert.True(t, canManageBanTemplates(broadcaster, &channel))
	assert.False(t, canManageBanTemplates(viewer, &channel))
	assert.False(t, canManageBanTemplates(noTwitch, &channel))
	assert.False(t, canManageBanTemplates(nil, &channel))

	// Templates outside any channel are for platform staff only.
	assert.False(t, canManageBanTemplates(broadcaster, nil))
	assert.False(t, canManageBanTemplates(broadcaster, &empty))

	for _, staff := range []*models.User{platformMod, admin} {
		assert.True(t, canManageBanTemplates(staff, &channel))
		assert.True(t, canManageBanTemplates(staff, nil))
	}
}

// The checks run before any repository call, so a nil repository proves the
// request was refused without reading or writing templates.
func TestBanReasonTemplateService_RefusesOtherChannels(t *testing.T) {
	ctx := context.Background()
	channel := "12345"
	otherChannel := "67890"
	viewer := &models.User{ID: uuid.New(), Role: models.RoleUser, TwitchID: &otherChannel}
	service := &BanReasonTemplateService{}

	_, err := service.CreateTemplate(ctx, viewer, &models.CreateBanReasonTemplateRequest{
		Name: "Spam", Reason: "Spam", BroadcasterID: &channel,
	})
	assert.ErrorIs(t, err, ErrUnauthorizedTemplate)

	_, err = service.ListTemplates(ctx, viewer, &channel, true)
	assert.ErrorIs(t, err, ErrUnauthorizedTemplate)

	_, err = service.GetUsageStats(ctx, viewer, &channel)
	assert.ErrorIs(t, err, ErrUnauthorizedTemplate)
}
