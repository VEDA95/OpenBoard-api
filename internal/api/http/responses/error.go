package responses

type ErrorResponse struct {
	BaseResponse
	Error GenericMessage `json:"error"`
}

func ErrorResp(code int, message string) *ErrorResponse {
	return &ErrorResponse{
		BaseResponse: BaseResponse{
			Code: code,
		},
		Error: GenericMessage{Message: message},
	}
}
