package koko

import (
	"fmt"
	"net"
	"time"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"

	"github.com/jumpserver/koko/pkg/auth"
	"github.com/jumpserver/koko/pkg/common"
	"github.com/jumpserver/koko/pkg/config"
	"github.com/jumpserver/koko/pkg/logger"
	"github.com/jumpserver/koko/pkg/sshd"
)


const (
	nextAuthMethod = "keyboard-interactive"
)

func (a *Application) GetSSHAddr() string {
	cf := config.GlobalConfig
	return net.JoinHostPort(cf.BindHost, cf.SSHPort)
}
func (a *Application) GetSSHSigner() ssh.Signer {
	singer, err := sshd.ParsePrivateKeyFromString(a.terminalConf.HostKey)
	if err != nil {
		logger.Fatal(err)
	}
	return singer
}

func (a *Application) KeyboardInteractiveAuth(ctx ssh.Context,
	challenger gossh.KeyboardInteractiveChallenge) sshd.AuthStatus {
	return auth.SSHKeyboardInteractiveAuth(ctx, challenger)
}

const ctxID = "ctxID"

func (a *Application) PasswordAuth(ctx ssh.Context, password string) sshd.AuthStatus {
	ctx.SetValue(ctxID, ctx.SessionID())
	sshAuthHandler := auth.SSHPasswordAndPublicKeyAuth(a.jmsService)
	return sshAuthHandler(ctx, password, "")
}

func (a *Application) PublicKeyHandler(ctx ssh.Context, key ssh.PublicKey) sshd.AuthStatus {
	ctx.SetValue(ctxID, ctx.SessionID())
	publicKey := common.Base64Encode(string(key.Marshal()))
	sshAuthHandler := auth.SSHPasswordAndPublicKeyAuth(a.jmsService)
	return sshAuthHandler(ctx, "", publicKey)
}

func (a *Application) NextAuthMethodsHandler(ctx ssh.Context) []string {
	return []string{nextAuthMethod}
}

func (a *Application) SessionHandler(sess ssh.Session) {
	fmt.Println("sid: ", sess.Context().Value(ctxID))
	time.Sleep(100 * time.Second)
}

func (a *Application) SFTPHandler(ssh.Session) {

}

func (a *Application) LocalPortForwardingPermission(ctx ssh.Context, destinationHost string, destinationPort uint32) bool {
	return config.GlobalConfig.EnableLocalPortForward
}
func (a *Application) DirectTCPIPChannelHandler(ctx ssh.Context, newChan gossh.NewChannel, destAddr string) {
	// todo vscode 支持
}
