package srvconn

import (
	"net"
	"strconv"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

type SSHOption func(conf *SSHConfig)

type SSHConfig struct {
	Host         string
	Port         string
	Username     string
	Password     string
	PrivateKey   string
	Passphrase   string
	Timeout      int
	keyboardAuth gossh.KeyboardInteractiveChallenge
	PrivateAuth  gossh.Signer
	proxyClient  *gossh.Client
}

func (cfg *SSHConfig) AuthMethods() []gossh.AuthMethod {
	authMethods := make([]gossh.AuthMethod, 0, 3)
	if cfg.Password != "" {
		authMethods = append(authMethods, gossh.Password(cfg.Password))
	}
	if cfg.keyboardAuth == nil {
		cfg.keyboardAuth = func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
			return []string{cfg.Password}, nil
		}
	}
	authMethods = append(authMethods, gossh.KeyboardInteractive(cfg.keyboardAuth))

	if cfg.PrivateKey != "" {
		var (
			signer gossh.Signer
			err    error
		)
		if cfg.Passphrase != "" {
			// 先使用 passphrase 获取 signer
			signer, err = gossh.ParsePrivateKeyWithPassphrase([]byte(cfg.PrivateKey), []byte(cfg.Passphrase))
			if err != nil {
				// 如果失败，则去掉 passphrase 再尝试获取 signer 防止错误的passphrase
				signer, err = gossh.ParsePrivateKey([]byte(cfg.PrivateKey))
			}
		} else {
			signer, err = gossh.ParsePrivateKey([]byte(cfg.PrivateKey))
		}
		if signer != nil {
			authMethods = append(authMethods, gossh.PublicKeys(signer))
		}
	}

	if cfg.PrivateAuth != nil {
		authMethods = append(authMethods, gossh.PublicKeys(cfg.PrivateAuth))
	}

	return authMethods
}

func SSHUsername(username string) SSHOption {
	return func(args *SSHConfig) {
		args.Username = username
	}
}

func SSHPassword(password string) SSHOption {
	return func(args *SSHConfig) {
		args.Password = password
	}
}

func SSHPrivateKey(privateKey string) SSHOption {
	return func(args *SSHConfig) {
		args.PrivateKey = privateKey
	}
}

func SSHPassphrase(passphrase string) SSHOption {
	return func(args *SSHConfig) {
		args.Passphrase = passphrase
	}
}

func SSHHost(host string) SSHOption {
	return func(args *SSHConfig) {
		args.Host = host
	}
}

func SSHPort(port int) SSHOption {
	return func(args *SSHConfig) {
		args.Port = strconv.Itoa(port)
	}
}

func SSHTimeout(timeout int) SSHOption {
	return func(args *SSHConfig) {
		args.Timeout = timeout
	}
}

func SSHProxyClient(proxyClient *gossh.Client) SSHOption {
	return func(option *SSHConfig) {
		option.proxyClient = proxyClient
	}
}

func SSHPrivateAuth(privateAuth gossh.Signer) SSHOption {
	return func(conf *SSHConfig) {
		conf.PrivateAuth = privateAuth
	}
}

func SSHKeyboardAuth(keyboardAuth gossh.KeyboardInteractiveChallenge) SSHOption {
	return func(conf *SSHConfig) {
		conf.keyboardAuth = keyboardAuth
	}
}

func NewSSHClient(opts ...SSHOption) (*SSHClient, error) {
	conf := &SSHConfig{
		Host: "127.0.0.1",
		Port: "22",
	}
	for _, setter := range opts {
		setter(conf)
	}
	gosshConfig := gossh.ClientConfig{
		User:              conf.Username,
		Auth:              conf.AuthMethods(),
		Timeout:           time.Duration(conf.Timeout) * time.Second,
		HostKeyCallback:   gossh.InsecureIgnoreHostKey(),
		HostKeyAlgorithms: supportedHostKeyAlgos,
		Config: gossh.Config{
			KeyExchanges: supportedKexAlgos,
			Ciphers:      supportedCiphers,
		},
	}
	destAddr := net.JoinHostPort(conf.Host, conf.Port)
	if conf.proxyClient != nil {
		destConn, err := conf.proxyClient.Dial("tcp", destAddr)
		if err != nil {
			return nil, err
		}
		proxyConn, chans, reqs, err := gossh.NewClientConn(destConn, destAddr, &gosshConfig)
		if err != nil {
			_ = destConn.Close()
			return nil, err
		}
		gosshClient := gossh.NewClient(proxyConn, chans, reqs)
		return &SSHClient{Cfg: conf, Conn: gosshClient}, nil
	}
	gosshClient, err := gossh.Dial("tcp", destAddr, &gosshConfig)
	if err != nil {
		return nil, err
	}

	return &SSHClient{Cfg: conf, Conn: gosshClient}, nil
}

type SSHClient struct {
	Conn *gossh.Client
	Cfg  *SSHConfig
}
