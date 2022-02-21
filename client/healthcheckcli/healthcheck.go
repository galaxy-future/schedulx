package healthCheckcli

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/go-resty/resty/v2"
)

type HealthCheckClient struct {
	httpClient *resty.Client
}

type HealthCheck struct {
	Mode               string `json:"mode'"`
	Path               string `json:"path"`
	Port               int    `json:"port"`
	InitTime           int    `json:"init_time"`
	TimeoutTime        int    `json:"timeout_time"`
	HealthThreshold    int    `json:"health_threshold"`
	UnhealthyThreshold int    `json:"unhealthy_threshold"`
	CheckPeriod        int    `json:"check_period"`
}

var healthCheckCli *HealthCheckClient
var healthCheckOnce sync.Once

func GetHealthCheckXCli(ctx context.Context) *HealthCheckClient {
	healthCheckOnce.Do(func() {
		healthCheckCli = &HealthCheckClient{
			resty.New().SetTimeout(3 * time.Second),
		}
	})
	return healthCheckCli
}

func (c *HealthCheckClient) HealthCheck(ctx context.Context, healthCheck *HealthCheck, instanceInfo *types.InstanceInfo) (err error) {
	url := instanceInfo.IpInner + ":" + strconv.Itoa(healthCheck.Port) + healthCheck.Path
	rr, err := c.httpClient.R().Get(url)
	if err != nil {
		log.Logger.Errorf("url:%s, err:%v", url, err)
		return err
	}
	if rr.RawResponse.StatusCode != http.StatusOK {
		log.Logger.Errorf("url:%s, statusCode:%d", url, rr.RawResponse.StatusCode)
		return errors.New(rr.RawResponse.Status)
	}
	log.Logger.Infof("url:%s, body:%s", url, rr.Body())
	return nil
}
