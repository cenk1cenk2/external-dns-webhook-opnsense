package fixtures

import (
	"encoding/json/jsontext"
	json "encoding/json/v2"
)

func MustJsonMarshal[T any](in T) string {
	bytes, err := json.Marshal(in, jsontext.EscapeForHTML(true))
	if err != nil {
		panic(err)
	}

	return string(bytes)
}

func MustJsonUnmarshal[T any, K []byte | string](out T, in K) T {
	if err := json.Unmarshal([]byte(in), &out, json.RejectUnknownMembers(true)); err != nil {
		panic(err)
	}

	return out
}
