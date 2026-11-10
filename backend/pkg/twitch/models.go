package twitch

import "time"

// ClipParams contains parameters for fetching clips
type ClipParams struct {
	BroadcasterID string
	GameID        string
	ClipIDs       []string
	StartedAt     time.Time
	EndedAt       time.Time
	First         int    // Max 100
	After         string // Pagination cursor
	Before        string // Pagination cursor
}

// Clip represents a Twitch clip from the API
type Clip struct {
	ID              string    `json:"id"`
	URL             string    `json:"url"`
	EmbedURL        string    `json:"embed_url"`
	BroadcasterID   string    `json:"broadcaster_id"`
	BroadcasterName string    `json:"broadcaster_name"`
	CreatorID       string    `json:"creator_id"`
	CreatorName     string    `json:"creator_name"`
	VideoID         string    `json:"video_id"`
	GameID          string    `json:"game_id"`
	Language        string    `json:"language"`
	Title           string    `json:"title"`
	ViewCount       int       `json:"view_count"`
	CreatedAt       time.Time `json:"created_at"`
	ThumbnailURL    string    `json:"thumbnail_url"`
	Duration        float64   `json:"duration"`
	VodOffset       *int      `json:"vod_offset"`
}

// VOD returns the clip's past broadcast and offset as stored values: nil
// when Twitch reports none.
func (c *Clip) VOD() (*string, *int) {
	if c.VideoID == "" {
		return nil, nil
	}
	videoID := c.VideoID
	return &videoID, c.VodOffset
}

// ClipsResponse represents the response from the clips endpoint
type ClipsResponse struct {
	Data       []Clip     `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// User represents a Twitch user from the API
type User struct {
	ID              string    `json:"id"`
	Login           string    `json:"login"`
	DisplayName     string    `json:"display_name"`
	Type            string    `json:"type"`
	BroadcasterType string    `json:"broadcaster_type"`
	Description     string    `json:"description"`
	ProfileImageURL string    `json:"profile_image_url"`
	OfflineImageURL string    `json:"offline_image_url"`
	ViewCount       int       `json:"view_count"`
	Email           string    `json:"email,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// UsersResponse represents the response from the users endpoint
type UsersResponse struct {
	Data []User `json:"data"`
}

// Game represents a Twitch game/category from the API
type Game struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BoxArtURL string `json:"box_art_url"`
	IGDBID    string `json:"igdb_id"`
}

// GamesResponse represents the response from the games endpoint
type GamesResponse struct {
	Data []Game `json:"data"`
}

// TopGamesResponse represents the response from the top games endpoint
type TopGamesResponse struct {
	Data       []Game     `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Pagination contains cursor information for paginated responses
type Pagination struct {
	Cursor string `json:"cursor"`
	Total  int    `json:"total,omitempty"`
}

// Stream represents a Twitch stream from the API
type Stream struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	UserLogin    string    `json:"user_login"`
	UserName     string    `json:"user_name"`
	GameID       string    `json:"game_id"`
	GameName     string    `json:"game_name"`
	Type         string    `json:"type"` // "live" or ""
	Title        string    `json:"title"`
	ViewerCount  int       `json:"viewer_count"`
	StartedAt    time.Time `json:"started_at"`
	Language     string    `json:"language"`
	ThumbnailURL string    `json:"thumbnail_url"`
	TagIDs       []string  `json:"tag_ids"`
	IsMature     bool      `json:"is_mature"`
}

// StreamsResponse represents the response from the streams endpoint
type StreamsResponse struct {
	Data       []Stream   `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Channel represents channel information from the API
type Channel struct {
	BroadcasterID       string   `json:"broadcaster_id"`
	BroadcasterLogin    string   `json:"broadcaster_login"`
	BroadcasterName     string   `json:"broadcaster_name"`
	BroadcasterLanguage string   `json:"broadcaster_language"`
	GameID              string   `json:"game_id"`
	GameName            string   `json:"game_name"`
	Title               string   `json:"title"`
	Delay               int      `json:"delay"`
	Tags                []string `json:"tags"`
	ContentLabels       []string `json:"content_classification_labels"`
	IsBrandedContent    bool     `json:"is_branded_content"`
}

// ChannelsResponse represents the response from the channels endpoint
type ChannelsResponse struct {
	Data []Channel `json:"data"`
}

// Video represents a video from the API
type Video struct {
	ID            string    `json:"id"`
	StreamID      string    `json:"stream_id"`
	UserID        string    `json:"user_id"`
	UserLogin     string    `json:"user_login"`
	UserName      string    `json:"user_name"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
	PublishedAt   time.Time `json:"published_at"`
	URL           string    `json:"url"`
	ThumbnailURL  string    `json:"thumbnail_url"`
	Viewable      string    `json:"viewable"`
	ViewCount     int       `json:"view_count"`
	Language      string    `json:"language"`
	Type          string    `json:"type"`
	Duration      string    `json:"duration"`
	MutedSegments []struct {
		Duration int `json:"duration"`
		Offset   int `json:"offset"`
	} `json:"muted_segments"`
}

// VideosResponse represents the response from the videos endpoint
type VideosResponse struct {
	Data       []Video    `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Follower represents a follower relationship from the API
type Follower struct {
	UserID     string    `json:"user_id"`
	UserLogin  string    `json:"user_login"`
	UserName   string    `json:"user_name"`
	FollowedAt time.Time `json:"followed_at"`
}

// FollowersResponse represents the response from the followers endpoint
type FollowersResponse struct {
	Data       []Follower `json:"data"`
	Total      int        `json:"total"`
	Pagination Pagination `json:"pagination"`
}

// BannedUser represents a banned user from the moderation/banned endpoint
type BannedUser struct {
	UserID      string    `json:"user_id"`
	UserLogin   string    `json:"user_login"`
	UserName    string    `json:"user_name"`
	ExpiresAt   time.Time `json:"expires_at"` // Empty/zero time for permanent bans (Twitch API returns empty string)
	CreatedAt   time.Time `json:"created_at"`
	Reason      string    `json:"reason"`
	ModeratorID string    `json:"moderator_id"`
}

// BannedUsersResponse represents the response from the moderation/banned endpoint
type BannedUsersResponse struct {
	Data       []BannedUser `json:"data"`
	Pagination Pagination   `json:"pagination"`
}

// BanUserRequest represents the request body for banning a user
type BanUserRequest struct {
	UserID   string  `json:"user_id"`
	Duration *int    `json:"duration,omitempty"` // Duration in seconds for timeout, omit for permanent ban
	Reason   *string `json:"reason,omitempty"`
}

// BanUserResponse represents the response from the ban user endpoint
type BanUserResponse struct {
	Data []BanData `json:"data"`
}

// BanData represents ban information returned from ban/unban endpoints
type BanData struct {
	BroadcasterID string `json:"broadcaster_id"`
	ModeratorID   string `json:"moderator_id"`
	UserID        string `json:"user_id"`
	CreatedAt     string `json:"created_at,omitempty"`
	EndTime       string `json:"end_time,omitempty"` // For temporary bans
}
