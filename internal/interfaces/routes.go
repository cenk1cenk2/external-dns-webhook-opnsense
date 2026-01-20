package interfaces

import "github.com/labstack/echo/v5"

type RegisterRoutes interface {
	RegisterRoutes(*echo.Group)
}
