package interfaces

import "github.com/labstack/echo/v4"

type RegisterRoutes interface {
	RegisterRoutes(*echo.Group)
}
