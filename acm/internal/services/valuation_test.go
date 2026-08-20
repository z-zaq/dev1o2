package services

import (
	"acm/internal/models"
	"math"
	"testing"
	"time"
)

func TestCalculateInvestmentValueAtStart(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 90)

	investment := models.Investment{
		Principal: 1000,
		StartedAt: start,
		EndsAt:    end,
		Status:    "active",
	}

	plan := models.Plan{
		Duration:  90,
		RateType:  "fixed_compounding",
		RateValue: 15,
	}

	result, err := CalculateInvestmentValue(
		investment,
		plan,
		start,
	)

	if err != nil {
		t.Fatal(err)
	}

	if math.Abs(result.CurrentValue-1000) > 0.01 {
		t.Fatalf(
			"expected starting value to be 1000, got %.2f",
			result.CurrentValue,
		)
	}

	if math.Abs(result.Profit) > 0.01 {
		t.Fatalf(
			"expected starting profit to be 0, got %.2f",
			result.Profit,
		)
	}
}

func TestCalculateInvestmentValueAtMaturity(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 90)

	investment := models.Investment{
		Principal: 1000,
		StartedAt: start,
		EndsAt:    end,
		Status:    "active",
	}

	plan := models.Plan{
		Duration:  90,
		RateType:  "fixed_compounding",
		RateValue: 15,
	}

	result, err := CalculateInvestmentValue(
		investment,
		plan,
		end,
	)

	if err != nil {
		t.Fatal(err)
	}

	expectedValue := 1150.0
	expectedProfit := 150.0

	if math.Abs(result.CurrentValue-expectedValue) > 0.01 {
		t.Fatalf(
			"expected maturity value %.2f, got %.2f",
			expectedValue,
			result.CurrentValue,
		)
	}

	if math.Abs(result.Profit-expectedProfit) > 0.01 {
		t.Fatalf(
			"expected maturity profit %.2f, got %.2f",
			expectedProfit,
			result.Profit,
		)
	}

	if result.DaysElapsed != 90 {
		t.Fatalf(
			"expected 90 elapsed days, got %d",
			result.DaysElapsed,
		)
	}

	if result.DaysRemaining != 0 {
		t.Fatalf(
			"expected 0 remaining days, got %d",
			result.DaysRemaining,
		)
	}

	if !result.Matured {
		t.Fatal("expected investment to be matured")
	}

	if math.Abs(result.Progress-1.0) > 0.0001 {
		t.Fatalf(
			"expected progress to be 100%%, got %.4f",
			result.Progress,
		)
	}
}

func TestCalculateInvestmentValueHalfway(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 90)

	investment := models.Investment{
		Principal: 1000,
		StartedAt: start,
		EndsAt:    end,
		Status:    "active",
	}

	plan := models.Plan{
		Duration:  90,
		RateType:  "fixed_compounding",
		RateValue: 15,
	}

	midpoint := start.AddDate(0, 0, 45)

	result, err := CalculateInvestmentValue(
		investment,
		plan,
		midpoint,
	)

	if err != nil {
		t.Fatal(err)
	}

	if result.CurrentValue <= 1000 {
		t.Fatalf(
			"expected investment to have earned some profit, got %.2f",
			result.CurrentValue,
		)
	}

	if result.CurrentValue >= 1150 {
		t.Fatalf(
			"expected midpoint value to be below maturity value, got %.2f",
			result.CurrentValue,
		)
	}

	if result.DaysElapsed != 45 {
		t.Fatalf(
			"expected 45 elapsed days, got %d",
			result.DaysElapsed,
		)
	}

	if result.DaysRemaining != 45 {
		t.Fatalf(
			"expected 45 remaining days, got %d",
			result.DaysRemaining,
		)
	}

	if result.Matured {
		t.Fatal("expected investment to still be active")
	}

	if result.Progress <= 0.49 || result.Progress >= 0.51 {
		t.Fatalf(
			"expected progress to be approximately 50%%, got %.4f",
			result.Progress,
		)
	}
}

func TestCalculateInvestmentValueBeforeStart(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 90)

	investment := models.Investment{
		Principal: 1000,
		StartedAt: start,
		EndsAt:    end,
		Status:    "active",
	}

	plan := models.Plan{
		Duration:  90,
		RateType:  "fixed_compounding",
		RateValue: 15,
	}

	beforeStart := start.Add(-24 * time.Hour)

	result, err := CalculateInvestmentValue(
		investment,
		plan,
		beforeStart,
	)

	if err != nil {
		t.Fatal(err)
	}

	if math.Abs(result.CurrentValue-1000) > 0.01 {
		t.Fatalf(
			"expected value before start to remain 1000, got %.2f",
			result.CurrentValue,
		)
	}

	if result.DaysElapsed != 0 {
		t.Fatalf(
			"expected 0 elapsed days before start, got %d",
			result.DaysElapsed,
		)
	}
}

func TestCalculateInvestmentValueAfterMaturity(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 90)

	investment := models.Investment{
		Principal: 1000,
		StartedAt: start,
		EndsAt:    end,
		Status:    "active",
	}

	plan := models.Plan{
		Duration:  90,
		RateType:  "fixed_compounding",
		RateValue: 15,
	}

	afterMaturity := end.AddDate(0, 0, 30)

	result, err := CalculateInvestmentValue(
		investment,
		plan,
		afterMaturity,
	)

	if err != nil {
		t.Fatal(err)
	}

	if math.Abs(result.CurrentValue-1150) > 0.01 {
		t.Fatalf(
			"expected value to remain capped at 1150 after maturity, got %.2f",
			result.CurrentValue,
		)
	}

	if result.DaysRemaining != 0 {
		t.Fatalf(
			"expected 0 remaining days after maturity, got %d",
			result.DaysRemaining,
		)
	}

	if !result.Matured {
		t.Fatal("expected investment to remain matured")
	}
}
