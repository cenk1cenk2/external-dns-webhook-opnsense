package services

import (
	"encoding/json/jsontext"
	json "encoding/json/v2"

	"github.com/labstack/echo/v5"
)

type JsonSerializer struct{}

var _ echo.JSONSerializer = (*JsonSerializer)(nil)

func NewJsonSerializer() *JsonSerializer {
	return &JsonSerializer{}
}

func (s *JsonSerializer) Serialize(c *echo.Context, target any, indent string) error {
	opts := []json.Options{jsontext.EscapeForHTML(true)}
	if indent != "" {
		opts = append(opts, jsontext.WithIndent(indent))
	}

	return json.MarshalWrite(c.Response(), target, opts...)
}

func (s *JsonSerializer) Deserialize(c *echo.Context, target any) error {
	if err := json.UnmarshalRead(c.Request().Body, target, json.RejectUnknownMembers(true)); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	return nil
}
