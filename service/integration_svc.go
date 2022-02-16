package service

import (
	"context"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/galaxy-future/schedulx/register/constant"
	"github.com/galaxy-future/schedulx/repository"
)

type IntegrationService struct {
	ZadigToken string
	HttpClient *resty.Client
}

func init() {
	go refreshZadigToken()
}

var integrationService *IntegrationService
var integrationServiceOnce sync.Once

func GetIntegrationService() *IntegrationService {
	integrationServiceOnce.Do(func() {
		integrationService = &IntegrationService{
			ZadigToken: "",
			HttpClient: resty.New().SetTimeout(2 * time.Second),
		}
	})
	return integrationService
}

func (s *IntegrationService) GetZadigHost(ctx context.Context) string {
	zadig, err := repository.GetIntegrationInstance().GetOneByType(ctx, constant.Zadig)
	if err != nil {
		return constant.DefaultLocalHost
	}
	return zadig.Host
}

func (s *IntegrationService) GetZadigToken() string {
	if s.ZadigToken != "" {
		return s.ZadigToken
	}
	return generateZadigToken()
}

func refreshZadigToken() {
	time.AfterFunc(5*time.Minute, func() {
		generateZadigToken()
	})
}

func generateZadigToken() string {
	s := GetIntegrationService()
	zadig, err := repository.GetIntegrationInstance().GetOneByType(context.Background(), constant.Zadig)
	if err != nil {
		log.Logger.Errorf("generateZadigToken error:%v", err)
		//has error, return old token
		return s.ZadigToken
	}
	params := map[string]string{
		"account":  zadig.Account,
		"password": zadig.Password,
	}
	resp := &ZadigResp{}
	res, err := s.HttpClient.R().SetBody(params).SetResult(resp).SetError(resp).Post(zadigLoginUrl(zadig.Host))
	log.Logger.Infof("zadig login response:%s", string(res.Body()))
	if err != nil {
		log.Logger.Errorf("login zadig error:%v", err)
		return s.ZadigToken
	}
	s.ZadigToken = resp.Token
	return s.ZadigToken
}

type ZadigResp struct {
	Token string `json:"token"`
}

func zadigLoginUrl(host string) string {
	if host[len(host)-1] == '/' {
		host = host[:len(host)-1]
	}
	return host + "/api/v1/login"
}
