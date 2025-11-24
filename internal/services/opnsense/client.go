package opnsense

import (
	"context"
	"fmt"

	"github.com/browningluke/opnsense-go/pkg/api"
	opnsensecore "github.com/browningluke/opnsense-go/pkg/core"
	opnsenseunbound "github.com/browningluke/opnsense-go/pkg/unbound"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	"go.uber.org/zap"
)

type ClientAdapter interface {
	CheckUnboundService(ctx context.Context) error
	UnboundSearchHostOverrides(ctx context.Context) (*opnsenseunbound.SearchHostOverrideResponse, error)
	UnboundCreateHostOverride(ctx context.Context, req *opnsenseunbound.HostOverride) (string, error)
	UnboundUpdateHostOverride(ctx context.Context, uuid string, req *opnsenseunbound.HostOverride) error
	UnboundDeleteHostOverride(ctx context.Context, uuid string) error
}

type Client struct {
	Log    services.ZapSugaredLogger
	Config OpnsenseClientConfig

	client  *api.Client
	core    opnsensecore.Controller
	unbound opnsenseunbound.Controller
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
		core:    opnsensecore.Controller{Api: client},
		unbound: opnsenseunbound.Controller{Api: client},
		Config:  conf,
		Log:     log,
	}, nil
}

func (c *Client) CheckUnboundService(ctx context.Context) error {
	s, err := c.core.ServiceSearch(ctx)
	if err != nil {
		return err
	}

	for _, service := range s.Rows {
		if service.Name == "unbound" && service.Running == 1 {
			return nil
		}
	}

	return fmt.Errorf("unbound service is not running")
}

func (c *Client) UnboundSearchHostOverrides(ctx context.Context) (*opnsenseunbound.SearchHostOverrideResponse, error) {
	result, err := c.unbound.SearchHostOverride(ctx, "-1")
	if err != nil {
		return nil, err
	}

	c.Log.Debugf("Fetched host overrides: count %d/%d -> %+v", result.RowCount, result.Total, result.Rows)

	return result, nil
}

func (c *Client) UnboundCreateHostOverride(ctx context.Context, req *opnsenseunbound.HostOverride) (string, error) {
	c.Log.Debugf("Creating host override: for %s.%s -> %+v", req.Hostname, req.Domain, req)

	if c.Config.DryRun {
		c.Log.Warnln("Dryrun enabled: not creating host override.")

		return "", nil
	}

	return c.unbound.AddHostOverride(ctx, req)
}

func (c *Client) UnboundUpdateHostOverride(ctx context.Context, uuid string, req *opnsenseunbound.HostOverride) error {
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
