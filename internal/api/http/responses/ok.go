package responses

type OkResponse[T interface{}] struct {
	BaseResponse
	Data T `json:"data"`
}

type OkCollectionResponse[T interface{}] struct {
	BaseCollectionResponse
	Data []T `json:"data"`
}

func OKResponse[T interface{}](code int, data T) *OkResponse[T] {
	return &OkResponse[T]{
		BaseResponse: BaseResponse{
			Code: code,
		},
		Data: data,
	}
}

func OKCollectionResponse[T interface{}](code int, data []T) *OkCollectionResponse[T] {
	return &OkCollectionResponse[T]{
		BaseCollectionResponse: BaseCollectionResponse{
			BaseResponse: BaseResponse{
				Code: code,
			},
			Count: len(data),
		},
		Data: data,
	}
}
