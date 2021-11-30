package service

import (
	"context"
	"strings"
	"time"

	"github.com/galaxy-future/schedulx/register/config"

	"github.com/galaxy-future/schedulx/pkg/goph"
	"github.com/galaxy-future/schedulx/register/config/log"
)

// RemoteCmdExec 复制 inScript 到 远端 OutScript 并执行 | todo 参数结构化
func RemoteCmdExec(ctx context.Context, localCmd string, remoteScript string, ip, uname, pwd string) ([]byte, error) {
	var err error
	var data []byte
	uname = strings.TrimSpace(uname)
	pwd = strings.TrimSpace(pwd)
	var SSHClient *goph.Client
	wt := 5 * time.Second
	errCnt := 0
	for {
		if errCnt >= 3 {
			return nil, err
		}
		log.Logger.Info("new goph start")
		SSHClient, err = goph.NewUnknown(uname, ip, goph.Password(pwd))
		if err != nil {
			errCnt++
			log.Logger.Error("new goph,", err)
			log.Logger.Infof("%s 后重试", wt)
			time.Sleep(wt)
			continue
		}
		break
	}

	log.Logger.Infof("cmd:%s, remoteScript:%s", localCmd, remoteScript)
	cmd := localCmd
	log.Logger.Infof("cmd: %s", cmd)
	data, err = SSHClient.Run(cmd)
	if err != nil {
		log.Logger.Error("ssh client Run", err)
		return data, err
	}
	log.Logger.Infof("ssh client Run:%s", data)
	return data, nil
}

func IsAlibabaCloudAccountValid(account config.AlibabaCloudAccount) bool {
	if strings.Trim(account.Region, " ") == "" || strings.Trim(account.AccessKey, " ") == "" || strings.Trim(account.Secret, " ") == "" {
		return false
	}
	return true
}
