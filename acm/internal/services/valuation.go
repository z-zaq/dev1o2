package services

import (
	"acm/internal/models"
	"errors"
	"math"
	"time"
)

// InvestmentValuation represents the calculated financial state
// of an investment at a specific point in time.
type InvestmentValuation struct {
	Principal     float64
	CurrentValue  float64
	Profit        float64
	Progress      float64
	DaysElapsed   int
	DaysRemaining int
	Matured       bool
}

// CalculateInvestmentValue calculates the current value of an investment.
//
// RateValue represents the TOTAL percentage return over the investment's
// complete duration. The return is distributed using daily compounding.
//
// Example:
//
//	Principal = 1000
//	RateValue = 15
//	Duration  = 90 days
//
// At maturity the value will be exactly 1150.
func CalculateInvestmentValue(
	investment models.Investment,
	plan models.Plan,
	now time.Time,
) (InvestmentValuation, error) {

	if investment.Principal < 0 {
		return InvestmentValuation{}, errors.New("investment principal cannot be negative")
	}

	if plan.Duration <= 0 {
		return InvestmentValuation{}, errors.New("investment plan duration must be greater than zero")
	}

	if plan.RateType != "fixed_compounding" {
		return InvestmentValuation{}, errors.New("unsupported investment rate type")
	}

	if plan.RateValue < 0 {
		return InvestmentValuation{}, errors.New("investment rate cannot be negative")
	}

	start := investment.StartedAt
	end := investment.EndsAt

	if end.Before(start) {
		return InvestmentValuation{}, errors.New("investment end date cannot be before start date")
	}

	// Calculate the total investment duration from the plan.
	duration := end.Sub(start)

	if duration <= 0 {
		return InvestmentValuation{}, errors.New("investment duration must be greater than zero")
	}

	// Prevent calculations from going beyond maturity.
	effectiveNow := now
	if effectiveNow.Before(start) {
		effectiveNow = start
	}

	if effectiveNow.After(end) {
		effectiveNow = end
	}

	elapsed := effectiveNow.Sub(start)

	// Daily compounding rate required to reach the advertised
	// total return exactly at maturity.
	totalReturn := plan.RateValue / 100

	daysInDuration := duration.Hours() / 24

	dailyRate := math.Pow(
		1+totalReturn,
		1/daysInDuration,
	) - 1

	elapsedDays := elapsed.Hours() / 24

	currentValue := investment.Principal *
		math.Pow(1+dailyRate, elapsedDays)

	maturityValue := investment.Principal * (1 + totalReturn)

	if currentValue > maturityValue {
		currentValue = maturityValue
	}

	profit := currentValue - investment.Principal

	progress := elapsedDays / daysInDuration

	if progress < 0 {
		progress = 0
	}

	if progress > 1 {
		progress = 1
	}

	daysElapsed := int(math.Floor(elapsedDays))
	daysRemaining := int(math.Ceil(daysInDuration - elapsedDays))

	if daysRemaining < 0 {
		daysRemaining = 0
	}

	return InvestmentValuation{
		Principal:     investment.Principal,
		CurrentValue:  currentValue,
		Profit:        profit,
		Progress:      progress,
		DaysElapsed:   daysElapsed,
		DaysRemaining: daysRemaining,
		Matured:       !effectiveNow.Before(end),
	}, nil
}
