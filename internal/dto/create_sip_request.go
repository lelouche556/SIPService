package dto

type CreateSIPRequest struct {
	UserID          string `json:"userId"`
	FundID          string `json:"fundId"`
	Mode            string `json:"mode"`
	StartAt         string `json:"startAt"`
	BaseAmountPaise int64  `json:"baseAmountPaise"`
	StepUpEnabled   bool   `json:"stepUpEnabled"`
	StepUpBps       int32  `json:"stepUpBps"`
}
