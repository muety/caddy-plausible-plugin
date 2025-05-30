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
	"regexp"
	"strings"
	"time"
)

const DefaultBaseUrl = "https://plausible.io"

var regexStaticAssets *regexp.Regexp

func init() {
	regexStaticAssets = regexp.MustCompile(`\.(css|js|png|jpg|jpeg|gif|svg|webp|ico|bmp|tiff|mp3|mp4|avi|mov|webm|ogg|wav|flac|woff|woff2|ttf|map)$`)
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
	rw := &responseWriter{ResponseWriter: w}
	if err := h.ServeHTTP(rw, r); err != nil {
		return err
	}
	go m.recordEvent(r.Clone(context.TODO()), rw.statusCode)
	return nil
}

func (m *PlausiblePlugin) recordEvent(r *http.Request, status int) {
	if status >= 400 {
		return // don't record events for error statuses
	}

	if regexStaticAssets.MatchString(r.URL.Path) {
		return // don't record typical static web assets like css, js, fonts, images and other media
	}

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
	req.Header.Set("X-Forwarded-For", extractIp(r))

	m.logger.Debug("sending plausible event", zap.String("domain", event.Domain), zap.String("url", event.Url))

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
