package microgate

import (
	"encoding/gob"
	"github.com/corazawaf/coraza/v2"
	"github.com/corazawaf/coraza/v2/seclang"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/secure"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	corazagin "github.com/jptosso/coraza-gin"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type LimeServer struct {
	Server   *http.Server
	Router   *gin.Engine
	Security struct {
		AllowedHosts []string
		STSSeconds   int64
		AllowOrigins []string
	}
}

var server *LimeServer

func Server(port string, readTimeout time.Duration, writeTimeout time.Duration) *LimeServer {
	server = &LimeServer{}
	server.Server = &http.Server{
		Addr:           ":" + port,
		ReadTimeout:    readTimeout * time.Second,
		WriteTimeout:   writeTimeout * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return server
}

type PathSecurity struct {
	UseCache    bool
	CachePublic bool
	MaxAge      string
	Expires     time.Time
}

func (lime *LimeServer) InitSecurity(allowedHosts []string, sTSSeconds int64, allowOrigins []string) {
	lime.Security.STSSeconds = sTSSeconds
	lime.Security.AllowedHosts = allowedHosts
	lime.Security.AllowOrigins = allowOrigins
}

func (lime *LimeServer) GET(sec PathSecurity, relativePath string, handlers ...gin.HandlerFunc) {
	if sec.UseCache {
		lime.Router.GET(relativePath, handlers...).Use(addCache(sec))
	} else {
		lime.Router.GET(relativePath, handlers...)
	}
}

func (lime *LimeServer) POST(sec PathSecurity, relativePath string, handlers ...gin.HandlerFunc) {
	if sec.UseCache {
		lime.Router.POST(relativePath, handlers...).Use(addCache(sec))
	} else {
		lime.Router.POST(relativePath, handlers...)
	}
}

func (lime *LimeServer) PUT(sec PathSecurity, relativePath string, handlers ...gin.HandlerFunc) {
	if sec.UseCache {
		lime.Router.PUT(relativePath, handlers...).Use(addCache(sec))
	} else {
		lime.Router.PUT(relativePath, handlers...)
	}
}

func (lime *LimeServer) DELETE(sec PathSecurity, relativePath string, handlers ...gin.HandlerFunc) {
	if sec.UseCache {
		lime.Router.DELETE(relativePath, handlers...).Use(addCache(sec))
	} else {
		lime.Router.DELETE(relativePath, handlers...)
	}
}

func (lime *LimeServer) ANY(sec PathSecurity, relativePath string, handlers ...gin.HandlerFunc) {
	if sec.UseCache {
		lime.Router.Any(relativePath, handlers...).Use(addCache(sec))
	} else {
		lime.Router.Any(relativePath, handlers...)
	}
}

func (lime *LimeServer) PATCH(sec PathSecurity, relativePath string, handlers ...gin.HandlerFunc) {
	if sec.UseCache {
		lime.Router.PATCH(relativePath, handlers...).Use(addCache(sec))
	} else {
		lime.Router.PATCH(relativePath, handlers...)
	}
}

func (lime *LimeServer) HEAD(sec PathSecurity, relativePath string, handlers ...gin.HandlerFunc) {
	if sec.UseCache {
		lime.Router.HEAD(relativePath, handlers...).Use(addCache(sec))
	} else {
		lime.Router.HEAD(relativePath, handlers...)
	}
}

func addCache(sec PathSecurity) gin.HandlerFunc {
	return func(c *gin.Context) {
		if sec.CachePublic {
			c.Header("Cache-Control", "public, max-age:"+sec.MaxAge)
		} else {
			c.Header("Cache-Control", "private, max-age:"+sec.MaxAge)
		}
		c.Header("Last-Modified", time.Now().Format(http.TimeFormat))
		if sec.Expires.String() == "" {
			c.Header("Expires", time.Now().AddDate(0, 1, 0).Format(http.TimeFormat))
		} else {
			c.Header("Expires", sec.Expires.String())
		}
	}
}

func (lime *LimeServer) SetSecurity(waf, usecors, proxy, recovery bool) {
	if waf {
		waf := coraza.NewWaf()
		parser, _ := seclang.NewParser(waf)
		parser.FromString(`#... some rules`)
		lime.Router.Use(corazagin.Coraza(waf))
	}
	if usecors {
		config := cors.DefaultConfig()
		config.AllowOrigins = lime.Security.AllowOrigins
		lime.Router.Use(cors.New(config))
	}
	if proxy {
		lime.Router.Use(secure.New(secure.Config{
			AllowedHosts:          lime.Security.AllowedHosts,
			SSLRedirect:           true,
			IsDevelopment:         false,
			STSSeconds:            lime.Security.STSSeconds,
			STSIncludeSubdomains:  true,
			FrameDeny:             true,
			ContentTypeNosniff:    true,
			BrowserXssFilter:      true,
			ContentSecurityPolicy: "",
			IENoOpen:              true,
			ReferrerPolicy:        "strict-origin-when-cross-origin",
			SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"},
		}))
	}
	if recovery {
		lime.Router.Use(gin.Recovery())
	}
}

func (lime *LimeServer) Start() {
	if err := lime.Server.ListenAndServe(); err != nil {
		logger.Info("There was an error with the http server",
			// Structured context as strongly typed Field values.
			zap.String("error", err.Error()),
		)
	}
}

func (lime *LimeServer) AddCookie(secret string) {
	gob.Register(map[string]interface{}{})
	store := cookie.NewStore([]byte(secret))
	lime.Router.Use(sessions.Sessions("auth-session", store))
}

func (lime *LimeServer) LoadStaticFiles(path string) {
	lime.Router.LoadHTMLGlob(path)
}
