package models

import "time"

// CreatorDiscoveryProfile summarizes the signals used to place a creator in a
// discovery rail. Twitch category data is supporting context, not identity.
type CreatorDiscoveryProfile struct {
	BroadcasterID       string    `json:"broadcaster_id"`
	BroadcasterName     string    `json:"broadcaster_name"`
	TotalClips          int       `json:"total_clips"`
	RecentClips         int       `json:"recent_clips"`
	TotalViews          int64     `json:"total_views"`
	RecentViews         int64     `json:"recent_views"`
	ViewVelocity        float64   `json:"view_velocity"`
	FollowerCount       int       `json:"follower_count"`
	FirstDiscoveredAt   time.Time `json:"first_discovered_at"`
	LatestClipAt        time.Time `json:"latest_clip_at"`
	LatestClipThumbnail *string   `json:"latest_clip_thumbnail,omitempty"`
	LatestClipTitle     *string   `json:"latest_clip_title,omitempty"`
	TwitchCategoryName  *string   `json:"twitch_category_name,omitempty"`
	Score               float64   `json:"score"`
}

// CreatorDiscoveryRails groups creators by distinct discovery intent.
type CreatorDiscoveryRails struct {
	Trending []CreatorDiscoveryProfile `json:"trending"`
	Rising   []CreatorDiscoveryProfile `json:"rising"`
	New      []CreatorDiscoveryProfile `json:"new"`
}
