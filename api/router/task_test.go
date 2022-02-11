package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTask_Instancelist(t *testing.T) {
	type args struct {
		Method string
		Url    string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "TaskInstanceList",
			args: args{
				Method: "GET",
				Url:    "/api/v1/schedulx/task/instancelist?task_id=2777",
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

func TestTask_Info(t *testing.T) {
	type args struct {
		Method string
		Url    string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "TaskInfo",
			args: args{
				Method: "GET",
				Url:    "/api/v1/schedulx/task/info?task_id=237",
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
