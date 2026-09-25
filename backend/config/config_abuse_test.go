package config

import (
	"reflect"
	"testing"
	"time"
)

func TestLoadAbuseDetectionDefaults(t *testing.T) {
	for _, key := range []string{"ABUSE_DETECTION_ENABLED", "ABUSE_READ_PER_MINUTE", "ABUSE_READ_PER_HOUR", "ABUSE_WRITE_PER_MINUTE", "ABUSE_WRITE_PER_HOUR", "ABUSE_BAN_DURATIONS", "ABUSE_OFFENSE_WINDOW", "TRUSTED_PROXIES"} {
		t.Setenv(key, "")
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	abuse := cfg.RateLimit.Abuse
	if !abuse.Enabled || abuse.ReadPerMinute != 1200 || abuse.ReadPerHour != 12000 || abuse.WritePerMinute != 120 || abuse.WritePerHour != 1500 {
		t.Fatalf("unexpected abuse defaults: %+v", abuse)
	}
	if !reflect.DeepEqual(abuse.BanDurations, DefaultAbuseBanDurations) || abuse.OffenseWindow != 24*time.Hour {
		t.Fatalf("unexpected ban defaults: %v / %v", abuse.BanDurations, abuse.OffenseWindow)
	}
	if len(cfg.RateLimit.TrustedProxies) != 6 || cfg.RateLimit.TrustedProxies[2] != "10.0.0.0/8" {
		t.Fatalf("unexpected trusted proxies: %v", cfg.RateLimit.TrustedProxies)
	}
}

func TestLoadAbuseDetectionOverrides(t *testing.T) {
	t.Setenv("ABUSE_DETECTION_ENABLED", "false")
	t.Setenv("ABUSE_READ_PER_MINUTE", "50")
	t.Setenv("ABUSE_BAN_DURATIONS", "5m, 30m")
	t.Setenv("ABUSE_OFFENSE_WINDOW", "6h")
	t.Setenv("TRUSTED_PROXIES", "172.27.0.0/16")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	abuse := cfg.RateLimit.Abuse
	if abuse.Enabled || abuse.ReadPerMinute != 50 || abuse.OffenseWindow != 6*time.Hour {
		t.Fatalf("overrides not applied: %+v", abuse)
	}
	if want := []time.Duration{5 * time.Minute, 30 * time.Minute}; !reflect.DeepEqual(abuse.BanDurations, want) {
		t.Fatalf("BanDurations = %v, want %v", abuse.BanDurations, want)
	}
	if want := []string{"172.27.0.0/16"}; !reflect.DeepEqual(cfg.RateLimit.TrustedProxies, want) {
		t.Fatalf("TrustedProxies = %v, want %v", cfg.RateLimit.TrustedProxies, want)
	}
}

func TestLoadLaunchLoadSheddingDefaults(t *testing.T) {
	for _, key := range []string{"PUBLIC_CACHE_TTL", "PUBLIC_CACHE_STALE_TTL", "ENGAGEMENT_PUBLISH_INTERVAL", "LIVE_STATUS_CANDIDATE_LIMIT"} {
		t.Setenv(key, "")
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Server.PublicCacheTTL != 30*time.Second || cfg.Server.PublicCacheStaleTTL != 15*time.Minute {
		t.Fatalf("public cache defaults = %v / %v", cfg.Server.PublicCacheTTL, cfg.Server.PublicCacheStaleTTL)
	}
	if cfg.Jobs.EngagementPublishInterval != 5*time.Minute || cfg.Jobs.LiveStatusCandidateLimit != 300 {
		t.Fatalf("job defaults = %v / %d", cfg.Jobs.EngagementPublishInterval, cfg.Jobs.LiveStatusCandidateLimit)
	}

	t.Setenv("PUBLIC_CACHE_TTL", "45s")
	t.Setenv("ENGAGEMENT_PUBLISH_INTERVAL", "2m")
	cfg, err = Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Server.PublicCacheTTL != 45*time.Second || cfg.Jobs.EngagementPublishInterval != 2*time.Minute {
		t.Fatalf("overrides not applied: %v / %v", cfg.Server.PublicCacheTTL, cfg.Jobs.EngagementPublishInterval)
	}
}

func TestLoadAbuseBanDurationsRejectsInvalidList(t *testing.T) {
	t.Setenv("ABUSE_BAN_DURATIONS", "15m,forever")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !reflect.DeepEqual(cfg.RateLimit.Abuse.BanDurations, DefaultAbuseBanDurations) {
		t.Fatalf("invalid list should fall back to defaults, got %v", cfg.RateLimit.Abuse.BanDurations)
	}
}
