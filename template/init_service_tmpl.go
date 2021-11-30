package template

import (
	"bytes"
	"text/template"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/pkg/tool"
)

type Data struct {
	DockerRun string
	Params    *types.ParamsServiceEnv
}

const initServiceCmd = `
#!/bin/bash
HarborUrl={{pickDomainFromUrl .Params.ImageUrl}}
ImageUrl={{.Params.ImageUrl}}
PSM={{.Params.ServiceName}}
Port={{.Params.Port}}
{{if eq .Params.ImageStorageType "acr"}}
echo "dock login" >> /root/result.log
docker login --username={{.Params.Account}} --password={{.Params.Password}} $HarborUrl
{{else}}
echo 'INSECURE_REGISTRY="--insecure-registry docker.io --insecure-registry '$HarborUrl'"' >> /etc/sysconfig/docker
{{end}}
service docker start >> /root/result.log

docker pull $ImageUrl >> /root/result.log
{{.DockerRun}} --name $PSM $ImageUrl >> /root/result.log

{{if .Params.Port}}
#检查镜像服务是否启动
TIMES=5
EXIST=0
for((i=0;i<$TIMES;i++));
do
        echo "check $Port time $i ..." >> /root/result.log
       res=$(netstat -an | grep LISTEN | grep -w $Port)
        if [ "" != "$res" ]; then
                EXIST=1
                break
        fi
        sleep 2
done

if [ "1" == "$EXIST" ]
then
  echo "success" >> /root/result.log
  echo "success"
else
  echo "service start error" >> /root/result.log
  echo "service start error"
fi
{{else}}
echo "success" >> /root/result.log
echo "success"
{{end}}
`

func GetInitServiceCmd(params *types.ParamsServiceEnv, dockerRun string) (string, error) {
	pass, err := tool.AesDecrypt(params.Password, []byte(params.Account))
	if err != nil {
		return "", err
	}
	pwd := string(pass)
	data := Data{
		DockerRun: dockerRun,
		Params: &types.ParamsServiceEnv{
			ImageStorageType: params.ImageStorageType,
			ImageUrl:         params.ImageUrl,
			ServiceName:      params.ServiceName,
			Port:             params.Port,
			Account:          params.Account,
			Password:         pwd,
		},
	}
	fm := template.FuncMap{
		"pickDomainFromUrl": tool.PickDomainFromUrl,
	}
	tmpl, _ := template.New("service").Funcs(fm).Parse(initServiceCmd)
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
