package model

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/jumpserver/koko/pkg/common"
)

var (
	AccessKeyNotFound     = errors.New("access key not found")
	AccessKeyFileNotFound = errors.New("access key file not found")
	AccessKeyInvalid      = errors.New("access key not valid")
	AccessKeyUnauthorized = errors.New("access key unauthorized")
)

type AccessKey struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

func (a *AccessKey) Sign(r *http.Request) error {
	date := common.HTTPGMTDate()
	signature := common.MakeSignature(a.Secret, date)
	r.Header.Set("Date", date)
	r.Header.Set("Authorization", fmt.Sprintf("Sign %s:%s", a.ID, signature))
	return nil
}

func (a *AccessKey) SaveToFile(keyFilePath string) error {
	keyDir := path.Dir(keyFilePath)
	if !common.FileExists(keyDir) {
		err := os.MkdirAll(keyDir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	f, err := os.Create(keyFilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("%s:%s", a.ID, a.Secret))
	return err
}

func ParseAccessKeyFromStr(key string) (accessKey AccessKey, err error) {
	if key == "" {
		return AccessKey{}, AccessKeyNotFound
	}
	keySlice := strings.Split(strings.TrimSpace(key), ":")
	if len(keySlice) != 2 {
		return AccessKey{}, AccessKeyInvalid
	}
	accessKey.ID = keySlice[0]
	accessKey.Secret = keySlice[1]
	return
}

func ParseAccessKeyFromFile(keyPath string) (accessKey AccessKey, err error) {
	if keyPath == "" {
		return AccessKey{}, AccessKeyNotFound
	}
	_, err = os.Stat(keyPath)
	if err != nil {
		return AccessKey{}, AccessKeyFileNotFound
	}
	buf, err := ioutil.ReadFile(keyPath)
	if err != nil {
		msg := fmt.Sprintf("read access key failed: %s", err)
		return AccessKey{}, errors.New(msg)
	}
	return ParseAccessKeyFromStr(string(buf))
}
