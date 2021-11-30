package router

import (
	"net/http"

	"github.com/galaxy-future/schedulx/api/middleware/authorization"

	"github.com/galaxy-future/schedulx/api/handler"
	"github.com/galaxy-future/schedulx/register/config"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	var router *gin.Engine
	if config.GlobalConfig.DebugMode {
		gin.SetMode(gin.DebugMode)
		router = gin.Default()
		//visit http://0.0.0.0:9090/debug/pprof/
		pprof.Register(router)
	} else {
		gin.SetMode(gin.ReleaseMode)
		router = gin.Default()
	}
	router.GET("/", func(context *gin.Context) {
		context.String(http.StatusOK, "NodeAct")
	})

	v1Api := router.Group("/api/v1/")
	v1Api.Use(authorization.CheckTokenAuth())
	{
		servicePath := v1Api.Group("schedulx/service/")
		{
			h := &handler.Service{}
			servicePath.GET("expand", h.Expand)
			servicePath.GET("shrink", h.Shrink)
			servicePath.GET("detail", h.Detail)
			servicePath.GET("list", h.List)
			servicePath.GET("breathrecord", h.BreathRecord)
			servicePath.POST("update", h.Update)
			servicePath.POST("create", h.Create)
		}
		tmplExpandPath := v1Api.Group("schedulx/template/expand/")
		{
			h := &handler.TmplExpand{}
			tmplExpandPath.POST("create", h.Create) // 创建扩缩容模板
			tmplExpandPath.GET("list", h.List)
			tmplExpandPath.GET("info", h.Info)
			tmplExpandPath.POST("update", h.Update)
			tmplExpandPath.POST("delete", h.Delete)
		}
		taskPath := v1Api.Group("schedulx/task/")
		{
			h := &handler.Task{}
			taskPath.GET("info", h.Info)
			taskPath.GET("instancelist", h.InstanceList)
		}
	}
	return router
}
