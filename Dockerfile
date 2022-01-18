# Builder image
FROM golang:1.17-alpine as builder

RUN echo "https://mirror.tuna.tsinghua.edu.cn/alpine/v3.4/main/" > /etc/apk/repositories && \
    apk add --no-cache \
    wget \
    git

RUN mkdir -p /home/build && \
    mkdir -p /home/schedulx

ARG build_dir=/home/build
ARG api_dir=/home/schedulx

ENV ServiceName=gf.ops.schedulx

WORKDIR $build_dir

COPY . .

# Cache dependencies
ENV GO111MODULE on
ENV GOPROXY https://goproxy.cn,direct

COPY go.mod go.mod
COPY go.sum go.sum
#RUN  go mod download

RUN mkdir -p output/register/conf output/bin

# detect mysql start
COPY wait-for-schedulx.sh output/bin/wait-for-schedulx.sh

RUN find register/conf/ -type f ! -name "*_local.*" | xargs -I{} cp {} output/register/conf/
RUN cp script/run_api.sh output/

RUN CGO_ENABLED=0 GO111MODULE=on go build -o output/bin/${ServiceName}

RUN cp -rf output/* $api_dir

# --------------------------------------------------------------------------------- #
# Executable image
FROM alpine:3.14

RUN echo "https://mirror.tuna.tsinghua.edu.cn/alpine/v3.4/main/" > /etc/apk/repositories

RUN apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone
ENV TZ Asia/Shanghai

RUN apk add --no-cache bash

ENV ServiceName=gf.ops.schedulx

COPY --from=builder /home/schedulx /home/schedulx
WORKDIR /home/schedulx
RUN chmod +x run_api.sh && chmod +x bin/wait-for-schedulx.sh

EXPOSE 9091
CMD ["/home/schedulx/run_api.sh"]