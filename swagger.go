package swagger

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/obnahsgnaw/application"
	"github.com/obnahsgnaw/application/endtype"
	"github.com/obnahsgnaw/application/pkg/debug"
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
	EndType        endtype.EndType
	Debugger       debug.Debugger
	LogCnf         *logger.Config
	Prefix         string
	GatewayOrigin  func() string
	SubDocs        []DocItem
	Tokens         []string
	RegTtl         int64
	AccessWriter   io.Writer
	ErrWriter      io.Writer
	TrustedProxies []string
	RouteDebug     bool
}

type DocItem struct {
	Module      string
	Title       string
	Url         url.Url
	LocalPath   string
	DebugOrigin url.Origin
}

type Swagger struct {
	id        string
	name      string
	app       *application.Application
	cnf       *Config
	manager   *internal.Manager
	logger    *zap.Logger
	err       error
	watchInfo *regCenter.RegInfo
	engine    *gin.Engine
	host      url.Host
	engineCus bool
}

func New(app *application.Application, id, name string, cnf *Config) *Swagger {
	if cnf.Debugger == nil {
		cnf.Debugger = app.Debugger()
	}
	if cnf.LogCnf == nil {
		cnf.LogCnf = logger.CopyCnf(app.LogConfig())
		if cnf.LogCnf != nil {
			cnf.LogCnf.AddSubDir(cnf.EndType.String(), "swagger", id)
			cnf.LogCnf.SetFilename("swagger")
			cnf.LogCnf.ReplaceTraceLevel(zap.NewAtomicLevelAt(zap.FatalLevel))
		}
	}
	s := &Swagger{
		id:      id,
		name:    name,
		app:     app,
		cnf:     cnf,
		manager: internal.NewManager(),
	}
	s.logger, s.err = logger.New("swagger:swagger", cnf.LogCnf, cnf.Debugger.Debug())

	return s
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
	return s.cnf.EndType
}

func (s *Swagger) Release() {
	if s.logger != nil {
		s.logger.Info("released")
		_ = s.logger.Sync()
	}
}

func (s *Swagger) Run(failedCb func(err error)) {
	if s.err != nil {
		failedCb(s.err)
		return
	}
	s.logger.Info("init staring...")
	if s.cnf.Prefix != "" {
		s.cnf.Prefix = "/" + strings.Trim(s.cnf.Prefix, "/")
	}
	if !s.engineCus {
		s.logger.Debug("engine initialized(default)")
	} else {
		s.logger.Debug("engine initialized(customer)")
	}
	if err := internal.RegisterRoute(s.engine, &internal.RouteConfig{
		Manager:       s.manager,
		Prefix:        s.cnf.Prefix,
		GatewayOrigin: s.cnf.GatewayOrigin,
		Tokens:        s.cnf.Tokens,
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

	s.logger.Info(utils.ToStr("visit ["+url.HTTP.String(), "://", s.host.String(), s.cnf.Prefix, "/swagger/index] to show"))
	if !s.engineCus {
		go func() {
			s.logger.Info(utils.ToStr("server[", s.host.String(), "] listen and serving..."))
			if err := s.engine.Run(s.host.String()); err != nil {
				failedCb(err)
			}
		}()
	}
}

func (s *Swagger) Engine() *gin.Engine {
	return s.engine
}

func (s *Swagger) WithEngineIns(e *http2.PortedEngine) {
	s.engine = e.Engine()
	s.host = e.Host()
	s.engineCus = true
	s.initWatchInfo()
}

func (s *Swagger) WithEngine(host url.Host) {
	if s.err != nil {
		return
	}
	s.engine, s.err = internal.NewEngine(&http2.Config{
		Name:           "swagger",
		DebugMode:      s.cnf.RouteDebug,
		LogDebug:       s.cnf.AccessWriter == nil,
		AccessWriter:   s.cnf.AccessWriter,
		ErrWriter:      s.cnf.ErrWriter,
		TrustedProxies: s.cnf.TrustedProxies,
		Cors:           nil,
		LogCnf:         s.cnf.LogCnf,
	})
	s.host = host
	s.initWatchInfo()
}

func (s *Swagger) initWatchInfo() {
	s.watchInfo = &regCenter.RegInfo{
		AppId:   s.app.ID(),
		RegType: regtype.Doc,
		ServerInfo: regCenter.ServerInfo{
			Id:      s.id,
			Name:    s.name,
			EndType: s.cnf.EndType.String(),
			Type:    servertype.Api.String(),
		},
		Host:      s.host.String(),
		Val:       s.host.String(),
		Ttl:       s.cnf.RegTtl,
		KeyPreGen: regCenter.DefaultRegKeyPrefixGenerator(),
	}
}
func (s *Swagger) watch() error {
	if len(s.cnf.SubDocs) > 0 {
		for _, doc := range s.cnf.SubDocs {
			s.logger.Debug(utils.ToStr("sub doc[", doc.Module, "] added"))
			host := s.host.String()
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
