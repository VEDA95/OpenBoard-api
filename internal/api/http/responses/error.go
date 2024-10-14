package responses

type ErrorResponse struct {
	BaseResponse
	Error GenericMessage `json:"error"`
}

type ErrorCollectionResponse[T interface{}] struct {
	BaseCollectionResponse
	Errors []T `json:"errors"`
}

func ErrorResp(code int, message string) *ErrorResponse {
	return &ErrorResponse{
		BaseResponse: BaseResponse{
			Code: code,
		},
		Error: GenericMessage{Message: message},
	}
}

func ErrorCollectionResp[T interface{}](code int, errors []T) *ErrorCollectionResponse[T] {
	return &ErrorCollectionResponse[T]{
		BaseCollectionResponse: BaseCollectionResponse{
			BaseResponse: BaseResponse{
				Code: code,
			},
			Count: len(errors),
		},
		Errors: errors,
	}
}
