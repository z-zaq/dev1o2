package models

type Plan struct {
	ID            int
	Name          string
	AssetClass    string
	Duration      int
	RateStructure string
	MinDeposit    float64
	MaxDeposit    float64
}
