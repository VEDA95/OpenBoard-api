package responses

type ErrorMessageResponse struct {
	BaseResponse
	Error GenericMessage `json:"error"`
}

type ErrorResponse[T interface{}] struct {
	BaseResponse
	Errors T `json:"errors"`
}

type ErrorCollectionResponse[T interface{}] struct {
	BaseCollectionResponse
	Errors []T `json:"errors"`
}

func ErrorRespMessage(code int, message string) *ErrorMessageResponse {
	return &ErrorMessageResponse{
		BaseResponse: BaseResponse{
			Code: code,
		},
		Error: GenericMessage{Message: message},
	}
}

func ErrorResp[T interface{}](code int, data T) *ErrorResponse[T] {
	return &ErrorResponse[T]{
		BaseResponse: BaseResponse{Code: code},
		Errors:       data,
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
