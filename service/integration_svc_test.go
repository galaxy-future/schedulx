package service

import (
	"context"
	"testing"

	"github.com/galaxy-future/schedulx/register/config"
	"github.com/galaxy-future/schedulx/register/config/client"
	"github.com/galaxy-future/schedulx/register/config/log"
)

func InitTest() {
	config.Init("../register/conf/config.yml.local")
	log.Init()
	client.Init()
}

func TestGetZadigHost(t *testing.T) {
	InitTest()
	s := GetIntegrationService()
	t.Logf(s.GetZadigHost(context.Background()))
}

func Test_generateZadigToken(t *testing.T) {
	InitTest()

	tests := []struct {
		name string
		want string
	}{
		{"TEST1",
			""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateZadigToken(); got != tt.want {
				t.Errorf("generateZadigToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
