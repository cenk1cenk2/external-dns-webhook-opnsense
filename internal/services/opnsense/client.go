package opnsense

import (
	"context"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/unbound"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	"go.uber.org/zap"
)

type ClientAdapter interface {
	UnboundSearchHostOverrides(ctx context.Context) (*unbound.SearchHostOverrideResponse, error)
	UnboundCreateHostOverride(ctx context.Context, req *unbound.HostOverride) (string, error)
	UnboundUpdateHostOverride(ctx context.Context, uuid string, req *unbound.HostOverride) error
	UnboundDeleteHostOverride(ctx context.Context, uuid string) error
}

type Client struct {
	client  *api.Client
	unbound unbound.Controller

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

var _ ClientAdapter = (*Client)(nil)

func NewClient(svc *OpnsenseClientSvc, conf OpnsenseClientConfig) (*Client, error) {
	client := api.NewClient(conf.Options)

	log := svc.Logger.WithCaller().With(zap.String("service", "opnsense"))

	if conf.DryRun {
		log.Warnln("Dryrun mode enabled: no changes will be applied to OPNsense.")
	}
	conf.Logger = nil

	return &Client{
		client:  client,
		unbound: unbound.Controller{Api: client},
		Config:  conf,
		Log:     log,
	}, nil
}

func (c *Client) UnboundSearchHostOverrides(ctx context.Context) (*unbound.SearchHostOverrideResponse, error) {
	result, err := c.unbound.SearchHostOverride(ctx, "-1")
	if err != nil {
		return nil, err
	}

	c.Log.Debugf("Fetched host overrides: count %d/%d -> %+v", result.RowCount, result.Total, result.Rows)

	return result, nil
}

func (c *Client) UnboundCreateHostOverride(ctx context.Context, req *unbound.HostOverride) (string, error) {
	c.Log.Debugf("Creating host override: for %s.%s -> %+v", req.Hostname, req.Domain, req)

	if c.Config.DryRun {
		c.Log.Warnln("Dryrun enabled: not creating host override.")

		return "", nil
	}

	return c.unbound.AddHostOverride(ctx, req)
}

func (c *Client) UnboundUpdateHostOverride(ctx context.Context, uuid string, req *unbound.HostOverride) error {
	c.Log.Debugf("Updating host override: %s -> %+v", uuid, req)

	if c.Config.DryRun {
		c.Log.Warnln("Dryrun enabled: not updating host override.")

		return nil
	}

	return c.unbound.UpdateHostOverride(ctx, uuid, req)
}

func (c *Client) UnboundDeleteHostOverride(ctx context.Context, uuid string) error {
	c.Log.Debugf("Deleting host override: %s", uuid)

	if c.Config.DryRun {
		c.Log.Warnln("Dryrun enabled: not deleting host override.")
		return nil
	}

	return c.unbound.DeleteHostOverride(ctx, uuid)
}
