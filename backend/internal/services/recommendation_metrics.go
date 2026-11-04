package services

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var onboardingCompletedTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "recommendations_onboarding_completed_total",
		Help: "Total number of users who completed onboarding",
	},
)

// RecordOnboardingCompleted counts a completed creator-first onboarding.
func RecordOnboardingCompleted() {
	onboardingCompletedTotal.Inc()
}
