package errors

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code" example:"validation_error"`
	Message string `json:"message" example:"invalid request"`
}
