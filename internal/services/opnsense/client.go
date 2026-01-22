package opnsense

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type ClientAdapter interface {
	CheckUnboundService(ctx context.Context) error
	UnboundSearchHostOverrides(ctx context.Context, req *UnboundSearchHostOverrideRequest) (*UnboundSearchHostOverrideResponse, error)
	UnboundCreateHostOverride(ctx context.Context, req *UnboundHostOverride) (string, error)
	UnboundUpdateHostOverride(ctx context.Context, uuid string, req *UnboundHostOverride) error
	UnboundDeleteHostOverride(ctx context.Context, uuid string) error
	ReconfigureService(ctx context.Context) error
}

type Client struct {
	client        *retryablehttp.Client
	url           string
	apiKey        string
	apiSecret     string
	allowInsecure bool
	isDryRun      bool
	log           *zap.SugaredLogger
}

type ClientSvc struct {
	Logger *services.Logger
}

type ClientConfig struct {
	Uri           string
	APIKey        string
	APISecret     string
	AllowInsecure bool
	DryRun        bool
	MaxRetries    int
	MinBackoff    time.Duration
	MaxBackoff    time.Duration
}

var _ ClientAdapter = (*Client)(nil)

func NewClient(svc *ClientSvc, conf ClientConfig) (*Client, error) {
	httpClient := retryablehttp.NewClient()
	httpClient.Logger = nil

	httpClient.HTTPClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.AllowInsecure},
	}

	httpClient.RetryWaitMax = conf.MaxBackoff
	httpClient.RetryWaitMin = conf.MinBackoff
	httpClient.RetryMax = conf.MaxRetries

	return &Client{
		client:        httpClient,
		url:           conf.Uri,
		apiKey:        conf.APIKey,
		apiSecret:     conf.APISecret,
		allowInsecure: conf.AllowInsecure,
		isDryRun:      conf.DryRun,
		log:           svc.Logger.Sugar(),
	}, nil
}

func (c *Client) auth() string {
	return base64.StdEncoding.EncodeToString([]byte(c.apiKey + ":" + c.apiSecret))
}

func (c *Client) do(ctx context.Context, method string, endpoint string, body any, res any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}

		reader = bytes.NewReader(data)
	}

	path, err := url.JoinPath(c.url, "/api", endpoint)
	if err != nil {
		return fmt.Errorf("failed to build URL path: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, method, path, reader)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", c.auth()))
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	r, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("status code non-200; status code %d", r.StatusCode)
	}

	if res != nil {
		err = json.NewDecoder(r.Body).Decode(res)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) CheckUnboundService(ctx context.Context) error {
	c.log.Debug("Checking Unbound service status")

	req := &ServiceSearchRequest{
		SearchPhrase: "unbound",
		RowCount:     -1,
	}
	res := &ServiceSearchResponse{}
	err := c.do(ctx, http.MethodPost, "/core/service/search", req, res)
	if err != nil {
		return err
	}

	for _, service := range res.Rows {
		if service.Name == "unbound" && service.Running == 1 {
			c.log.Debug("Unbound service is running.")

			return nil
		}
	}

	return fmt.Errorf("unbound service is not running")
}

func (c *Client) UnboundSearchHostOverrides(ctx context.Context, req *UnboundSearchHostOverrideRequest) (*UnboundSearchHostOverrideResponse, error) {
	c.log.Debug("Searching host overrides.")

	if req == nil {
		req = &UnboundSearchHostOverrideRequest{RowCount: -1}
	}

	res := &UnboundSearchHostOverrideResponse{}
	err := c.do(ctx, http.MethodPost, "/unbound/settings/search_host_override", req, res)
	if err != nil {
		return nil, err
	}

	c.log.Debugf("Found host overrides: %d", res.Total)

	return res, nil
}

func (c *Client) UnboundCreateHostOverride(ctx context.Context, override *UnboundHostOverride) (string, error) {
	c.log.Debugf("Creating host override: %+v", override)

	if c.isDryRun {
		c.log.Warn("Dry run enabled, skipping create.")

		return "", nil
	}

	wrapped := map[string]*UnboundHostOverride{
		"host": override,
	}

	res := &UnboundAddHostOverrideResponse{}
	err := c.do(ctx, "POST", "/unbound/settings/addHostOverride", wrapped, res)
	if err != nil {
		return "", err
	}

	if res.Result != "saved" {
		return "", fmt.Errorf("resource not changed. result: %s. errors: %v", res.Result, res.Validations)
	}

	c.log.Debug("Created host override: %+v -> %+v", override, res)

	return res.UUID, nil
}

func (c *Client) UnboundUpdateHostOverride(ctx context.Context, uuid string, override *UnboundHostOverride) error {
	c.log.Debugf("Updating host override: %+v", override)

	if c.isDryRun {
		c.log.Warn("Dry run enabled, skipping update.")

		return nil
	}

	wrapped := map[string]*UnboundHostOverride{
		"host": override,
	}

	res := &UnboundAddHostOverrideResponse{}
	err := c.do(ctx, "POST", fmt.Sprintf("/unbound/settings/setHostOverride/%s", uuid), wrapped, res)
	if err != nil {
		return err
	}

	if res.Result != "saved" {
		return fmt.Errorf("resource not changed. result: %s. errors: %v", res.Result, res.Validations)
	}

	c.log.Debugf("Updated host override: %+v -> %+v", override, res)

	return nil
}

func (c *Client) UnboundDeleteHostOverride(ctx context.Context, uuid string) error {
	c.log.Debugf("Deleting host override: %s", uuid)

	if c.isDryRun {
		c.log.Warn("Dry run enabled, skipping delete.")

		return nil
	}

	res := &UnboundDeleteHostOverrideResponse{}
	err := c.do(ctx, "POST", fmt.Sprintf("/unbound/settings/delHostOverride/%s", uuid), nil, res)
	if err != nil {
		return err
	}

	if res.Result != "deleted" {
		return fmt.Errorf("resource not deleted. result: %s", res.Result)
	}

	c.log.Debugf("Deleted host override: %s", uuid)

	return nil
}

func (c *Client) ReconfigureService(ctx context.Context) error {
	c.log.Debug("Reconfiguring Unbound service.")

	if c.isDryRun {
		c.log.Warn("Dry run enabled, skipping reconfigure.")

		return nil
	}

	resp := &ServiceResponse{}
	err := c.do(ctx, "POST", "/unbound/service/reconfigure", nil, resp)
	if err != nil {
		return err
	}

	status := ""
	if resp.Status != "" {
		status = resp.Status
	} else if resp.Result != "" {
		status = resp.Result
	} else {
		return fmt.Errorf("reconfigure returned with unknown status response")
	}

	status = cases.Lower(language.English).String(strings.TrimSpace(status))
	if status != "ok" {
		return fmt.Errorf("reconfigure failed. status: %s", status)
	}

	c.log.Debug("Reconfigured Unbound service.")

	return nil
}
