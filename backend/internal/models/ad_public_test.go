package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAdPublicRedactsCampaignInternals(t *testing.T) {
	ad := Ad{
		Name: "creative", AdvertiserName: "advertiser", DailyBudgetCents: int64Pointer(1000),
		SpentTotalCents: 500, CPMCents: 250, TargetingCriteria: map[string]interface{}{"country": "US"},
	}
	payload, err := json.Marshal(ad.Public())
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"budget", "spent", "cpm", "targeting", "experiment"} {
		if strings.Contains(strings.ToLower(string(payload)), forbidden) {
			t.Fatalf("public ad leaked %q: %s", forbidden, payload)
		}
	}
}

func int64Pointer(value int64) *int64 { return &value }
