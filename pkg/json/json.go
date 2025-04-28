package json

import (
	"encoding/json"
	"io"
)

func UnmarshalJson[T any](data []byte) (T, error) {
	var obj T
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return obj, err
	}
	return obj, nil
}

func MarshalJson[T any](obj T) ([]byte, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func Encode(writer io.Writer, v any) error {
	if writer == nil {
		return nil
	}
	return json.NewEncoder(writer).Encode(v)
}
