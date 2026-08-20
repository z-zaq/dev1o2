package models

// Plan describes an investment offering. RateType currently only supports
// "fixed_compounding": RateValue is the total percentage return over the
// full Duration, compounded daily. See docs/formula note in
// repository/plan.go and AGENT.md §1.3 for the exact formula and the
// reasoning behind it — this field must stay traceable to a disclosed,
// deterministic calculation, not an arbitrary number.
type Plan struct {
	ID         int
	Name       string
	AssetClass string
	Duration   int // days
	RateType   string
	RateValue  float64 // total % return over Duration, compounded daily
	MinDeposit float64
	MaxDeposit float64
}
