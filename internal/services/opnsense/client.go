package opnsense

import (
	"context"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/unbound"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	"go.uber.org/zap"
)

type OpnsenseClientAdapter interface {
	SearchHostOverrides(ctx context.Context) (*unbound.SearchHostOverrideResponse, error)
	CreateHostOverride(ctx context.Context, req *unbound.HostOverride) (string, error)
	UpdateHostOverride(ctx context.Context, uuid string, req *unbound.HostOverride) error
	DeleteHostOverride(ctx context.Context, uuid string) error
}

type OpnsenseClient struct {
	client     *api.Client
	controller unbound.Controller

	Log    services.ZapSugaredLogger
	Config OpnsenseClientConfig
}

type OpnsenseClientSvc struct {
	Logger *services.Logger
}

type OpnsenseClientConfig struct {
	api.Options

	DryRun bool
}

var _ OpnsenseClientAdapter = (*OpnsenseClient)(nil)

func NewOpnsenseClient(svc *OpnsenseClientSvc, conf OpnsenseClientConfig) (*OpnsenseClient, error) {
	client := api.NewClient(conf.Options)

	return &OpnsenseClient{
		client:     client,
		controller: unbound.Controller{Api: client},
		Config:     conf,
		Log:        svc.Logger.WithCaller().With(zap.String("service", "opnsense")),
	}, nil
}

func (c *OpnsenseClient) SearchHostOverrides(ctx context.Context) (*unbound.SearchHostOverrideResponse, error) {
	result, err := c.controller.SearchHostOverride(ctx, "-1")
	if err != nil {
		return nil, err
	}

	c.Log.Debugf("Fetched host overrides: count %d/%d -> %+v", result.RowCount, result.Total, result.Rows)

	return result, nil
}

func (c *OpnsenseClient) CreateHostOverride(ctx context.Context, req *unbound.HostOverride) (string, error) {
	c.Log.Debugf("Creating host override: for %s.%s -> %+v", req.Hostname, req.Domain, req)

	if c.Config.DryRun {
		c.Log.Warnln("Dryrun enabled: not creating host override.")

		return "", nil
	}

	return c.controller.AddHostOverride(ctx, req)
}

func (c *OpnsenseClient) UpdateHostOverride(ctx context.Context, uuid string, req *unbound.HostOverride) error {
	c.Log.Debugf("Updating host override: %s -> %+v", uuid, req)

	if c.Config.DryRun {
		c.Log.Warnln("Dryrun enabled: not updating host override.")

		return nil
	}

	return c.controller.UpdateHostOverride(ctx, uuid, req)
}

func (c *OpnsenseClient) DeleteHostOverride(ctx context.Context, uuid string) error {
	c.Log.Debugf("Deleting host override: %s", uuid)

	if c.Config.DryRun {
		c.Log.Warnln("Dryrun enabled: not deleting host override.")
		return nil
	}

	return c.controller.DeleteHostOverride(ctx, uuid)
}
