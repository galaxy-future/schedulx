package zadigcli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
	"net/http"
	"sync"
	"time"
)

type ZadigClient struct {
	httpClient *resty.Client
}

var zadigCli *ZadigClient
var zadigOnce sync.Once

type WorkflowsReq struct {
	ZadigHost   string
	ZadigToken  string
	ProjectName string
}

type WorkflowResp struct {
	WorkflowName string `json:"name"`
}

type WorkflowTasksReq struct {
	ZadigHost    string
	ZadigToken   string
	ProjectName  string
	WorkflowName string
	FileType     string
	Limit        int
	Skip         int
}

type WorkflowTaskResp struct {
	TaskId    int    `json:"task_id"`
	ImageName string `json:"image_name"`
}

type S3StorageResp struct {
	Bucket    string `json:"bucket"`
	Endpoint  string `json:"endpoint"`
	Subfolder string `json:"subfolder"`
}

type OriginalWorkflowTasksResp struct {
	Data  []OriginalWorkflowTaskResp `json:"data"`
	Total int                        `json:"total"`
}

type OriginalWorkflowTaskResp struct {
	TaskId        int      `json:"task_id"`
	BuildServices []string `json:"build_services"`
	Stages        []Stages `json:"stages"`
	Status        string   `json:"status"`
}

type Stages struct {
	Type     string      `json:"type"`
	SubTasks interface{} `json:"sub_tasks"`
	Status   string      `json:"status"`
}

func GetZadigXCli(ctx context.Context) *ZadigClient {
	zadigOnce.Do(func() {
		zadigCli = &ZadigClient{
			resty.New().SetTimeout(3 * time.Second),
		}
	})
	return zadigCli
}

func (c *ZadigClient) GetWorkflows(ctx context.Context, cliReq WorkflowsReq) (resp []*WorkflowResp, err error) {
	resp = make([]*WorkflowResp, 0)
	url := cliReq.ZadigHost + listWorkflowUrl
	params := map[string]string{
		"projectName": cliReq.ProjectName,
	}
	rr, err := c.httpClient.R().SetQueryParams(params).SetResult(&resp).SetError(&resp).SetAuthToken(cliReq.ZadigToken).Get(url)
	if err != nil {
		log.Logger.Errorf("url:%s, err:%v", url, err)
		return nil, err
	}
	if rr.RawResponse.StatusCode != http.StatusOK {
		log.Logger.Errorf("url:%s, statusCode:%d", url, rr.RawResponse.StatusCode)
		return nil, errors.New(rr.RawResponse.Status)
	}
	log.Logger.Infof("url:%s, body:%s", url, rr.Body())
	return resp, nil
}

func (c *ZadigClient) GetWorkflowTasks(ctx context.Context, cliReq WorkflowTasksReq) (total int, resp []*WorkflowTaskResp, err error) {
	resp = make([]*WorkflowTaskResp, 0)
	originalResp := OriginalWorkflowTasksResp{}
	url := cliReq.ZadigHost + fmt.Sprintf(listWorkflowTasksUrl, cliReq.Limit, cliReq.Skip, cliReq.WorkflowName)
	params := map[string]string{
		"projectName": cliReq.ProjectName,
		"queryType":   "taskStatus",
		"filters":     "passed",
	}
	rr, err := c.httpClient.R().SetQueryParams(params).SetResult(&originalResp).SetError(&originalResp).SetAuthToken(cliReq.ZadigToken).Get(url)
	if err != nil {
		log.Logger.Errorf("url:%s, err:%v", url, err)
		return 0, nil, err
	}
	if rr.RawResponse.StatusCode != http.StatusOK {
		log.Logger.Errorf("url:%s, statusCode:%d", url, rr.RawResponse.StatusCode)
		return 0, nil, errors.New(rr.RawResponse.Status)
	}
	log.Logger.Infof("url:%s, body:%s", url, rr.Body())
	if cliReq.FileType == "file" {
		domain, err := c.getS3StorageDomain(cliReq.ZadigHost, cliReq.ZadigToken)
		if err != nil {
			return 0, nil, err
		}
		for _, task := range originalResp.Data {
			for _, stage := range task.Stages {
				if stage.Type == "distribute2kodo" {
					subTaskByte, _ := json.Marshal(stage.SubTasks)
					remoteFileKey := gjson.Get(string(subTaskByte), task.BuildServices[0]+".remote_file_key")
					resp = append(resp, &WorkflowTaskResp{
						TaskId:    task.TaskId,
						ImageName: domain + remoteFileKey.Str,
					})
				}
			}
		}
	} else {
		for _, task := range originalResp.Data {
			for _, stage := range task.Stages {
				if stage.Type == "buildv2" {
					subTaskByte, _ := json.Marshal(stage.SubTasks)
					imageName := gjson.Get(string(subTaskByte), task.BuildServices[0]+".docker_build_status.image_name")
					resp = append(resp, &WorkflowTaskResp{
						TaskId:    task.TaskId,
						ImageName: imageName.Str,
					})
				}
			}
		}
	}
	return originalResp.Total, resp, nil
}

func (c *ZadigClient) getS3StorageDomain(zadigHost, zadigToken string) (string, error) {
	s3StotageList := make([]S3StorageResp, 0)
	url := zadigHost + listS3storageUrl
	rr, err := c.httpClient.R().SetResult(&s3StotageList).SetError(&s3StotageList).SetAuthToken(zadigToken).Get(url)
	if err != nil {
		log.Logger.Errorf("url:%s, err:%v", url, err)
		return "", err
	}
	log.Logger.Infof("url:%s, body:%s", url, rr.Body())
	if len(s3StotageList) == 0 {
		return "", errors.New("s3Stotage is not configured")
	}
	return fmt.Sprintf("https://%s.%s/", s3StotageList[0].Bucket, s3StotageList[0].Endpoint), nil
}
