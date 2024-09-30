package responses

type BaseResponse struct {
	Code int `json:"code"`
}

type BaseCollectionResponse struct {
	BaseResponse
	Count int `json:"count"`
}

type GenericMessage struct {
	Message string `json:"message"`
}
