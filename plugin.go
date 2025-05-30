package caddy_plausible_plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

const DefaultBaseUrl = "https://plausible.io"

func init() {
	caddy.RegisterModule(PlausiblePlugin{})
}

type PlausiblePlugin struct {
	BaseURL    string `json:"base_url,omitempty"`
	DomainName string `json:"domain_name,omitempty"`

	logger *zap.Logger
	client *http.Client
}

type EventPayload struct {
	Name     string `json:"name"`
	Url      string `json:"url"`
	Domain   string `json:"domain"`
	Referrer string `json:"referrer"`
}

func (m PlausiblePlugin) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.plausible",
		New: func() caddy.Module { return new(PlausiblePlugin) },
	}
}

func (m *PlausiblePlugin) Provision(ctx caddy.Context) error {
	if m.DomainName == "" {
		return errors.New("domain_name is required")
	}
	if m.BaseURL == "" {
		m.BaseURL = DefaultBaseUrl
	}
	m.BaseURL = strings.TrimSuffix(m.BaseURL, "/")

	m.client = &http.Client{Timeout: 5 * time.Second}
	m.logger = ctx.Logger(m)

	return nil
}

func (m *PlausiblePlugin) ServeHTTP(w http.ResponseWriter, r *http.Request, h caddyhttp.Handler) error {
	go m.recordEvent(r.Clone(context.TODO()))
	return h.ServeHTTP(w, r)
}

func (m *PlausiblePlugin) recordEvent(r *http.Request) {
	event := EventPayload{
		Name:     "pageview",
		Url:      r.URL.RequestURI(),
		Domain:   m.DomainName,
		Referrer: r.Referer(),
	}
	eventPayload, err := json.Marshal(event)
	if err != nil {
		m.logger.Error("failed to marshal event json", zap.Error(err))
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/event", m.BaseURL), bytes.NewBuffer(eventPayload))
	if err != nil {
		m.logger.Error("failed to construct request", zap.Error(err))
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", r.Header.Get("User-Agent"))
	req.Header.Set("X-Forwarded-For", r.Header.Get("X-Forwarded-For"))

	res, err := m.client.Do(req)
	if err != nil {
		m.logger.Error("failed to post plausible event", zap.Error(err))
		return
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		m.logger.Error("failed to post plausible event, got unsuccessful response", zap.Int("status_code", res.StatusCode))
		return
	}
}
