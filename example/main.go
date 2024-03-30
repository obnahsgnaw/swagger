package main

import (
	"github.com/obnahsgnaw/application"
	"github.com/obnahsgnaw/application/endtype"
	"github.com/obnahsgnaw/application/pkg/logging/logger"
	http2 "github.com/obnahsgnaw/http"
	"github.com/obnahsgnaw/http/engine"
	"github.com/obnahsgnaw/swagger"
)

func main() {
	app := application.New(
		"demo",
		application.Debug(func() bool {
			return true
		}),
		application.Logger(&logger.Config{
			Dir:        "",
			MaxSize:    5,
			MaxBackup:  1,
			MaxAge:     1,
			Level:      "debug",
			TraceLevel: "error",
		}),
	)
	defer app.Release()

	e, _ := http2.Default("127.0.0.1", 8001, &engine.Config{
		Name:           "",
		DebugMode:      false,
		AccessWriter:   nil,
		ErrWriter:      nil,
		TrustedProxies: nil,
		Cors:           nil,
		DefFavicon:     false,
	})

	s := swagger.New(app, "swg", "swg", e, endtype.Frontend)
	s.With(swagger.Prefix("v1"))
	s.With(swagger.SubDocs(
	//swagger.DocItem{
	//	Module:    "notify-backend",
	//	Title:     "通知管理",
	//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/notify/backend.swagger.json",
	//},
	//swagger.DocItem{
	//	Module:    "notify-frontend",
	//	Title:     "通知服务",
	//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/notify/frontend.swagger.json",
	//},
	//swagger.DocItem{
	//	Module:    "company-backend",
	//	Title:     "公司管理",
	//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/company/backend.swagger.json",
	//},
	//swagger.DocItem{
	//	Module:    "company-frontend",
	//	Title:     "公司服务",
	//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/company/frontend.swagger.json",
	//},
	//swagger.DocItem{
	//	Module:    "state-backend",
	//	Title:     "设备状态管理",
	//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/state/backend.swagger.json",
	//},
	//swagger.DocItem{
	//	Module:    "state-frontend",
	//	Title:     "设备状态服务",
	//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/state/frontend.swagger.json",
	//},
	//swagger.DocItem{
	//	Module:    "perm-backend",
	//	Title:     "权限管理",
	//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/perm/backend.swagger.json",
	//},
	//swagger.DocItem{
	//	Module:    "perm-frontend",
	//	Title:     "权限服务",
	//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/perm/frontend.swagger.json",
	//},
	//swagger.DocItem{
	//	Module:    "uavext-backend",
	//	Title:     "设备扩展管理",
	//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/uavext/backend.swagger.json",
	//},
	//swagger.DocItem{
	//	Module:    "uavext-frontend",
	//	Title:     "设备扩展服务",
	//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/uavext/frontend.swagger.json",
	//},
	//swagger.DocItem{
	//	Module:    "dynamic-backend",
	//	Title:     "dynamic管理",
	//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/dynamic/backend.swagger.json",
	//},
	//swagger.DocItem{
	//	Module:    "dynamic-frontend",
	//	Title:     "dynamic服务",
	//	LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/out/demo/dynamic/frontend.swagger.json",
	//},
	))
	s.With(swagger.Tokens("123"))

	app.AddServer(s)

	app.Run(func(err error) {
		panic(err)
	})

	app.Wait()
}
