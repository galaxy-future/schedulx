# **SchedulX**

[![Go Report Card](https://goreportcard.com/badge/github.com/galaxy-future/BridgX)](https://goreportcard.com/report/github.com/galaxy-future/BridgX) &nbsp;
[![CodeFactor](https://www.codefactor.io/repository/github/galaxy-future/bridgx/badge)](https://www.codefactor.io/repository/github/galaxy-future/bridgx)


语言(language)
----

中文 | [English](https://github.com/galaxy-future/schedulx/blob/master/docs/EN-README.md)

简介
--------

SchedulX 是基于开源 BridgX 项目的云原生服务编排和部署解决方案，目标是让开发者在 BridgX 获取的计算资源上进行服务编排和部署。

它具有如下关键特性:

1、具备结合动态扩缩容特性进行服务部署的能力；

2、一个平台统一管理在不同云平台上的服务操作；

3、简洁易用，轻松上手；


安装部署
--------

1、配置要求

为了系统稳定运行，建议系统型号2核4G内存；SchedulX已经在Linux系统以及macOS系统进行了安装和测试。

2、环境依赖

SchedulX依赖BridgX和BridgX_FE，请根据[安装指南](hhttps://github.com/galaxy-future/BridgX/blob/dev/README.md)安装BridgX和BridgX_FE。要求内网部署环境，能够跟云厂商vpc连通。

3、安装步骤

* (1)源码下载
  - 后端工程：
  > `git clone git@github.com:galaxy-future/SchedulX.git`
  - 代码下载完成后，修改配置文件`register/conf/config.yml`，填入云账号的Accesskey、Secret及Region信息。

* (2)macOS系统部署
  - 后端部署,在SchedulX目录下运行
    > `make docker-run-mac`

* (3)Linux安装部署
  - 1）针对使用者
    - 后端部署,在SchedulX目录下运行,
      > `make docker-run-linux`

  - 2）针对开发者
    - 后端部署
      - SchedulX依赖mysql组件，
           - 如果使用内置的mysql，则进入SchedulX根目录，则使用以下命令：            
             > `docker-compose up -d`    //启动SchedulX <br>
             > `docker-compose down`    //停止SchedulX  <br>
           - 如果已经有了外部的mysql服务，则可以到 `cd conf` 下修改对应的ip和port配置信息,然后进入SchedulX的根目录，使用以下命令:
             > `docker-compose up -d schedulx`   //启动schedulx服务 <br>
             > `docker-compose down`     //停止SchedulX服务

4、前端界面化操作

如果需要使用web前端进行操作，请下载安装[ComandX](https://github.com/galaxy-future/comandx/blob/main/README.md)

行为准则
------
[贡献者公约](https://github.com/galaxy-future/schedulx/blob/master/CODE_OF_CONDUCT.md)

授权
-----

SchedulX 使用[Apache License 2.0](https://github.com/galaxy-future/schedulx/blob/master/README.md)授权协议进行授权
