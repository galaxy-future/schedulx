#!/usr/bin/env bash
# deploy mysql
docker run -d --name schedulx_db -e MYSQL_ROOT_PASSWORD=mtQ8chN2 -e MYSQL_DATABASE=schedulx -e MYSQL_USER=gf -e MYSQL_PASSWORD=db@galaxy-future.com -p 3316:3306 -v $(pwd)/init/mysql:/docker-entrypoint-initdb.d yobasystems/alpine-mariadb:10.5.11
# deploy schedulx
sed "s/127.0.0.1/host.docker.internal/g" $(pwd)/register/conf/config.yml > $(pwd)/register/conf/config.yml.1
sed "s/9090/9099/g" $(pwd)/register/conf/config.yml.1 > $(pwd)/register/conf/config.yml.mac
docker run -d --name schedulx_api --platform linux/amd64 --add-host host.docker.internal:host-gateway -v $(pwd)/register/conf/config.yml.mac:/home/schedulx/register/conf/config.yml -p 9091:9091 galaxyfuture/schedulx-api bin/wait-for-schedulx.sh
