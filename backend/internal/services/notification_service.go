package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

// NotificationService handles notification business logic
type NotificationService struct {
	repo         *repository.NotificationRepository
	userRepo     *repository.UserRepository
	commentRepo  *repository.CommentRepository
	clipRepo     *repository.ClipRepository
	favoriteRepo *repository.FavoriteRepository
	emailService *EmailService
}

// NewNotificationService creates a new NotificationService
func NewNotificationService(
	repo *repository.NotificationRepository,
	userRepo *repository.UserRepository,
	commentRepo *repository.CommentRepository,
	clipRepo *repository.ClipRepository,
	favoriteRepo *repository.FavoriteRepository,
	emailService *EmailService,
) *NotificationService {
	return &NotificationService{
		repo:         repo,
		userRepo:     userRepo,
		commentRepo:  commentRepo,
		clipRepo:     clipRepo,
		favoriteRepo: favoriteRepo,
		emailService: emailService,
	}
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(
	ctx context.Context,
	userID uuid.UUID,
	notificationType string,
	title string,
	message string,
	link *string,
	sourceUserID *uuid.UUID,
	sourceContentID *uuid.UUID,
	sourceContentType *string,
) (*models.Notification, error) {
	// Check user's notification preferences
	prefs, err := s.repo.GetPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification preferences: %w", err)
	}

	// Check if user has notifications enabled and specific type is enabled
	if !prefs.InAppEnabled {
		return nil, nil // User has disabled notifications
	}

	// Check type-specific preferences
	if !s.shouldNotify(prefs, notificationType) {
		return nil, nil // User has disabled this type of notification
	}

	notification := &models.Notification{
		ID:                uuid.New(),
		UserID:            userID,
		Type:              notificationType,
		Title:             title,
		Message:           message,
		Link:              link,
		IsRead:            false,
		CreatedAt:         time.Now(),
		SourceUserID:      sourceUserID,
		SourceContentID:   sourceContentID,
		SourceContentType: sourceContentType,
	}

	err = s.repo.Create(ctx, notification)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return notification, nil
}

// CreateNotificationWithEmail creates a notification and optionally sends an email
func (s *NotificationService) CreateNotificationWithEmail(
	ctx context.Context,
	userID uuid.UUID,
	notificationType string,
	title string,
	message string,
	link *string,
	sourceUserID *uuid.UUID,
	sourceContentID *uuid.UUID,
	sourceContentType *string,
	emailData map[string]interface{},
) (*models.Notification, error) {
	// Create in-app notification
	notification, err := s.CreateNotification(
		ctx, userID, notificationType, title, message, link,
		sourceUserID, sourceContentID, sourceContentType,
	)
	if err != nil {
		return nil, err
	}

	// If notification was created, try to send email
	if notification != nil {
		// Check if user has email notifications enabled
		prefs, err := s.repo.GetPreferences(ctx, userID)
		if err == nil && prefs.EmailEnabled && s.shouldNotify(prefs, notificationType) {
			// Get user info for email
			user, err := s.userRepo.GetByID(ctx, userID)
			if err == nil && s.emailService != nil {
				// Send email asynchronously with proper tracking
				s.emailService.SendNotificationEmailAsync(
					ctx, user, notificationType, notification.ID, emailData,
				)
			}
		}
	}

	return notification, nil
}

// shouldNotify checks if a user should be notified based on their preferences
func (s *NotificationService) shouldNotify(prefs *models.NotificationPreferences, notificationType string) bool {
	switch notificationType {
	// Account & Security
	case models.NotificationTypeLoginNewDevice:
		return prefs.NotifyLoginNewDevice
	case models.NotificationTypeFailedLogin:
		return prefs.NotifyFailedLogin
	case models.NotificationTypePasswordChanged:
		return prefs.NotifyPasswordChanged
	case models.NotificationTypeEmailChanged:
		return prefs.NotifyEmailChanged

	// Content notifications
	case models.NotificationTypeReply:
		return prefs.NotifyReplies
	case models.NotificationTypeMention:
		return prefs.NotifyMentions
	case models.NotificationTypeVoteMilestone:
		return prefs.NotifyVotes
	case models.NotificationTypeFavoritedClipComment:
		return prefs.NotifyFavoritedClipComment
	case models.NotificationTypeContentTrending:
		return prefs.NotifyContentTrending
	case models.NotificationTypeContentFlagged:
		return prefs.NotifyContentFlagged

	// Community notifications
	case models.NotificationTypeModeratorMessage:
		return prefs.NotifyModeratorMessage
	case models.NotificationTypeUserFollowed:
		return prefs.NotifyUserFollowed
	case models.NotificationTypeCommentOnContent:
		return prefs.NotifyCommentOnContent
	case models.NotificationTypeDiscussionReply:
		return prefs.NotifyDiscussionReply
	case models.NotificationTypeBadgeEarned:
		return prefs.NotifyBadges
	case models.NotificationTypeRankUp:
		return prefs.NotifyRankUp
	case models.NotificationTypeContentRemoved, models.NotificationTypeWarning,
		models.NotificationTypeBan, models.NotificationTypeAppealDecision:
		return prefs.NotifyModeration

	// Creator-specific notification preferences (including clip submissions)
	case models.NotificationTypeSubmissionApproved:
		return prefs.NotifyClipApproved
	case models.NotificationTypeSubmissionRejected:
		return prefs.NotifyClipRejected
	case models.NotificationTypeClipComment:
		return prefs.NotifyClipComments
	case models.NotificationTypeClipViewThreshold, models.NotificationTypeClipVoteThreshold:
		return prefs.NotifyClipThreshold

	// Broadcaster notifications
	case models.NotificationTypeBroadcasterLive:
		return prefs.NotifyBroadcasterLive

	// Global/Marketing
	case models.NotificationTypeMarketing:
		return prefs.NotifyMarketing
	case models.NotificationTypePolicyUpdate:
		return prefs.NotifyPolicyUpdates
	case models.NotificationTypePlatformAnnouncement:
		return prefs.NotifyPlatformAnnouncements

	default:
		return true // Default to notifying for unknown types
	}
}

// GetUserNotifications retrieves notifications for a user
func (s *NotificationService) GetUserNotifications(
	ctx context.Context,
	userID uuid.UUID,
	filter string,
	limit, offset int,
) ([]models.NotificationWithSource, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	notifications, err := s.repo.ListByUserID(ctx, userID, filter, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user notifications: %w", err)
	}

	return notifications, nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := s.repo.CountUnread(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) error {
	err := s.repo.MarkAsRead(ctx, notificationID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	err := s.repo.MarkAllAsRead(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) error {
	err := s.repo.Delete(ctx, notificationID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	return nil
}

// GetPreferences retrieves notification preferences for a user
func (s *NotificationService) GetPreferences(ctx context.Context, userID uuid.UUID) (*models.NotificationPreferences, error) {
	prefs, err := s.repo.GetPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get preferences: %w", err)
	}

	return prefs, nil
}

// UpdatePreferences updates notification preferences for a user
func (s *NotificationService) UpdatePreferences(ctx context.Context, prefs *models.NotificationPreferences) error {
	err := s.repo.UpdatePreferences(ctx, prefs)
	if err != nil {
		return fmt.Errorf("failed to update preferences: %w", err)
	}

	return nil
}

// ResetPreferences resets notification preferences to defaults for a user
func (s *NotificationService) ResetPreferences(ctx context.Context, userID uuid.UUID) (*models.NotificationPreferences, error) {
	prefs, err := s.repo.ResetPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to reset preferences: %w", err)
	}

	return prefs, nil
}

// NotifyCommentReply notifies a user when someone replies to their comment
func (s *NotificationService) NotifyCommentReply(
	ctx context.Context,
	clipID uuid.UUID,
	parentCommentID uuid.UUID,
	replyAuthorID uuid.UUID,
) error {
	// Get parent comment to find the author
	parentComment, err := s.commentRepo.GetByID(ctx, parentCommentID, nil)
	if err != nil {
		return fmt.Errorf("failed to get parent comment: %w", err)
	}

	// Don't notify if replying to own comment
	if parentComment.UserID == replyAuthorID {
		return nil
	}

	// Get reply author info
	replyAuthor, err := s.userRepo.GetByID(ctx, replyAuthorID)
	if err != nil {
		return fmt.Errorf("failed to get reply author: %w", err)
	}

	// Get clip info
	clip, err := s.clipRepo.GetByID(ctx, clipID)
	if err != nil {
		return fmt.Errorf("failed to get clip: %w", err)
	}

	title := fmt.Sprintf("%s replied to your comment", replyAuthor.DisplayName)
	message := fmt.Sprintf("on \"%s\"", clip.Title)
	link := fmt.Sprintf("/clips/%s", clipID.String())

	// Prepare email data
	commentPreview := parentComment.Content
	if len(commentPreview) > 100 {
		commentPreview = commentPreview[:100] + "..."
	}
	emailData := map[string]interface{}{
		"AuthorName":     replyAuthor.DisplayName,
		"ClipTitle":      clip.Title,
		"ClipURL":        fmt.Sprintf("%s/clips/%s", s.getBaseURL(), clipID.String()),
		"CommentPreview": commentPreview,
	}

	contentType := "comment"
	_, err = s.CreateNotificationWithEmail(
		ctx,
		parentComment.UserID,
		models.NotificationTypeReply,
		title,
		message,
		&link,
		&replyAuthorID,
		&clipID,
		&contentType,
		emailData,
	)

	return err
}

// getBaseURL returns the base URL for the application
func (s *NotificationService) getBaseURL() string {
	if s.emailService != nil {
		return s.emailService.baseURL
	}
	return "http://localhost:5173" // fallback
}

// NotifyMentions notifies users mentioned in a comment
func (s *NotificationService) NotifyMentions(
	ctx context.Context,
	content string,
	clipID uuid.UUID,
	commentAuthorID uuid.UUID,
) error {
	// Extract mentions from content
	mentions := extractMentions(content)
	if len(mentions) == 0 {
		return nil
	}

	// Get comment author info
	author, err := s.userRepo.GetByID(ctx, commentAuthorID)
	if err != nil {
		return fmt.Errorf("failed to get comment author: %w", err)
	}

	// Get clip info
	clip, err := s.clipRepo.GetByID(ctx, clipID)
	if err != nil {
		return fmt.Errorf("failed to get clip: %w", err)
	}

	// Prepare comment preview for email
	commentPreview := content
	if len(commentPreview) > 100 {
		commentPreview = commentPreview[:100] + "..."
	}

	// Notify each mentioned user
	for _, username := range mentions {
		// Get user by username
		user, err := s.userRepo.GetByUsername(ctx, username)
		if err != nil {
			continue // Skip if user not found
		}

		// Don't notify if mentioning yourself
		if user.ID == commentAuthorID {
			continue
		}

		title := fmt.Sprintf("%s mentioned you in a comment", author.DisplayName)
		message := fmt.Sprintf("on \"%s\"", clip.Title)
		link := fmt.Sprintf("/clips/%s", clipID.String())

		// Prepare email data
		emailData := map[string]interface{}{
			"AuthorName":     author.DisplayName,
			"ClipTitle":      clip.Title,
			"ClipURL":        fmt.Sprintf("%s/clips/%s", s.getBaseURL(), clipID.String()),
			"CommentPreview": commentPreview,
		}

		contentType := "comment"
		_, err = s.CreateNotificationWithEmail(
			ctx,
			user.ID,
			models.NotificationTypeMention,
			title,
			message,
			&link,
			&commentAuthorID,
			&clipID,
			&contentType,
			emailData,
		)
		if err != nil {
			// Log error but continue with other mentions
			continue
		}
	}

	return nil
}

// NotifyVoteMilestone notifies a user when their comment reaches a vote milestone
func (s *NotificationService) NotifyVoteMilestone(
	ctx context.Context,
	commentID uuid.UUID,
	voteScore int,
) error {
	// Only notify for specific milestones
	milestones := []int{10, 25, 50, 100, 250, 500, 1000}
	isMilestone := false
	for _, m := range milestones {
		if voteScore == m {
			isMilestone = true
			break
		}
	}

	if !isMilestone {
		return nil
	}

	// Get comment to find the author
	comment, err := s.commentRepo.GetByID(ctx, commentID, nil)
	if err != nil {
		return fmt.Errorf("failed to get comment: %w", err)
	}

	// Get clip info
	clip, err := s.clipRepo.GetByID(ctx, comment.ClipID)
	if err != nil {
		return fmt.Errorf("failed to get clip: %w", err)
	}

	title := fmt.Sprintf("Your comment received %d upvotes!", voteScore)
	message := fmt.Sprintf("on \"%s\"", clip.Title)
	link := fmt.Sprintf("/clips/%s", comment.ClipID.String())

	contentType := "comment"
	_, err = s.CreateNotification(
		ctx,
		comment.UserID,
		models.NotificationTypeVoteMilestone,
		title,
		message,
		&link,
		nil,
		&commentID,
		&contentType,
	)

	return err
}

// NotifyBadgeEarned notifies a user when they earn a badge
func (s *NotificationService) NotifyBadgeEarned(
	ctx context.Context,
	userID uuid.UUID,
	badgeName string,
) error {
	title := fmt.Sprintf("You earned the %s badge!", badgeName)
	message := "Check out your profile to see your new badge"
	link := "/profile"

	_, err := s.CreateNotification(
		ctx,
		userID,
		models.NotificationTypeBadgeEarned,
		title,
		message,
		&link,
		nil,
		nil,
		nil,
	)

	return err
}

// NotifyRankUp notifies a user when they rank up
func (s *NotificationService) NotifyRankUp(
	ctx context.Context,
	userID uuid.UUID,
	newRank string,
) error {
	title := fmt.Sprintf("You ranked up to %s!", newRank)
	message := "Keep up the great work!"
	link := "/profile"

	_, err := s.CreateNotification(
		ctx,
		userID,
		models.NotificationTypeRankUp,
		title,
		message,
		&link,
		nil,
		nil,
		nil,
	)

	return err
}

// NotifyFavoritedClipComment notifies users when a clip they favorited gets a new comment
func (s *NotificationService) NotifyFavoritedClipComment(
	ctx context.Context,
	clipID uuid.UUID,
	commentAuthorID uuid.UUID,
) error {
	// Get clip info
	clip, err := s.clipRepo.GetByID(ctx, clipID)
	if err != nil {
		return fmt.Errorf("failed to get clip: %w", err)
	}

	// Get all users who favorited this clip
	favorites, err := s.favoriteRepo.GetByClipID(ctx, clipID)
	if err != nil {
		return fmt.Errorf("failed to get favorites: %w", err)
	}

	// Get comment author info
	author, err := s.userRepo.GetByID(ctx, commentAuthorID)
	if err != nil {
		return fmt.Errorf("failed to get comment author: %w", err)
	}

	// Notify each user who favorited the clip (except the comment author)
	for _, favorite := range favorites {
		if favorite.UserID == commentAuthorID {
			continue // Don't notify the comment author
		}

		title := fmt.Sprintf("%s commented on a clip you favorited", author.DisplayName)
		message := fmt.Sprintf("\"%s\"", clip.Title)
		link := fmt.Sprintf("/clips/%s", clipID.String())

		contentType := "clip"
		_, err = s.CreateNotification(
			ctx,
			favorite.UserID,
			models.NotificationTypeFavoritedClipComment,
			title,
			message,
			&link,
			&commentAuthorID,
			&clipID,
			&contentType,
		)
		if err != nil {
			// Log error but continue with other users
			continue
		}
	}

	return nil
}

// NotifySubmissionApproved notifies a user when their submission is approved
func (s *NotificationService) NotifySubmissionApproved(
	ctx context.Context,
	submitterID uuid.UUID,
	submissionID uuid.UUID,
	clipTitle string,
) error {
	title := "Your clip submission was approved!"
	message := fmt.Sprintf("\"%s\" is now live on clpr", clipTitle)
	link := "/submissions"

	contentType := "submission"
	_, err := s.CreateNotification(
		ctx,
		submitterID,
		models.NotificationTypeSubmissionApproved,
		title,
		message,
		&link,
		nil,
		&submissionID,
		&contentType,
	)

	return err
}

// NotifySubmissionRejected notifies a user when their submission is rejected
func (s *NotificationService) NotifySubmissionRejected(
	ctx context.Context,
	submitterID uuid.UUID,
	submissionID uuid.UUID,
	clipTitle string,
	reason string,
) error {
	title := "Your clip submission was not approved"
	message := fmt.Sprintf("\"%s\" - Reason: %s", clipTitle, reason)
	link := "/submissions"

	contentType := "submission"
	_, err := s.CreateNotification(
		ctx,
		submitterID,
		models.NotificationTypeSubmissionRejected,
		title,
		message,
		&link,
		nil,
		&submissionID,
		&contentType,
	)

	return err
}

// extractMentions extracts @username mentions from text
func extractMentions(text string) []string {
	// Match @username pattern (alphanumeric and underscore)
	re := regexp.MustCompile(`@([a-zA-Z0-9_]+)`)
	matches := re.FindAllStringSubmatch(text, -1)

	var mentions []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			username := strings.ToLower(match[1])
			if !seen[username] {
				mentions = append(mentions, username)
				seen[username] = true
			}
		}
	}

	return mentions
}

// RegisterDeviceToken registers a device token for push notifications
func (s *NotificationService) RegisterDeviceToken(
	ctx context.Context,
	userID uuid.UUID,
	deviceToken string,
	devicePlatform string,
) error {
	// Update user's device token in the database
	err := s.userRepo.UpdateDeviceToken(ctx, userID, deviceToken, devicePlatform)
	if err != nil {
		return fmt.Errorf("failed to register device token for user %s: %w", userID.String(), err)
	}

	return nil
}

// UnregisterDeviceToken removes a device token
func (s *NotificationService) UnregisterDeviceToken(
	ctx context.Context,
	userID uuid.UUID,
	deviceToken string,
) error {
	// Clear user's device token from the database
	err := s.userRepo.ClearDeviceToken(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to unregister device token for user %s: %w", userID.String(), err)
	}

	return nil
}

// NotifyClipComment notifies a clip creator when someone comments on their clip
func (s *NotificationService) NotifyClipComment(
	ctx context.Context,
	clipID uuid.UUID,
	commentAuthorID uuid.UUID,
	clipCreatorID string,
) error {
	// Get clip creator user by Twitch ID
	clipCreator, err := s.userRepo.GetByTwitchID(ctx, clipCreatorID)
	if err != nil {
		// Creator might not be a registered user, silently skip
		return nil
	}

	// Don't notify if creator comments on their own clip
	if clipCreator.ID == commentAuthorID {
		return nil
	}

	// Get comment author info
	author, err := s.userRepo.GetByID(ctx, commentAuthorID)
	if err != nil {
		return fmt.Errorf("failed to get comment author: %w", err)
	}

	// Get clip info
	clip, err := s.clipRepo.GetByID(ctx, clipID)
	if err != nil {
		return fmt.Errorf("failed to get clip: %w", err)
	}

	title := fmt.Sprintf("%s commented on your clip", author.DisplayName)
	message := fmt.Sprintf("\"%s\"", clip.Title)
	link := fmt.Sprintf("/clips/%s", clipID.String())

	// Prepare email data
	emailData := map[string]interface{}{
		"AuthorName": author.DisplayName,
		"ClipTitle":  clip.Title,
		"ClipURL":    fmt.Sprintf("%s/clips/%s", s.getBaseURL(), clipID.String()),
	}

	contentType := "clip"
	_, err = s.CreateNotificationWithEmail(
		ctx,
		clipCreator.ID,
		models.NotificationTypeClipComment,
		title,
		message,
		&link,
		&commentAuthorID,
		&clipID,
		&contentType,
		emailData,
	)

	return err
}

// NotifyClipViewThreshold notifies a clip creator when their clip reaches view milestones
func (s *NotificationService) NotifyClipViewThreshold(
	ctx context.Context,
	clipID uuid.UUID,
	viewCount int64,
	clipCreatorID string,
) error {
	// Only notify for specific milestones
	milestones := []int64{100, 500, 1000, 5000, 10000, 50000, 100000}
	isMilestone := false
	for _, m := range milestones {
		if viewCount == m {
			isMilestone = true
			break
		}
	}

	if !isMilestone {
		return nil
	}

	// Get clip creator user by Twitch ID
	clipCreator, err := s.userRepo.GetByTwitchID(ctx, clipCreatorID)
	if err != nil {
		// Creator might not be a registered user, silently skip
		return nil
	}

	// Get clip info
	clip, err := s.clipRepo.GetByID(ctx, clipID)
	if err != nil {
		return fmt.Errorf("failed to get clip: %w", err)
	}

	title := fmt.Sprintf("Your clip reached %s views!", formatNumber(viewCount))
	message := fmt.Sprintf("\"%s\" is trending!", clip.Title)
	link := fmt.Sprintf("/clips/%s", clipID.String())

	contentType := "clip"
	_, err = s.CreateNotification(
		ctx,
		clipCreator.ID,
		models.NotificationTypeClipViewThreshold,
		title,
		message,
		&link,
		nil,
		&clipID,
		&contentType,
	)

	return err
}

// NotifyClipVoteThreshold notifies a clip creator when their clip reaches vote milestones
func (s *NotificationService) NotifyClipVoteThreshold(
	ctx context.Context,
	clipID uuid.UUID,
	voteScore int,
	clipCreatorID string,
) error {
	// Only notify for specific milestones
	milestones := []int{10, 25, 50, 100, 250, 500, 1000}
	isMilestone := false
	for _, m := range milestones {
		if voteScore == m {
			isMilestone = true
			break
		}
	}

	if !isMilestone {
		return nil
	}

	// Get clip creator user by Twitch ID
	clipCreator, err := s.userRepo.GetByTwitchID(ctx, clipCreatorID)
	if err != nil {
		// Creator might not be a registered user, silently skip
		return nil
	}

	// Get clip info
	clip, err := s.clipRepo.GetByID(ctx, clipID)
	if err != nil {
		return fmt.Errorf("failed to get clip: %w", err)
	}

	title := fmt.Sprintf("Your clip reached %d upvotes!", voteScore)
	message := fmt.Sprintf("\"%s\" is popular!", clip.Title)
	link := fmt.Sprintf("/clips/%s", clipID.String())

	contentType := "clip"
	_, err = s.CreateNotification(
		ctx,
		clipCreator.ID,
		models.NotificationTypeClipVoteThreshold,
		title,
		message,
		&link,
		nil,
		&clipID,
		&contentType,
	)

	return err
}

// formatNumber formats a number with commas for readability
func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	return fmt.Sprintf("%d,%03d,%03d", n/1000000, (n%1000000)/1000, n%1000)
}
