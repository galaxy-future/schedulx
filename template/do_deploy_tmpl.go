package template

import (
	"bytes"
	"strings"
	"text/template"
)

type DeployParams struct {
	EnvVariables []string `json:"env_variables"` //KEY=VAL
	RawCmd       string   `json:"raw_cmd"`
}

const deployCmd = `
#!/bin/bash
{{Unfold .EnvVariables}}
{{.RawCmd}}
`

func GetDeployCmd() string {
	return deployCmd
}

func ParseDeployCmd(params DeployParams) (string, error) {
	tmpl, _ := template.New("deploy").Funcs(template.FuncMap{
		"Unfold": func(envs []string) string {
			for i, env := range envs {
				if strings.TrimSpace(env) == "" {
					continue
				}
				envs[i] = "export " + env
			}
			return strings.Join(envs, "\n")
		},
	}).Parse(deployCmd)
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, params)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
