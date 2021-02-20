package service

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/jumpserver/koko/pkg/common"
	"github.com/jumpserver/koko/pkg/config"
	"github.com/jumpserver/koko/pkg/logger"
	"github.com/jumpserver/koko/pkg/model"
)

var authClient common.Client
var authKey model.AccessKey

func Initial() {
	cf := config.Conf
	authClient = newClient()
	authClient.SetHeader("X-JMS-ORG", "ROOT")
	var err error
	if authKey, err = getAccessKeyFromConfig(cf); err != nil {
		res := RegisterTerminal(cf.Name, cf.BootstrapToken, model.ComponentName)
		if res.Name != cf.Name {
			msg := "register access key failed"
			logger.Error(msg)
			os.Exit(1)
		}
		authKey = res.ServiceAccount.AccessKey
	}
	authClient.SetAuth(&authKey)
	validateAccessAuth()
	MustLoadServerConfigOnce()
}

func newClient() common.Client {
	cf := config.Conf
	cli := common.NewClient(30, cf.CoreHost)
	return cli
}

func validateAccessAuth() {
	cf := config.Conf
	maxTry := 30
	count := 0
	newKeyTry := 0
	for {
		user, err := GetProfile()
		if err == nil && user.Role == "App" {
			break
		}
		if err == model.AccessKeyUnauthorized && cf.AccessKey == "" {
			if newKeyTry > 0 {
				os.Exit(1)
			}
			logger.Error("Access key unauthorized, try to register new access key")
			//registerNewAccessKey()
			newKeyTry++
			continue
		}
		if err != nil {
			msg := "Connect server error or access key is invalid, remove %s run again"
			logger.Errorf(msg, cf.AccessKeyFile)
		} else if user.Role != "App" {
			logger.Error("Access role is not App, is: ", user.Role)
		}
		count++
		time.Sleep(3 * time.Second)
		if count >= maxTry {
			os.Exit(1)
		}
	}
}

func MustLoadServerConfigOnce() {
	var data map[string]interface{}
	_, err := authClient.Get(TerminalConfigURL, &data)
	if err != nil {
		logger.Error("Load config from server error: ", err)
		os.Exit(1)
		return
	}
	data["TERMINAL_HOST_KEY"] = "Hidden"
	msg, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("Marsha server config error: %s", err)
		return
	}
	logger.Debug("Load config from server: " + string(msg))
	_, err = GetTerminalConfig()
	if err != nil {
		logger.Error("Load config from server error: ", err)
	}
}

func GetTerminalConfig() (conf model.TerminalConfig, err error) {
	_, err = authClient.Get(TerminalConfigURL, &conf)
	if err != nil {
		logger.Error("Load config from server error: ", err)
		return
	}
	return
}

func getAccessKeyFromConfig(cf *config.Config) (ak model.AccessKey, err error) {
	if ak, err = model.ParseAccessKeyFromStr(cf.AccessKey); err == nil {
		return
	}
	if ak, err = model.ParseAccessKeyFromFile(cf.AccessKeyFile); err == nil {
		return
	}
	return model.AccessKey{}, model.AccessKeyNotFound
}

func validateAccessKey() (needRegister bool, ok bool) {
	var user model.User
	if res, err := authClient.Get(UserProfileURL, &user); err != nil {
		logger.Error(err)
		if res != nil {
			needRegister = res.StatusCode == http.StatusUnauthorized
		}
		return
	}
	ok = user.Role == "App"
	return
}


//func registerNewAccessKey() {
//	cf := config.Conf
//	keyPath := cf.AccessKeyFile
//	if !path.IsAbs(cf.AccessKeyFile) {
//		keyPath = filepath.Join(cf.RootPath, keyPath)
//	}
//	ak := AccessKey{Path: keyPath}
//	err := ak.RegisterKey()
//	if err != nil {
//		logger.Errorf("Register access key failed: %s", err)
//		os.Exit(1)
//	}
//	authClient.SetAuth(ak)
//}
