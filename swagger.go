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
	"github.com/obnahsgnaw/swagger/internal"
	"go.uber.org/zap"
	"io"
	"strings"
)

type Config struct {
	EndType        endtype.EndType
	Host           url.Host
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
}

func New(app *application.Application, id, name string, cnf *Config) *Swagger {
	if cnf.Debugger == nil {
		cnf.Debugger = app.Debugger()
	}
	if cnf.LogCnf == nil {
		cnf.LogCnf = app.LogConfig()
	}
	s := &Swagger{
		id:      id,
		name:    name,
		app:     app,
		cnf:     cnf,
		manager: internal.NewManager(),
	}
	s.logger, s.err = logger.New(utils.ToStr("Swg[", cnf.EndType.String(), ":", id, "]"), cnf.LogCnf, cnf.Debugger.Debug())
	s.watchInfo = &regCenter.RegInfo{
		AppId:   app.ID(),
		RegType: regtype.Doc,
		ServerInfo: regCenter.ServerInfo{
			Id:      s.id,
			Name:    s.name,
			EndType: s.cnf.EndType.String(),
			Type:    servertype.Api.String(),
		},
		Host:      cnf.Host.String(),
		Val:       cnf.Host.String(),
		Ttl:       cnf.RegTtl,
		KeyPreGen: regCenter.DefaultRegKeyPrefixGenerator(),
	}

	return s
}

// ID return the api service id
func (s *Swagger) ID() string {
	return s.id
}

// Name return the api service name
func (s *Swagger) Name() string {
	return s.name
}

// Type return the api server end type
func (s *Swagger) Type() servertype.ServerType {
	return servertype.Api
}

// EndType return the api server end type
func (s *Swagger) EndType() endtype.EndType {
	return s.cnf.EndType
}

func (s *Swagger) Release() {

}

func (s *Swagger) Run(failedCb func(err error)) {
	if s.err != nil {
		failedCb(s.err)
		return
	}
	if s.cnf.Prefix != "" {
		s.cnf.Prefix = "/" + strings.Trim(s.cnf.Prefix, "/")
	}
	s.engine, s.err = internal.NewEngine(&internal.Config{
		Prefix:         s.cnf.Prefix,
		Debugger:       s.cnf.Debugger,
		AccessWriter:   s.cnf.AccessWriter,
		ErrWriter:      s.cnf.ErrWriter,
		TrustedProxies: s.cnf.TrustedProxies,
		GatewayOrigin:  s.cnf.GatewayOrigin,
		Tokens:         s.cnf.Tokens,
		Manager:        s.manager,
	})
	if s.err != nil {
		failedCb(s.err)
		return
	}
	if err := s.watch(); err != nil {
		failedCb(err)
		return
	}
	go func() {
		s.logger.Info(utils.ToStr("swg[", s.cnf.Host.String(), "] listen and serving..."))

		s.logger.Info("listen and serving..., visit [" + url.HTTP.String() + "://" + s.cnf.Host.String() + s.cnf.Prefix + "/index] to show")
		if err := s.engine.Run(s.cnf.Host.String()); err != nil {
			failedCb(err)
		}
	}()
}

func (s *Swagger) watch() error {
	if len(s.cnf.SubDocs) > 0 {
		for _, doc := range s.cnf.SubDocs {
			s.debug(utils.ToStr("sub doc[", doc.Module, "] added"))
			host := s.cnf.Host.String()
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
				s.debug(utils.ToStr("swg[", module, "] leaved"))
				s.manager.Remove(module, host)
			} else {
				s.debug(utils.ToStr("swg[", module, "] added"))
				var url1, debugOrigin, name string
				if attr == "title" {
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

func (s *Swagger) debug(msg string) {
	if s.cnf.Debugger != nil && s.cnf.Debugger.Debug() {
		s.logger.Debug(msg)
	}
}
