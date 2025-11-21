package fixtures

import (
	"encoding/json"
)

func MustJsonMarshal[T interface{}](in T) string {
	bytes, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}

func MustJsonUnmarshal[T interface{}, K []byte | string](out T, in K) T {
	if err := json.Unmarshal([]byte(in), &out); err != nil {
		panic(err)
	}

	return out
}
