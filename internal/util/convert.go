package util

import "github.com/goccy/go-json"

func ConvertType[E interface{}, T interface{}](result E) (*T, error) {
	marshalResult, err := json.Marshal(result)

	if err != nil {
		return nil, err
	}

	var output T

	if err := json.Unmarshal(marshalResult, &output); err != nil {
		return nil, err
	}

	return &output, nil
}
