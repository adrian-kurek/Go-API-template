package jsonutil

import "encoding/json"

func MarshalData(data any) ([]byte, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return []byte{}, err
	}
	return dataBytes, nil
}

func UnmarshalData[T any](dataBytes []byte) (*T, error) {
	var data *T
	err := json.Unmarshal(dataBytes, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
