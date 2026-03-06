package dto

type PaymentCallbackRequest struct {
	PaymentRequestID string `json:"paymentRequestId"`
	Status           string `json:"status"`
	FailureReason    string `json:"failureReason"`
}
