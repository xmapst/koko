package service

import "github.com/jumpserver/koko/pkg/jms-sdk-go/httplib"

type AuthStatus int64

const (
	AuthSuccess AuthStatus = iota + 1
	AuthFailed
	AuthMFARequired
	AuthConfirmRequired
)

type UserClientOption func(*clientOptions)

func UserClientUsername(username string) UserClientOption {
	return func(args *clientOptions) {
		args.Username = username
	}
}

func UserClientPassword(password string) UserClientOption {
	return func(args *clientOptions) {
		args.Password = password
	}
}

func UserClientPublicKey(publicKey string) UserClientOption {
	return func(args *clientOptions) {
		args.PublicKey = publicKey
	}
}

func UserClientRemoteAddr(remoteAddr string) UserClientOption {
	return func(args *clientOptions) {
		args.RemoteAddr = remoteAddr
	}
}

func UserClientLoginType(loginType string) UserClientOption {
	return func(args *clientOptions) {
		args.LoginType = loginType
	}
}

func UserClientHttpClient(con *httplib.Client) UserClientOption {
	return func(args *clientOptions) {
		args.client = con
	}
}

type clientOptions struct {
	Username   string
	Password   string
	PublicKey  string
	RemoteAddr string
	LoginType  string
	client     *httplib.Client
}

