package internal

import (
	"github.com/gin-gonic/gin"
	"github.com/obnahsgnaw/application/pkg/url"
	"github.com/obnahsgnaw/application/pkg/utils"
	"github.com/obnahsgnaw/goutils/ginutil"
	knife4jvue "github.com/obnahsgnaw/swagger/knife4j-vue"
	"html/template"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

type RouteConfig struct {
	Manager       *Manager
	Prefix        string
	GatewayOrigin func() string
	Tokens        []string
}

var loginForm = `
<html><head><style>
.pwd{-webkit-appearance: none;background-color: #fff;background-image: none;border-radius: 4px;border: 1px solid #dcdfe6;box-sizing: border-box;color: #606266;diplay: inline-block;font-size: inherit;height: 40px;line-height: 40px;outline: none;padding: 0 15px;transition: border-color .2s cubic-bezier(.645,.045,.355,1);width: 300px;cursor: pointer;
}
.pwd:focus{outline: none;border-color: #409eff;}
.submit{display: inline-block;line-height: 1;white-space: nowrap;cursor: pointer;border: 1px solid #dcdfe6;-webkit-appearance: none;text-align: center;box-sizing: border-box;outline: none;margin: 0;transition: .1s;font-weight: 500;-moz-user-select: none;-webkit-user-select: none;-ms-user-select: none;padding: 12px 20px;font-size: 14px;color: #fff;background-color: #409eff;border-radius: 4px;
}
</style></head>
<body>
<form method="post" action="#"><input class="pwd" type="password" name="password" placeholder="Input your password" autofocus /><input class="submit" type="submit" value="Submit" /></form>
</body>
</html>
`

func RegisterRoute(engine *gin.Engine, cnf *RouteConfig) error {
	t := template.New("swagger_index.tmpl")
	indexTmpl, err := knife4jvue.Assets.ReadFile("dist/index.tmpl")
	if err != nil {
		return utils.NewWrappedError("init doc template failed", err)
	}
	_, err = t.Parse(string(indexTmpl))
	if err != nil {
		return utils.NewWrappedError("parse doc template failed", err)
	}
	engine.SetHTMLTemplate(t)
	regRoute(engine, cnf.Manager, cnf.Prefix, cnf.GatewayOrigin, cnf.Tokens)

	return nil
}

func regRoute(r *gin.Engine, manager *Manager, prefix string, gwOrigin func() string, tokens []string) {
	if prefix == "/" {
		prefix = ""
	}
	r.GET(prefix+"/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, prefix+"/swagger/index")
	})
	var indexHandler = func(c *gin.Context) {
		ses := GetSession(c.Request)
		if len(tokens) > 0 && ses.Values[prefix+"logined"] == nil {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, loginForm)
		} else {
			gws := ""
			if gwOrigin != nil {
				gws = gwOrigin()
			}
			c.HTML(http.StatusOK, "swagger_index.tmpl", gin.H{"gwHost": gws, "gwVersion": "v1", "prefix": prefix})
		}
	}
	// 主页
	r.GET(prefix+"/swagger/index", indexHandler)
	// 主页登录
	r.POST(prefix+"/swagger/index", func(c *gin.Context) {
		ses := GetSession(c.Request)
		pwd := c.Request.FormValue("password")
		success := false
		for _, t := range tokens {
			if pwd == t {
				success = true
				break
			}
		}
		if success {
			ses.Values[prefix+"logined"] = 1
			_ = ses.Save(c.Request, c.Writer)
		}
		indexHandler(c)
	})
	// 文档配置
	r.GET(prefix+"/swagger/static/services.json", func(c *gin.Context) {
		d := manager.DocServices("swaggers")
		if len(d) > 0 {
			c.JSON(http.StatusOK, d)
		} else {
			rawJson(c, []byte("[]"))
		}
	})
	// 图标
	r.GET(prefix+"/swagger/favicon.ico", func(c *gin.Context) {
		favicon, _ := knife4jvue.Assets.ReadFile("dist/favicon.ico")
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write(favicon)
	})
	// 静态资源
	etagManager := ginutil.NewStaticFsCache(r, "dist/webjars", ginutil.Fs(&knife4jvue.Assets), ginutil.RelativePath(prefix+"/swagger/webjars"), ginutil.CaCheTtl(86400))
	if err := etagManager.Init(); err != nil {
		panic("init swagger webjar failed, err=" + err.Error())
	}
	// 子文档代理
	r.GET(prefix+"/swagger/swaggers/:module", func(c *gin.Context) {
		docUrl := manager.GetModuleDocUrl(c.Param("module"))

		if docUrl == "" {
			rawJson(c, []byte("{}"))
		} else {
			if strings.HasPrefix(docUrl, "http") {
				schema, host1, path, _ := url.ParseUrl(docUrl)
				director := func(req *http.Request) {
					req.URL.Scheme = schema.String()
					req.URL.Host = host1.String()
					req.URL.Path = path
				}
				proxy := &httputil.ReverseProxy{Director: director}
				proxy.ServeHTTP(c.Writer, c.Request)
			} else {
				content, _ := os.ReadFile(docUrl)
				rawJson(c, content)
			}
		}
	})
}

func rawJson(c *gin.Context, jsonData []byte) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Status(http.StatusOK)
	_, _ = c.Writer.Write(jsonData)
}
