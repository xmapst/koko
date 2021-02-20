package model

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type TerminalConfig struct {
	AssetListPageSize   string                 `json:"TERMINAL_ASSET_LIST_PAGE_SIZE"`
	AssetListSortBy     string                 `json:"TERMINAL_ASSET_LIST_SORT_BY"`
	HeaderTitle         string                 `json:"TERMINAL_HEADER_TITLE"`
	HostKey             string                 `json:"TERMINAL_HOST_KEY"`
	PasswordAuth        bool                   `json:"TERMINAL_PASSWORD_AUTH"`
	PublicKeyAuth       bool                   `json:"TERMINAL_PUBLIC_KEY_AUTH"`
	CommandStorage      map[string]interface{} `json:"TERMINAL_COMMAND_STORAGE"`
	ReplayStorage       map[string]interface{} `json:"TERMINAL_REPLAY_STORAGE"`
	SessionKeepDuration time.Duration          `json:"TERMINAL_SESSION_KEEP_DURATION"`
	TelnetRegex         string                 `json:"TERMINAL_TELNET_REGEX"`
	MaxIdleTime         time.Duration          `json:"SECURITY_MAX_IDLE_TIME"`
	HeartbeatDuration   time.Duration          `json:"TERMINAL_HEARTBEAT_INTERVAL"`
}

func (conf *TerminalConfig) EnablePasswordAuth() bool {
	// 确保至少有一个认证
	if !conf.PasswordAuth && !conf.PublicKeyAuth {
		return true
	}
	return conf.PasswordAuth
}

func (conf *TerminalConfig) EnablePublicKeyAuth() bool {
	return conf.PublicKeyAuth
}

func (conf *TerminalConfig) LoadHostKey() (signer ssh.Signer, err error) {
	return ssh.ParsePrivateKey([]byte(conf.HostKey))
}

func (conf *TerminalConfig) String() string {
	var s strings.Builder
	elems := reflect.ValueOf(&conf).Elem()
	typeOfT := elems.Type()
	s.WriteString("{")
	for i := 0; i < elems.NumField(); i++ {
		f := elems.Field(i)
		switch strings.ToLower(typeOfT.Field(i).Name) {
		case "hostKey":
			continue
		default:
			s.WriteString(fmt.Sprintf("%s:%s,", typeOfT.Field(i).Name, f.Interface()))
		}
	}
	s.WriteString("}")
	return s.String()
}
