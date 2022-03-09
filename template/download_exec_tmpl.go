package template

import (
	"bytes"
	"fmt"
	"text/template"
)

type DownloadParams struct {
	DownloadFileUrl string `json:"download_file_url"`
	DeployFilePath  string `json:"deploy_file_path"`
	DeployFileName  string `json:"deploy_file_name"`
}

const downloadExecCmd = `
#!/bin/bash
mkdir -p {{.DeployFilePath}}
wget {{.DownloadFileUrl}} -O {{.DeployFilePath}}/{{.DeployFileName}}`

func GetDownloadExecCmd() string {
	return downloadExecCmd
}

func ParseDownloadExecCmd(params DownloadParams) (string, error) {
	if params.DownloadFileUrl == "" || params.DeployFileName == "" || params.DeployFilePath == "" {
		return "", fmt.Errorf("download params empty")
	}
	if params.DeployFilePath[len(params.DeployFilePath)-1] == '/' {
		params.DeployFilePath = params.DeployFilePath[:len(params.DeployFilePath)-1]
	}
	tmpl, _ := template.New("download").Parse(downloadExecCmd)
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, params)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
