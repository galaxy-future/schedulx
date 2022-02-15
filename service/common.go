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
		log.Logger.Info("SSH exec cmd start")
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

	cmd := localCmd
	data, err = SSHClient.Run(cmd)
	log.Logger.Infof("Exec cmd: %s", cmd)
	log.Logger.Infof("SSH client Run Result:%s", data)
	log.Logger.Infof("SSH client Run Error:%v", err)
	if err != nil {
		return data, err
	}
	return data, nil
}

func IsAlibabaCloudAccountValid(account config.AlibabaCloudAccount) bool {
	if strings.Trim(account.Region, " ") == "" || strings.Trim(account.AccessKey, " ") == "" || strings.Trim(account.Secret, " ") == "" {
		return false
	}
	return true
}

func StringSliceDiff(s1, s2 []string) []string {
	if len(s1) == 0 {
		return nil
	}
	if len(s2) == 0 {
		return s1
	}

	tmpMap := make(map[string]int8, len(s1)+len(s2))
	diff := make([]string, 0, len(s1))
	for _, v := range s1 {
		tmpMap[v] |= 1
	}
	for _, v := range s2 {
		tmpMap[v] |= 2
	}
	for k, v := range tmpMap {
		if v == 1 {
			diff = append(diff, k)
		}
	}
	return diff
}
