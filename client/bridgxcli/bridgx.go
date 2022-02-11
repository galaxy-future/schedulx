package bridgxcli

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/spf13/cast"

	"github.com/galaxy-future/schedulx/pkg/tool"
	"github.com/go-resty/resty/v2"

	"github.com/galaxy-future/schedulx/client"
	"github.com/galaxy-future/schedulx/pkg/bridgx"
	"github.com/galaxy-future/schedulx/register/config"
	"github.com/galaxy-future/schedulx/register/config/log"
)

type BridgXClient struct {
	httpClient *resty.Client
}

var bridgXCli *BridgXClient
var bridgXOnce sync.Once

type ClusterExpandReq struct {
	TaskName    string
	ClusterName string
	Count       int64
}

type ClusterExpandResp struct {
	client.HttpResp
	Data int64 `json:"data"`
}

type ClusterShrinkReq struct {
	TaskName    string
	ClusterName string
	Ips         []string
	Count       int64
}

type ClusterShrinkResp struct {
	client.HttpResp
	Data int64 `json:"data"`
}

type TaskDescribeReq struct {
	TaskId int64
}

type TaskDescribeResp struct {
	client.HttpResp
	Data *bridgx.TaskDescribe `json:"data"`
}

type TaskInstancesReq struct {
	TaskId         int64
	InstanceStatus bridgx.InstStatus
	PageNum        int64
	PageSize       int64
}

type ClusterInstancesReq struct {
	InstanceStatus bridgx.InstStatus
	ClusterName    string
	PageNum        int64
	PageSize       int64
}

type ClusterInstancesResp struct {
	client.HttpResp
	Data *bridgx.InstancesData `json:"data"`
}

type TaskInstancesResp struct {
	client.HttpResp
	Data *bridgx.TaskInstancesData `json:"data"`
}

type GetClusterByNameReq struct {
	ClusterName string
}

type GetClusterByNameResp struct {
	client.HttpResp
	Data *bridgx.ClusterInfo `json:"data"`
}

type ClusterInstanceStatResp struct {
	client.HttpResp
	Data *bridgx.ClusterInstanceStat `json:"data"`
}

func GetBridgXCli(ctx context.Context) *BridgXClient {
	bridgXOnce.Do(func() {
		bridgXCli = &BridgXClient{
			resty.New().SetTimeout(2 * time.Second),
		}
	})
	return bridgXCli
}

func (c *BridgXClient) entryLog(ctx context.Context, method string, req interface{}) {
	log.Logger.Infof("entry log | method[%s] | req:%s", method, tool.ToJson(req))
}

func (c *BridgXClient) exitLog(ctx context.Context, method string, req, resp interface{}, err error) {
	log.Logger.Infof("exit log | method[%s] | req:%s | resp:%s | err:%v", method, tool.ToJson(req), tool.ToJson(resp), err)
}

func (c *BridgXClient) ClusterExpand(ctx context.Context, cliReq *ClusterExpandReq) (resp *ClusterExpandResp, err error) {
	resp = &ClusterExpandResp{}
	if cliReq.TaskName == "" {
		err = client.ErrParamsMissing
		log.Logger.Error(err, ":task_name")
		return nil, err
	}
	if cliReq.ClusterName == "" {
		err = client.ErrParamsMissing
		log.Logger.Error(err, ":cluster_name")
		return nil, err
	}
	if cliReq.Count == 0 {
		err = client.ErrParamsMissing
		log.Logger.Error(err, ":count")
		return nil, err
	}
	params := map[string]interface{}{
		"task_name":    cliReq.TaskName,
		"cluster_name": cliReq.ClusterName,
		"count":        cliReq.Count,
	}
	url := c.genUrl(clusterExpandUrl)
	authToken := cast.ToString(ctx.Value(config.GlobalConfig.JwtToken.BindContextKeyName))

	ret, err := c.httpClient.R().SetBody(params).SetResult(resp).SetError(resp).SetAuthToken(authToken).Post(url)
	log.Logger.Infof("ret:%s", ret.Body())
	log.Logger.Infof("url:%+v", url)
	log.Logger.Infof("params:%+v", tool.ToJson(params))
	log.Logger.Infof("resp:%+v", tool.ToJson(resp))
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	if resp.Code != http.StatusOK {
		err = fmt.Errorf("http code:%v | msg:%v", resp.Code, resp.Msg)
		log.Logger.Error(err)
		return nil, err
	}
	return resp, nil
}

func (c *BridgXClient) ClusterShrink(ctx context.Context, cliReq *ClusterShrinkReq) (resp *ClusterShrinkResp, err error) {
	resp = &ClusterShrinkResp{}
	if cliReq.TaskName == "" {
		err = client.ErrParamsMissing
		log.Logger.Error(err, ":task_name")
		return nil, err
	}
	if cliReq.ClusterName == "" {
		err = client.ErrParamsMissing
		log.Logger.Error(err, ":cluster_name")
		return nil, err
	}
	if len(cliReq.Ips) == 0 && cliReq.Count == 0 { // ips 和 count 至少有一个
		err = client.ErrParamsMissing
		log.Logger.Error(err, ":ips || count")
		return nil, err
	}
	params := map[string]interface{}{
		"task_name":    cliReq.TaskName,
		"cluster_name": cliReq.ClusterName,
		"ips":          cliReq.Ips,
		"count":        cliReq.Count,
	}
	url := c.genUrl(clusterShrinkUrl)
	authToken := cast.ToString(ctx.Value(config.GlobalConfig.JwtToken.BindContextKeyName))
	rr, err := c.httpClient.R().SetBody(params).SetResult(resp).SetError(resp).SetAuthToken(authToken).Post(url)
	log.Logger.Infof("rr:%s", rr.Body())
	log.Logger.Infof("url:%v", url)
	log.Logger.Infof("params:%+v", tool.ToJson(params))
	log.Logger.Infof("resp:%v", tool.ToJson(resp))
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	if resp.Code != http.StatusOK {
		err = fmt.Errorf("http code:%v | msg:%v", resp.Code, resp.Msg)
		log.Logger.Error(err)
		return nil, err
	}
	return resp, err
}

func (c *BridgXClient) TaskDescribe(ctx context.Context, cliRq *TaskDescribeReq) (resp *TaskDescribeResp, err error) {
	resp = &TaskDescribeResp{}
	if cliRq.TaskId == 0 {
		err = client.ErrParamsMissing
		log.Logger.Error(err, ":task_id")
		return nil, err
	}
	params := map[string]string{
		"task_id": tool.Interface2String(cliRq.TaskId),
	}
	url := c.genUrl(taskDescribeUrl)
	authToken := cast.ToString(ctx.Value(config.GlobalConfig.JwtToken.BindContextKeyName))
	rr, err := c.httpClient.R().SetQueryParams(params).SetResult(resp).SetError(resp).SetAuthToken(authToken).Get(url)
	log.Logger.Infof("rr:%s", rr.Body())
	log.Logger.Infof("url:%v", url)
	log.Logger.Infof("params:%+v", tool.ToJson(params))
	log.Logger.Infof("resp:%v", tool.ToJson(resp))
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	if resp.Code != http.StatusOK {
		err = fmt.Errorf("http code:%v | msg:%v", resp.Code, resp.Msg)
		log.Logger.Error(err)
		return nil, err
	}
	return resp, err
}

func (c *BridgXClient) TaskInstances(ctx context.Context, cliReq *TaskInstancesReq) (resp *TaskInstancesResp, err error) {
	resp = &TaskInstancesResp{}
	if cliReq.TaskId == 0 {
		err = client.ErrParamsMissing
		log.Logger.Error(err, ":task_id")
		return nil, err
	}
	if cliReq.InstanceStatus == "" {
		err = client.ErrParamsMissing
		log.Logger.Error(err, ":instance_status")
		return nil, err
	}
	params := map[string]string{
		"task_id":         tool.Interface2String(cliReq.TaskId),
		"instance_status": tool.Interface2String(cliReq.InstanceStatus),
		"page_number":     tool.Interface2String(cliReq.PageNum),
		"page_size":       tool.Interface2String(cliReq.PageSize),
	}
	url := c.genUrl(taskInstancesUrl)
	authToken := cast.ToString(ctx.Value(config.GlobalConfig.JwtToken.BindContextKeyName))
	rr, err := c.httpClient.R().SetQueryParams(params).SetResult(resp).SetError(resp).SetAuthToken(authToken).Get(url)
	log.Logger.Infof("rr:%s", rr.Body())
	log.Logger.Infof("url:%v", url)
	log.Logger.Infof("params:%+v", tool.ToJson(params))
	log.Logger.Infof("resp:%v", tool.ToJson(resp))
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	if resp.Code != http.StatusOK {
		err = fmt.Errorf("http code:%v | msg:%v", resp.Code, resp.Msg)
		log.Logger.Error(err)
		return nil, err
	}
	return resp, err
}

func (c *BridgXClient) ClusterInstances(ctx context.Context, cliReq *ClusterInstancesReq) (resp *ClusterInstancesResp, err error) {
	resp = &ClusterInstancesResp{}

	if cliReq.InstanceStatus == "" {
		err = client.ErrParamsMissing
		log.Logger.Error(err, ":instance_status")
		return nil, err
	}
	params := map[string]string{
		"status":       string(cliReq.InstanceStatus),
		"cluster_name": cliReq.ClusterName,
		"page_number":  tool.Interface2String(cliReq.PageNum),
		"page_size":    tool.Interface2String(cliReq.PageSize),
	}
	url := c.genUrl(clusterInstancesUrl)
	authToken := cast.ToString(ctx.Value(config.GlobalConfig.JwtToken.BindContextKeyName))
	rr, err := c.httpClient.R().SetQueryParams(params).SetResult(resp).SetError(resp).SetAuthToken(authToken).Get(url)
	log.Logger.Infof("rr:%s", rr.Body())
	log.Logger.Infof("url:%v", url)
	log.Logger.Infof("params:%+v", tool.ToJson(params))
	log.Logger.Infof("resp:%v", tool.ToJson(resp))
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	if resp.Code != http.StatusOK {
		err = fmt.Errorf("http code:%v | msg:%v", resp.Code, resp.Msg)
		log.Logger.Error(err)
		return nil, err
	}
	return resp, err
}

func (c *BridgXClient) GetClusterByName(ctx context.Context, cliReq *GetClusterByNameReq) (resp *GetClusterByNameResp, err error) {
	resp = &GetClusterByNameResp{}
	if cliReq.ClusterName == "" {
		err = client.ErrParamsMissing
		log.Logger.Error(err, ":cluster_name")
		return nil, err
	}
	url := tool.StrAppend(c.genUrl(clusterGetByNameUrl), "/", cliReq.ClusterName)
	authToken := cast.ToString(ctx.Value(config.GlobalConfig.JwtToken.BindContextKeyName))
	rr, err := c.httpClient.R().SetResult(resp).SetError(resp).SetAuthToken(authToken).Get(url)
	log.Logger.Infof("rr:%s", rr.Body())
	log.Logger.Infof("url:%v", url)
	log.Logger.Infof("resp:%v", tool.ToJson(resp))
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	if resp.Code != http.StatusOK {
		err = fmt.Errorf("http code:%v | msg:%v", resp.Code, resp.Msg)
		log.Logger.Error(err)
		return nil, err
	}
	return resp, nil
}

func (c *BridgXClient) ClusterInstanceStat(ctx context.Context, bridgxClusterName string) (resp *ClusterInstanceStatResp, err error) {
	resp = &ClusterInstanceStatResp{}
	if bridgxClusterName == "" {
		err = client.ErrParamsMissing
		log.Logger.Error(err, ":cluster_name")
		return nil, err
	}
	params := map[string]string{
		"cluster_name": bridgxClusterName,
	}
	url := c.genUrl(clusterInstanceStat)
	authToken := cast.ToString(ctx.Value(config.GlobalConfig.JwtToken.BindContextKeyName))
	rr, err := c.httpClient.R().SetQueryParams(params).SetResult(resp).SetError(resp).SetAuthToken(authToken).Get(url)
	log.Logger.Infof("rr:%s", rr.Body())
	log.Logger.Infof("url:%v", url)
	log.Logger.Infof("resp:%v", tool.ToJson(resp))
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	if resp.Code != http.StatusOK {
		err = fmt.Errorf("http code:%v | msg:%v", resp.Code, resp.Msg)
		log.Logger.Error(err)
		return nil, err
	}
	return resp, nil
}

func (c *BridgXClient) genUrl(url string) string {
	return config.GlobalConfig.BridgXHost + url
}
