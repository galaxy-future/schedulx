package router

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/galaxy-future/schedulx/api/handler"
	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/register/config"
	"github.com/galaxy-future/schedulx/register/config/client"
	"github.com/galaxy-future/schedulx/register/config/log"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
)

const (
	authToken = "xxxx"
)

func TestMain(m *testing.M) {
	os.Setenv("env", "local")
	config.Init("../../register/conf/config.yml")
	log.Init()
	log.Logger.Info("TestMain Start ...")
	{

		client.Init()
	}
	m.Run()
	log.Logger.Info("TestMain End ...")
}
func TestService_Create(t *testing.T) {
	type args struct {
		Method string
		Url    string
		Param  *handler.ServiceCreateHttpRequest
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ServiceCreate pier",
			args: args{
				Method: "POST",
				Url:    "/api/v1/schedulx/service/create",
				Param: &handler.ServiceCreateHttpRequest{
					ServiceInfo: &types.ServiceInfo{
						ServiceName: "gf.dtexpress.pier",
						Description: "数据仓库 pier 项目",
						Language:    "GO",
					},
				},
			},
		},
		{
			name: "ServiceCreate downloader",
			args: args{
				Method: "POST",
				Url:    "/api/v1/schedulx/service/create",
				Param: &handler.ServiceCreateHttpRequest{
					ServiceInfo: &types.ServiceInfo{
						ServiceName: "gf.dtexpress.downloader",
						Description: "数据仓库 downloader 项目",
						Language:    "GO",
					},
				},
			},
		},
		{
			name: "ServiceCreate net_detector",
			args: args{
				Method: "POST",
				Url:    "/api/v1/schedulx/service/create",
				Param: &handler.ServiceCreateHttpRequest{
					ServiceInfo: &types.ServiceInfo{
						ServiceName: "gf.rd.net_detector",
						Description: "网络基调 net_detector 项目",
						Language:    "GO",
					},
				},
			},
		},
	}
	r := Init()
	w := httptest.NewRecorder()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonByte, _ := jsoniter.Marshal(tt.args.Param)
			req, _ := http.NewRequest(tt.args.Method, tt.args.Url, bytes.NewReader(jsonByte))
			req.Header.Set("Authorization", "Bearer "+authToken)
			r.ServeHTTP(w, req)
			t.Log(w.Body.String())
			assert.Equal(t, 200, w.Code)
		})
	}
}

func TestService_Expand(t *testing.T) {
	type input struct {
		ServiceClusterId int64 `json:"service_cluster_id"`
		Count            int64 `json:"count"`
	}
	type args struct {
		Method string
		Url    string
		Param  *input
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ServiceExpandT",
			args: args{
				Method: "GET",
				Url:    "/api/v1/schedulx/service/expand?service_cluster_id=93&count=2",
			},
		},
	}

	r := Init()
	w := httptest.NewRecorder()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonByte, _ := jsoniter.Marshal(tt.args.Param)
			req, _ := http.NewRequest(tt.args.Method, tt.args.Url, bytes.NewReader(jsonByte))
			req.Header.Set("Authorization", "Bearer "+authToken)
			r.ServeHTTP(w, req)
			t.Log(w.Body.String())
			assert.Equal(t, 200, w.Code)
		})
	}
}

func TestService_Shrink(t *testing.T) {
	type args struct {
		Method string
		Url    string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ServiceShrink",
			args: args{
				Method: "GET",
				Url:    "/api/v1/schedulx/service/shrink?service_cluster_id=57&count=2&exec_type=manual",
			},
		},
	}

	r := Init()
	w := httptest.NewRecorder()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.args.Method, tt.args.Url, nil)
			req.Header.Set("Authorization", "Bearer "+authToken)
			r.ServeHTTP(w, req)
			t.Log(w.Body.String())
			assert.Equal(t, 200, w.Code)
		})
	}
}
