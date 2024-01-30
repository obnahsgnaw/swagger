package main

import (
	"github.com/obnahsgnaw/application"
	"github.com/obnahsgnaw/application/endtype"
	"github.com/obnahsgnaw/application/pkg/logging/logger"
	"github.com/obnahsgnaw/application/pkg/url"
	http2 "github.com/obnahsgnaw/http"
	"github.com/obnahsgnaw/http/engine"
	"github.com/obnahsgnaw/swagger"
	"time"
)

func main() {
	app := application.New(
		application.NewCluster("dev", "dev"),
		"demo",
		application.Debug(func() bool {
			return true
		}),
		application.EtcdRegister([]string{"127.0.0.1:2379"}, 5*time.Second),
		application.Logger(&logger.Config{
			Dir:        "/Users/wangshanbo/Documents/Data/projects/swagger/out",
			MaxSize:    5,
			MaxBackup:  1,
			MaxAge:     1,
			Level:      "debug",
			TraceLevel: "error",
		}),
	)
	defer app.Release()

	e, _ := http2.Default(url.Host{Ip: "127.0.0.1", Port: 8001}, &engine.Config{
		Name:           "",
		DebugMode:      false,
		LogDebug:       true,
		AccessWriter:   nil,
		ErrWriter:      nil,
		TrustedProxies: nil,
		Cors:           nil,
		LogCnf:         swagger.LogCnf(app, "swg", endtype.Backend),
		DefFavicon:     false,
	})

	s := swagger.New(app, "swg", "swg", e, &swagger.Config{
		EndType:       endtype.Backend,
		GatewayOrigin: nil,
		Prefix:        "v1",
		SubDocs: []swagger.DocItem{
			//{
			//	Module:    "notify-backend",
			//	Title:     "通知管理",
			//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/notify/backend.swagger.json",
			//},
			//{
			//	Module:    "notify-frontend",
			//	Title:     "通知服务",
			//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/notify/frontend.swagger.json",
			//},
			{
				Module:    "company-backend",
				Title:     "公司管理",
				LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/company/backend.swagger.json",
			},
			{
				Module:    "company-frontend",
				Title:     "公司服务",
				LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/company/frontend.swagger.json",
			},
			//{
			//	Module:    "state-backend",
			//	Title:     "设备状态管理",
			//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/state/backend.swagger.json",
			//},
			//{
			//	Module:    "state-frontend",
			//	Title:     "设备状态服务",
			//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/state/frontend.swagger.json",
			//},
			//{
			//	Module:    "perm-backend",
			//	Title:     "权限管理",
			//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/perm/backend.swagger.json",
			//},
			//{
			//	Module:    "perm-frontend",
			//	Title:     "权限服务",
			//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/perm/frontend.swagger.json",
			//},
			//{
			//	Module:    "uavext-backend",
			//	Title:     "设备扩展管理",
			//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/uavext/backend.swagger.json",
			//},
			//{
			//	Module:    "uavext-frontend",
			//	Title:     "设备扩展服务",
			//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/uavext/frontend.swagger.json",
			//},
			//{
			//	Module:    "dynamic-backend",
			//	Title:     "dynamic管理",
			//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/dynamic/backend.swagger.json",
			//},
			//{
			//	Module:    "dynamic-frontend",
			//	Title:     "dynamic服务",
			//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/dynamic/frontend.swagger.json",
			//},
		},
		Tokens: nil,
	})

	app.AddServer(s)

	app.Run(func(err error) {
		panic(err)
	})

	app.Wait()
}
