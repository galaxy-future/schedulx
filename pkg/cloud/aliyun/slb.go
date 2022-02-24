package aliyun

import (
	"encoding/json"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
)

type SLB struct {
	loadBalancerId string
	regionId       string
	client         *slb.Client
}

type Server struct {
	ServerId string `json:"ServerId"`
	Weight   string `json:"Weight"`
	Type     string `json:"Type"` //增加机器默认type为ecs
	ServerIp string `json:"ServerIp"`
}

func NewSLB(region, accessKey, accessSecret, loadBalancerId string) (*SLB, error) {
	client, err := slb.NewClientWithAccessKey(region, accessKey, accessSecret)
	if err != nil {
		return nil, err
	}

	return &SLB{
		loadBalancerId: loadBalancerId,
		regionId:       region,
		client:         client,
	}, nil
}

func (sb *SLB) CreateServer(servers []Server) error {
	request := slb.CreateAddBackendServersRequest()
	request.LoadBalancerId = sb.loadBalancerId
	request.RegionId = sb.regionId
	serversJson, err := json.Marshal(servers)
	if err != nil {
		return err
	}
	request.BackendServers = string(serversJson)

	_, err = sb.client.AddBackendServers(request)
	return err
}

func (sb *SLB) RemoveServer(servers []Server) error {
	request := slb.CreateRemoveBackendServersRequest()
	request.LoadBalancerId = sb.loadBalancerId
	request.RegionId = sb.regionId
	serversJson, err := json.Marshal(servers)
	if err != nil {
		return err
	}
	request.BackendServers = string(serversJson)

	_, err = sb.client.RemoveBackendServers(request)
	return err
}
