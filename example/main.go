package main

import (
	"github.com/obnahsgnaw/application"
	"github.com/obnahsgnaw/application/endtype"
	"github.com/obnahsgnaw/application/pkg/url"
	"github.com/obnahsgnaw/swagger"
	"log"
	"time"
)

func main() {
	app := application.New("demo", "Demo")
	app.With(application.EtcdRegister([]string{"127.0.0.1:2379"}, 5*time.Second))

	s := swagger.New(app, "uav", "", &swagger.Config{
		EndType: endtype.Backend,
		Host:    url.Host{Ip: "127.0.0.1", Port: 8001},
		SubDocs: []swagger.DocItem{
			{
				Module:    "user",
				Title:     "用户",
				Url:       url.Url{},
				LocalPath: "/Users/wangshanbo/Documents/Data/projects/swagger/example/backend.swagger.json",
				DebugOrigin: url.Origin{
					Protocol: url.HTTP,
					Host: url.Host{
						Ip:   "127.0.0.1",
						Port: 8001,
					},
				},
			},
		},
	})

	app.AddServer(s)

	app.Run(func(err error) {
		panic(err)
	})

	log.Println("Exited")
}
