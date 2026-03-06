package dto

type CreateFundRequest struct {
	FundID   string `json:"fundId"`
	Name     string `json:"name"`
	AMC      string `json:"amc"`
	Category string `json:"category"`
	RiskTag  string `json:"riskTag"`
	IsActive *bool  `json:"isActive"`
	NAVMic   int64  `json:"navMic"`
}
