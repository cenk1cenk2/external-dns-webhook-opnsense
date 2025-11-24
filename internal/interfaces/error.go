package interfaces

import (
	"github.com/labstack/echo/v4"
)

type ApiError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var _ error = (*ApiError)(nil)

func (e *ApiError) Error() string {
	return e.Message
}

func NewHttpError(code int, err error) error {
	return echo.NewHTTPError(code, err.Error())
}
