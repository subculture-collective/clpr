package repository

import "errors"

var (
	// ErrMaxPresetsReached is returned when a user tries to create more than 10 presets
	ErrMaxPresetsReached = errors.New("maximum of 10 presets allowed per user")
	// ErrPresetNotFound is returned when a preset is not found
	ErrPresetNotFound = errors.New("preset not found")
	// ErrUnauthorizedPresetAccess is returned when a user tries to access another user's preset
	ErrUnauthorizedPresetAccess     = errors.New("unauthorized access to preset")
	ErrVerificationAlreadyVerified  = errors.New("user is already verified")
	ErrVerificationPending          = errors.New("verification application is already pending")
	ErrVerificationCooldown         = errors.New("verification application cooldown is active")
	ErrVerificationLimit            = errors.New("verification application limit reached")
	ErrVerificationDuplicateChannel = errors.New("Twitch channel is already associated with another user")
	ErrVerificationNotFound         = errors.New("verification application not found")
	ErrVerificationAlreadyReviewed  = errors.New("verification application has already been reviewed")
)
