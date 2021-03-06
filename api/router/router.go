package router

import (
	"net/http"

	"github.com/galaxy-future/schedulx/api/handler"
	"github.com/galaxy-future/schedulx/api/middleware/authorization"
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
		context.String(http.StatusOK, "SchedulX")
	})

	v1Api := router.Group("/api/v1/")
	v1Api.Use(authorization.CheckTokenAuth())
	{
		servicePath := v1Api.Group("schedulx/service/")
		{
			h := &handler.Service{}
			servicePath.GET("expand", h.Expand)
			servicePath.GET("shrink", h.Shrink)
			servicePath.GET("deploy", h.Deploy)
			servicePath.GET("detail", h.Detail)
			servicePath.GET("list", h.List)
			servicePath.GET("scheduling", h.Scheduling)
			servicePath.GET("cluster_list", h.ClusterList)
			servicePath.GET("breathrecord", h.BreathRecord)
			servicePath.POST("update", h.Update)
			servicePath.POST("create", h.Create)
			servicePath.POST("delete", h.Delete)
			servicePath.GET("zadig/workflow/list", h.GetWorkflows)
			servicePath.GET("zadig/artifact/list", h.GetWorkflowTasks)
		}
		instancePath := v1Api.Group("schedulx/instance/")
		{
			h := &handler.Instance{}
			instancePath.GET("count", h.Count)
			instancePath.GET("service", h.InstanceServiceByIp)
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
		deployExpandPath := v1Api.Group("schedulx/template/deploy/")
		{
			h := &handler.TmplDeploy{}
			deployExpandPath.GET("list", h.List)
		}
		taskPath := v1Api.Group("schedulx/task/")
		{
			h := &handler.Task{}
			taskPath.GET("info", h.Info)
			taskPath.GET("deploy/detail", h.GetDeployDetail)
			taskPath.GET("instancelist", h.InstanceList)
		}
		integrationPath := v1Api.Group("schedulx/integration/")
		{
			h := &handler.Integration{}
			integrationPath.GET("list", h.List)
			integrationPath.POST("create", h.Create)
			integrationPath.POST("delete", h.Delete)
		}
	}
	return router
}
