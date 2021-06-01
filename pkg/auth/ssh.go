package auth

import (
	"net"
	"strings"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"

	"github.com/jumpserver/koko/pkg/jms-sdk-go/service"
	"github.com/jumpserver/koko/pkg/logger"
	"github.com/jumpserver/koko/pkg/sshd"
)

type SSHAuthFunc func(ctx ssh.Context, password, publicKey string) (res sshd.AuthStatus)

func SSHPasswordAndPublicKeyAuth(jmsService *service.JMService) SSHAuthFunc {
	return func(ctx ssh.Context, password, publicKey string) (res sshd.AuthStatus) {
		username := ctx.User()
		authMethod := "publickey"
		action := actionAccepted
		res = sshd.AuthFailed
		if password != "" {
			authMethod = "password"
		}
		remoteAddr, _, _ := net.SplitHostPort(ctx.RemoteAddr().String())
		userClient, ok := ctx.Value(ContextKeyClient).(*service.UserAuthClient)
		if !ok {
			newClient := jmsService.CloneClient()
			sessionClient := service.NewUserAuthClient(service.UserClientUsername(username),
				service.UserClientPassword(remoteAddr), service.UserClientLoginType("T"),
				service.UserClientHttpClient(&newClient))
			userClient = &sessionClient
			ctx.SetValue(ContextKeyClient, userClient)
		}
		userClient.SetOption(service.UserClientPassword(password),
			service.UserClientPublicKey(publicKey))
		logger.Infof("SSH conn[%s] authenticating user %s %s", ctx.SessionID(), username, authMethod)
		user, authStatus := userClient.Authenticate(ctx)
		switch authStatus {
		case service.AuthMFARequired:
			action = actionPartialAccepted
			res = sshd.AuthPartiallySuccessful
		case service.AuthSuccess:
			res = sshd.AuthSuccessful
			ctx.SetValue(ContextKeyUser, &user)
		case service.AuthConfirmRequired:
			required := true
			ctx.SetValue(ContextKeyConfirmRequired, &required)
			action = actionPartialAccepted
			res = sshd.AuthPartiallySuccessful
		default:
			action = actionFailed
		}
		logger.Infof("SSH conn[%s] %s %s for %s from %s", ctx.SessionID(),
			action, authMethod, username, remoteAddr)
		return

	}
}

func SSHKeyboardInteractiveAuth(ctx ssh.Context, challenger gossh.KeyboardInteractiveChallenge) (res sshd.AuthStatus) {
	if value, ok := ctx.Value(ContextKeyConfirmFailed).(*bool); ok && *value {
		return sshd.AuthFailed
	}
	username := ctx.User()
	remoteAddr, _, _ := net.SplitHostPort(ctx.RemoteAddr().String())
	res = sshd.AuthFailed

	var confirmAction bool
	instruction := mfaInstruction
	question := mfaQuestion

	client, ok := ctx.Value(ContextKeyClient).(*service.UserAuthClient)
	if !ok {
		logger.Errorf("SSH conn[%s] user %s Mfa Auth failed: not found session client.",
			ctx.SessionID(), username)
		return
	}
	value, ok := ctx.Value(ContextKeyConfirmRequired).(*bool)
	if ok && *value {
		confirmAction = true
		instruction = confirmInstruction
		question = confirmQuestion
	}
	answers, err := challenger(username, instruction, []string{question}, []bool{true})
	if err != nil || len(answers) != 1 {
		if confirmAction {
			client.CancelConfirm()
		}
		logger.Errorf("SSH conn[%s] user %s happened err: %s", ctx.SessionID(), username, err)
		return
	}
	if confirmAction {
		switch strings.TrimSpace(strings.ToLower(answers[0])) {
		case "yes", "y", "":
			logger.Infof("SSH conn[%s] checking user %s login confirm", ctx.SessionID(), username)
			user, authStatus := client.CheckConfirm(ctx)
			switch authStatus {
			case service.AuthSuccess:
				res = sshd.AuthSuccessful
				ctx.SetValue(ContextKeyUser, &user)
				logger.Infof("SSH conn[%s] checking user %s login confirm success", ctx.SessionID(), username)
				return
			}
		case "no", "n":
			logger.Infof("SSH conn[%s] user %s cancel login", ctx.SessionID(), username)
			client.CancelConfirm()
		default:
			return
		}
		failed := true
		ctx.SetValue(ContextKeyConfirmFailed, &failed)
		logger.Infof("SSH conn[%s] checking user %s login confirm failed", ctx.SessionID(), username)
		return
	}
	mfaCode := answers[0]
	logger.Infof("SSH conn[%s] checking user %s mfa code", ctx.SessionID(), username)
	user, authStatus := client.CheckUserOTP(ctx, mfaCode)
	switch authStatus {
	case service.AuthSuccess:
		res = sshd.AuthSuccessful
		ctx.SetValue(ContextKeyUser, &user)
		logger.Infof("SSH conn[%s] %s MFA for %s from %s", ctx.SessionID(),
			actionAccepted, username, remoteAddr)
	case service.AuthConfirmRequired:
		res = sshd.AuthPartiallySuccessful
		required := true
		ctx.SetValue(ContextKeyConfirmRequired, &required)
		logger.Infof("SSH conn[%s] %s MFA for %s from %s", ctx.SessionID(),
			actionPartialAccepted, username, remoteAddr)
	default:
		logger.Errorf("SSH conn[%s] %s MFA for %s from %s", ctx.SessionID(),
			actionFailed, username, remoteAddr)
	}
	return
}

const (
	ContextKeyUser            = "CONTEXT_USER"
	ContextKeyClient          = "CONTEXT_CLIENT"
	ContextKeyConfirmRequired = "CONTEXT_CONFIRM_REQUIRED"
	ContextKeyConfirmFailed   = "CONTEXT_CONFIRM_FAILED"
)
