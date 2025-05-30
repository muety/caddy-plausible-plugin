[![Go](https://github.com/muety/caddy-pirsch-plugin/workflows/Go/badge.svg)](https://github.com/muety/caddy-plausible-plugin/actions)
![Coding Time](https://img.shields.io/endpoint?url=https://wakapi.dev/api/compat/shields/v1/n1try/interval:any/project:caddy-plausible-plugin&color=blue&label=coding%20time)

# caddy-plausible-plugin

A Caddy v2 plugin to track requests in [Plausible Analytics](https://plausible.io) from the server side. Inspired by [caddy-pirsch-plugin](https://github.com/muety/caddy-pirsch-plugin).

## Usage
```
plausible [<matcher>] {
    domain_name <your-project-domain>
    base_url <alternative-api-url>
}
```

Because this directive does not come standard with Caddy, you need to [put the directive in order](https://caddyserver.com/docs/caddyfile/options). The correct place is up to you, but usually putting it near the end works if no other terminal directives match the same requests. It's common to pair a Pirsch handler with a `file_server`, so ordering it just before is often a good choice:

```
{
	order plausible before file_server
}
```

Alternatively, you may use `route` to order it the way you want. For example:

```
localhost
root * /srv
route {
	plausible * {
		[...]
	}
	file_server
}
```

### Example
Track all requests to HTML pages in Plausible. You might want to extend the matcher regexp to also include `/` or, alternatively, match everything but assets (like `.css`, `.js`, ...) since usually you wouldn't want to track those.

```
{
    order plausible before file_server
}

http://localhost:8080 {
    @html path_regexp .*\.html$

    plausible @html {
    }

    file_server
}
```

## Development
### Build
```bash
xcaddy build --with github.com/muety/caddy-plausible-plugin=.
```

## License
Apache 2.0