package fixtures

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetRequestContentJson(req *http.Request) *http.Request {
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	return req
}
