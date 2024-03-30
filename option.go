package swagger

import (
	"io"
	"strings"
)

type Option func(*Swagger)

func Prefix(prefix string) Option {
	return func(s *Swagger) {
		s.prefix = "/" + strings.Trim(prefix, "/")
	}
}

func GatewayOrigin(gwo func() string) Option {
	return func(s *Swagger) {
		s.gwOriginPd = gwo
	}
}

func SubDocs(d ...DocItem) Option {
	return func(s *Swagger) {
		s.subDocs = append(s.subDocs, d...)
	}
}

func Tokens(tokens ...string) Option {
	return func(s *Swagger) {
		s.tokens = append(s.tokens, tokens...)
	}
}

func AccessWriter(w io.Writer) Option {
	return func(s *Swagger) {
		s.accessWriter = w
	}
}

func ErrWriter(w io.Writer) Option {
	return func(s *Swagger) {
		s.errWriter = w
	}
}

func TrustedProxies(proxies ...string) Option {
	return func(s *Swagger) {
		s.trustedProxies = append(s.trustedProxies, proxies...)
	}
}

func RouteDebug(enable bool) Option {
	return func(s *Swagger) {
		s.routeDebug = enable
	}
}

func EngineIgRun(ig bool) Option {
	return func(s *Swagger) {
		s.engineIgRun = ig
	}
}
