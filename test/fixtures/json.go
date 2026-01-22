package fixtures

import (
	"encoding/json"
)

func MustJsonMarshal[T any](in T) string {
	bytes, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}

func MustJsonUnmarshal[T any, K []byte | string](out T, in K) T {
	if err := json.Unmarshal([]byte(in), &out); err != nil {
		panic(err)
	}

	return out
}
