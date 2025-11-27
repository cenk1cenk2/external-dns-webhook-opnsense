package api

import (
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/api/webhook"
	"github.com/labstack/echo/v4"
)

func (a *Api) RegisterRoutes(group *echo.Group) {
	webhook.NewHandler(&webhook.HandlerSvc{
		Log:      a.Logger,
		Provider: a.Provider,
	}).
		RegisterRoutes(group)
}
