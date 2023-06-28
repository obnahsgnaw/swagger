package internal

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

type httpConfig struct {
	Debug          bool
	AccessWriter   io.Writer
	ErrWriter      io.Writer
	TrustedProxies []string
}

func newHttpEngine(cnf *httpConfig) (*gin.Engine, error) {
	if cnf.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	if cnf.AccessWriter != nil {
		gin.DisableConsoleColor()
	} else {
		gin.ForceConsoleColor()
	}
	r := gin.New()

	if len(cnf.TrustedProxies) > 0 {
		if err := r.SetTrustedProxies(cnf.TrustedProxies); err != nil {
			return nil, err
		}
	}

	if cnf.AccessWriter != nil {
		r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
			Formatter: func(param gin.LogFormatterParams) string {
				return fmt.Sprintf("[ %s ] - %s %s %s %d %s %v %s %s\n",
					param.TimeStamp.Format(time.RFC3339),
					param.ClientIP,
					param.Method,
					param.Path,
					param.StatusCode,
					param.Latency,
					param.Request.Body,
					param.Request.UserAgent(),
					param.ErrorMessage,
				)
			},
			Output: cnf.AccessWriter,
		}))
	} else {
		r.Use(gin.Logger())
	}
	if cnf.ErrWriter != nil {
		r.Use(gin.RecoveryWithWriter(cnf.ErrWriter))
	} else {
		r.Use(gin.Recovery())
	}
	return r, nil
}
