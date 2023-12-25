package internal

import (
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gin-gonic/gin"
	"github.com/obnahsgnaw/application/pkg/url"
	"github.com/obnahsgnaw/application/pkg/utils"
	"github.com/obnahsgnaw/swagger/asset"
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

func RegisterRoute(engine *gin.Engine, cnf *RouteConfig) error {
	t := template.New("swagger_index.tmpl")
	tmpl, err := asset.Asset("knife4j-vue/dist/index.tmpl")
	if err != nil {
		return utils.NewWrappedError("init doc template failed", err)
	}
	_, err = t.Parse(string(tmpl))
	if err != nil {
		return utils.NewWrappedError("parse doc template failed", err)
	}
	engine.SetHTMLTemplate(t)
	regRoute(engine, cnf.Manager, cnf.Prefix, cnf.GatewayOrigin, cnf.Tokens)

	return nil
}

func regRoute(r *gin.Engine, manager *Manager, prefix string, gwOrigin func() string, tokens []string) {
	r.GET(prefix+"/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, prefix+"/swagger/index")
	})
	// 主页
	r.GET(prefix+"/swagger/index", func(c *gin.Context) {
		ses := GetSession(c.Request)
		if len(tokens) > 0 && ses.Values["logined"] == nil {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, `<form method="post" action="#"><input type="password" name="password" placeholder="Input your password" /><input type="submit" value="Submit" /></form>`)
		} else {
			gws := ""
			if gwOrigin != nil {
				gws = gwOrigin()
			}
			c.HTML(http.StatusOK, "swagger_index.tmpl", gin.H{"gwHost": gws, "prefix": prefix})
		}
	})
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
			ses.Values["logined"] = 1
			_ = ses.Save(c.Request, c.Writer)
		}
		c.Redirect(http.StatusMovedPermanently, prefix+"/swagger/index")
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
		tmpl, _ := asset.Asset("knife4j-vue/dist/favicon.ico")
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write(tmpl)
	})
	// 静态资源
	r.StaticFS(prefix+"/swagger/webjars", &assetfs.AssetFS{
		Asset:    asset.Asset,
		AssetDir: asset.AssetDir,
		AssetInfo: func(path string) (os.FileInfo, error) {
			return os.Stat(path)
		},
		Prefix:   "knife4j-vue/dist/webjars",
		Fallback: "",
	})
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
