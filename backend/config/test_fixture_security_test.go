package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDevelopmentFixturesCannotEnterReleaseProfiles(t *testing.T) {
	for _, environment := range []string{"development", "test", "staging", "production", ""} {
		for _, mode := range []string{"debug", "release"} {
			cfg := &Config{Server: ServerConfig{Environment: environment, GinMode: mode}, Twitch: TwitchConfig{TestFixtureURL: "http://127.0.0.1:19000"}}
			allowed := mode != "release" && (environment == "development" || environment == "test")
			require.Equal(t, allowed, cfg.AllowsTestLogin())
			if allowed {
				require.NoError(t, cfg.validateEnabledFeatures(false))
			} else {
				require.Error(t, cfg.validateEnabledFeatures(mode == "release"))
			}
		}
	}
}
