package api

import (
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/api/probes"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/api/webhook"
)

func (a *Api) RegisterRoutes() {
	group := a.Echo.Group("")

	probes.NewHandler(&probes.HandlerSvc{
		Log:     a.Log,
		IsReady: a.IsReady,
	}).
		RegisterRoutes(group)

	webhook.NewHandler(&webhook.HandlerSvc{
		Log:      a.Log,
		Provider: a.Provider,
	}).
		RegisterRoutes(group)
}
