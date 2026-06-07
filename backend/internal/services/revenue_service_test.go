package services

import (
	"testing"

	"git.subcult.tv/subculture-collective/clpr/config"
)

func TestNewRevenueService(t *testing.T) {
	cfg := &config.Config{
		Stripe: config.StripeConfig{
			ProMonthlyPriceID: "price_monthly_123",
			ProYearlyPriceID:  "price_yearly_456",
		},
	}

	// Can't test with nil repo since it would panic on real calls
	// This just tests that the service can be constructed
	service := NewRevenueService(nil, cfg)

	if service == nil {
		t.Error("Expected service to be created")
	}

	// Verify price mapping was set up
	if len(service.priceMapping) != 2 {
		t.Errorf("Expected 2 price mappings, got %d", len(service.priceMapping))
	}
}

func TestRevenueService_SetPriceMapping(t *testing.T) {
	cfg := &config.Config{}
	service := NewRevenueService(nil, cfg)

	// Set custom price mapping
	customMapping := map[string]float64{
		"price_test_1": 1999,
		"price_test_2": 2499,
	}
	service.SetPriceMapping(customMapping)

	if len(service.priceMapping) != 2 {
		t.Errorf("Expected 2 price mappings after SetPriceMapping, got %d", len(service.priceMapping))
	}

	if service.priceMapping["price_test_1"] != 1999 {
		t.Errorf("Expected price_test_1 to be 1999, got %f", service.priceMapping["price_test_1"])
	}
}
