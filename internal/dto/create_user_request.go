package dto

type CreateUserRequest struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
}
