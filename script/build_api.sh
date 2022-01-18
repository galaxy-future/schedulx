#!/bin/bash
RUN_NAME="gf.ops.schedulx"
mkdir -p output/register/conf output/bin
sudo mkdir -p /var/log/app/${RUN_NAME}
sudo mkdir -p /home/tiger/containers/${RUN_NAME}/log/

find register/conf/ -type f ! -name "*.local*" -print0 | xargs -0 -I{} cp {} output/register/conf/

cp script/run_api.* output/

go mod tidy
go fmt ./...
go vet ./...
export GO111MODULE="on"
go build -o output/bin/${RUN_NAME}
