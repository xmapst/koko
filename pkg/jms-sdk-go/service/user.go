package service

import (
	"context"
	"github.com/jumpserver/koko/pkg/jms-sdk-go/httplib"
	"github.com/jumpserver/koko/pkg/jms-sdk-go/model"
	"github.com/jumpserver/koko/pkg/logger"
	"time"
)

type authResponse struct {
	Err  string       `json:"error,omitempty"`
	Msg  string       `json:"msg,omitempty"`
	Data dataResponse `json:"data,omitempty"`

	Username    string `json:"username,omitempty"`
	Token       string `json:"token,omitempty"`
	Keyword     string `json:"keyword,omitempty"`
	DateExpired string `json:"date_expired,omitempty"`

	User model.User `json:"user,omitempty"`
}

type dataResponse struct {
	Choices []string `json:"choices,omitempty"`
	Url     string   `json:"url,omitempty"`
}

type AuthOptions struct {
	Name string
	Url  string
}

func NewUserAuthClient(setters ...UserClientOption) UserAuthClient {
	clientOptions := &clientOptions{}
	for _, setter := range setters {
		setter(clientOptions)
	}
	clientOptions.client.SetHeader("X-Forwarded-For", clientOptions.RemoteAddr)
	clientOptions.client.SetHeader("X-JMS-LOGIN-TYPE", clientOptions.LoginType)
	return UserAuthClient{
		option:      clientOptions,
		client:      clientOptions.client,
		authOptions: make(map[string]AuthOptions),
	}
}

type UserAuthClient struct {
	option *clientOptions
	client *httplib.Client

	authOptions map[string]AuthOptions

	authResp *authResponse
}

func (u *UserAuthClient) SetOption(setters ...UserClientOption) {
	for _, setter := range setters {
		setter(u.option)
	}
}

func (u *UserAuthClient) Authenticate(ctx context.Context) (user model.User, authStatus AuthStatus) {
	authStatus = AuthFailed
	data := map[string]string{
		"username":    u.option.Username,
		"password":    u.option.Password,
		"public_key":  u.option.PublicKey,
		"remote_addr": u.option.RemoteAddr,
		"login_type":  u.option.LoginType,
	}
	var resp authResponse
	_, err := u.client.Post(UserTokenAuthURL, data, &resp)
	if err != nil {
		logger.Errorf("User %s Authenticate err: %s", u.option.Username, err)
		return
	}
	if resp.Err != "" {
		switch resp.Err {
		case ErrLoginConfirmWait:
			logger.Infof("User %s login need confirmation", u.option.Username)
			authStatus = AuthConfirmRequired
		case ErrMFARequired:
			for _, item := range resp.Data.Choices {
				u.authOptions[item] = AuthOptions{
					Name: item,
					Url:  resp.Data.Url,
				}
			}
			logger.Infof("User %s login need MFA", u.option.Username)
			authStatus = AuthMFARequired
		default:
			logger.Errorf("User %s login err: %s", u.option.Username, resp.Err)
		}
		return
	}
	if resp.Token != "" {
		u.authResp = &resp
		return resp.User, AuthSuccess
	}
	return
}

func (u *UserAuthClient) CheckUserOTP(ctx context.Context, code string) (user model.User, authStatus AuthStatus) {
	var err error
	authStatus = AuthFailed
	data := map[string]string{
		"code":        code,
		"remote_addr": u.option.RemoteAddr,
		"login_type":  u.option.LoginType,
	}
	for name, authData := range u.authOptions {
		var resp authResponse
		switch name {
		case "opt":
			data["type"] = name
		}
		_, err = u.client.Post(authData.Url, data, &resp)
		if err != nil {
			logger.Errorf("User %s use %s check MFA err: %s", u.option.Username, name, err)
			continue
		}
		if resp.Err != "" {
			logger.Errorf("User %s use %s check MFA err: %s", u.option.Username, name, resp.Err)
			continue
		}
		if resp.Msg == "ok" {
			logger.Infof("User %s check MFA success, check if need admin confirm", u.option.Username)
			return u.Authenticate(ctx)
		}
	}
	logger.Errorf("User %s failed to check MFA", u.option.Username)
	return
}

func (u *UserAuthClient) GetAuthUser() model.User {
	return u.authResp.User
}

const (
	ErrLoginConfirmWait     = "login_confirm_wait"
	ErrLoginConfirmRejected = "login_confirm_rejected"
	ErrLoginConfirmRequired = "login_confirm_required"
	ErrMFARequired          = "mfa_required"
	ErrPasswordFailed       = "password_failed"
)

func (u *UserAuthClient) CheckConfirm(ctx context.Context) (user model.User, authStatus AuthStatus) {
	var err error
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Errorf("User %s exit and cancel confirmation", u.option.Username)
			u.CancelConfirm()
			return
		case <-t.C:
			var resp authResponse
			_, err = u.client.Get(UserConfirmAuthURL, &resp)
			if err != nil {
				logger.Errorf("User %s check confirm err: %s", u.option.Username, err)
				return
			}
			if resp.Err != "" {
				switch resp.Err {
				case ErrLoginConfirmWait:
					logger.Infof("User %s still wait confirm", u.option.Username)
					continue
				case ErrLoginConfirmRejected:
					logger.Infof("User %s confirmation was rejected by admin", u.option.Username)
				default:
					logger.Infof("User %s confirmation was rejected by err: %s", u.option.Username, resp.Err)
				}
				return
			}
			if resp.Msg == "ok" {
				logger.Infof("User %s confirmation was accepted", u.option.Username)
				return u.Authenticate(ctx)
			}
		}
	}
}

func (u *UserAuthClient) CancelConfirm() {
	_, err := u.client.Delete(UserConfirmAuthURL, nil)
	if err != nil {
		logger.Errorf("Cancel User %s confirmation err: %s", u.option.Username, err)
		return
	}
	logger.Infof("Cancel User %s confirmation success", u.option.Username)
}
