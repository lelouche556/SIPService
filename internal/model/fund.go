package model

import "time"

type Fund struct {
	FundID   string `json:"fundId"`
	Name     string `json:"name"`
	AMC      string `json:"amc"`
	Category string `json:"category"`
	RiskTag  string `json:"riskTag"`
	IsActive bool   `json:"isActive"`
}

type MarketPrice struct {
	FundID string    `json:"fundId"`
	NAVMic int64     `json:"navMic"`
	AsOf   time.Time `json:"asOf"`
}
