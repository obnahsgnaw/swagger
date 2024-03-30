package swagger

import (
	"errors"
	"github.com/obnahsgnaw/application"
	"github.com/obnahsgnaw/application/endtype"
	"github.com/obnahsgnaw/application/pkg/logging/logger"
	"github.com/obnahsgnaw/application/pkg/url"
	"github.com/obnahsgnaw/application/pkg/utils"
	"github.com/obnahsgnaw/application/regtype"
	"github.com/obnahsgnaw/application/servertype"
	"github.com/obnahsgnaw/application/service/regCenter"
	http2 "github.com/obnahsgnaw/http"
	"github.com/obnahsgnaw/swagger/internal"
	"go.uber.org/zap"
	"io"
	"strings"
)

type Config struct {
}

type DocItem struct {
	Module      string
	Title       string
	Url         url.Url
	LocalPath   string
	DebugOrigin url.Origin
}

type Swagger struct {
	id             string
	name           string
	app            *application.Application
	manager        *internal.Manager
	logger         *zap.Logger
	endType        endtype.EndType
	prefix         string
	gwOriginPd     func() string
	subDocs        []DocItem
	tokens         []string
	accessWriter   io.Writer
	errWriter      io.Writer
	trustedProxies []string
	routeDebug     bool
	engineIgRun    bool
	err            error
	watchInfo      *regCenter.RegInfo
	engine         *http2.Http
}

func New(app *application.Application, id, name string, e *http2.Http, et endtype.EndType, options ...Option) *Swagger {
	s := &Swagger{
		id:      id,
		name:    name,
		app:     app,
		engine:  e,
		endType: et,
		manager: internal.NewManager(),
		logger:  app.Logger().Named("swagger-" + et.String() + "-" + id),
	}

	s.initWatchInfo()
	s.With(options...)

	return s
}

func LogCnf(app *application.Application, id string, et endtype.EndType) *logger.Config {
	cnf := logger.CopyCnf(app.LogConfig())
	if cnf != nil {
		cnf.SetFilename(utils.ToStr("swagger-", et.String(), "-", id))
		cnf.ReplaceTraceLevel(zap.NewAtomicLevelAt(zap.FatalLevel))
	}
	return cnf
}

func (s *Swagger) With(options ...Option) {
	for _, o := range options {
		if o != nil {
			o(s)
		}
	}
}

// ID return the service id
func (s *Swagger) ID() string {
	return s.id
}

// Name return the service name
func (s *Swagger) Name() string {
	return s.name
}

// Type return the server end type
func (s *Swagger) Type() servertype.ServerType {
	return servertype.Api
}

// EndType return the server end type
func (s *Swagger) EndType() endtype.EndType {
	return s.endType
}

func (s *Swagger) Release() {
	s.logger.Info("released")
	_ = s.logger.Sync()
}

func (s *Swagger) Run(failedCb func(err error)) {
	if s.err != nil {
		failedCb(s.err)
		return
	}
	s.logger.Info("init staring...")

	if err := internal.RegisterRoute(s.engine.Engine(), &internal.RouteConfig{
		Manager:       s.manager,
		Prefix:        s.prefix,
		GatewayOrigin: s.gwOriginPd,
		Tokens:        s.tokens,
	}); err != nil {
		failedCb(err)
		return
	}
	s.logger.Debug("engine routes initialized")

	s.logger.Debug("swagger doc watch start")
	if err := s.watch(); err != nil {
		failedCb(err)
		return
	}

	s.logger.Info("initialized")

	s.logger.Info(utils.ToStr("visit ["+url.HTTP.String(), "://", s.engine.Host(), s.prefix, "/swagger/index] to show"))
	if !s.engineIgRun {
		go func() {
			s.logger.Info(utils.ToStr("server[", s.engine.Host(), "] listen and serving..."))
			if err := s.engine.RunAndServ(); err != nil {
				failedCb(err)
			}
		}()
	}
}

func (s *Swagger) Engine() *http2.Http {
	return s.engine
}

func (s *Swagger) initWatchInfo() {
	s.watchInfo = &regCenter.RegInfo{
		AppId:   s.app.Cluster().Id(),
		RegType: regtype.Doc,
		ServerInfo: regCenter.ServerInfo{
			Id:      s.id,
			Name:    s.name,
			EndType: s.endType.String(),
			Type:    servertype.Api.String(),
		},
		Host:      s.engine.Host(),
		Val:       s.engine.Host(),
		Ttl:       s.app.RegTtl(),
		KeyPreGen: regCenter.DefaultRegKeyPrefixGenerator(),
	}
}

func (s *Swagger) watch() error {
	if len(s.subDocs) > 0 {
		for _, doc := range s.subDocs {
			s.logger.Debug(utils.ToStr("sub doc[", doc.Module, "] added"))
			host := s.engine.Host()
			url1 := doc.LocalPath
			if doc.LocalPath == "" {
				host = doc.Url.Origin.Host.String()
				url1 = doc.Url.String()
			}
			s.manager.Add(doc.Module, host, url1, doc.DebugOrigin.String(), doc.Title)
		}
	}
	if s.app.Register() != nil {
		prefix := s.watchInfo.Prefix()
		if prefix == "" {
			return errors.New("reg key prefix is empty")
		}
		return s.app.Register().Watch(s.app.Context(), prefix, func(key string, val string, isDel bool) {
			segments := strings.Split(key, "/")
			module := segments[len(segments)-3]
			host := segments[len(segments)-2]
			attr := segments[len(segments)-1]
			if isDel {
				if attr == "title" {
					s.logger.Debug(utils.ToStr("swagger doc[", module, "] leaved"))
				}
				s.manager.Remove(module, host)
			} else {
				var url1, debugOrigin, name string
				if attr == "title" {
					s.logger.Debug(utils.ToStr("swagger doc[", module, "] added"))
					name = val
				}
				if attr == "url" {
					url1 = val
				}
				if attr == "debugOrigin" {
					debugOrigin = val
				}
				s.manager.Add(module, host, url1, debugOrigin, name)
			}
		})
	}
	return nil
}
