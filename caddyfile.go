package caddy_plausible_plugin

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"regexp"
)

func init() {
	httpcaddyfile.RegisterHandlerDirective("plausible", parseCaddyfile)
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	p := new(PlausiblePlugin)

	for h.Next() {
		for h.NextBlock(0) {
			switch h.Val() {
			case "domain_name":
				var domainName string
				if !h.AllArgs(&domainName) {
					return nil, h.ArgErr()
				}
				p.DomainName = domainName
			case "base_url":
				var baseUrl string
				if !h.AllArgs(&baseUrl) {
					return nil, h.ArgErr()
				}
				urlRegex := regexp.MustCompile(`https?://(www\.)?[-a-zA-Z0-9@:%._+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)
				if !urlRegex.MatchString(baseUrl) {
					return nil, h.Errf("'%s' is not a valid url", baseUrl)
				}
				p.BaseURL = baseUrl
			default:
				return nil, h.Errf("unrecognized option '%s'", h.Val())
			}
		}
	}

	return p, nil
}
