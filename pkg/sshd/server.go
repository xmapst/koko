package sshd

import (
	"context"
	"net"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/pires/go-proxyproto"
	gossh "golang.org/x/crypto/ssh"

	"github.com/jumpserver/koko/pkg/auth"
	"github.com/jumpserver/koko/pkg/config"
	"github.com/jumpserver/koko/pkg/handler"
	"github.com/jumpserver/koko/pkg/logger"
)

var sshServer *ssh.Server

const (
	sshChannelSession     = "session"
	sshChannelDirectTCPIP = "direct-tcpip"
	sshSubSystemSFTP      = "sftp"
)

type SSHServer struct {
	srv *ssh.Server
}

func (s *SSHServer) Start() error {
	ln, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return err
	}
	proxyListener := &proxyproto.Listener{Listener: ln}
	return sshServer.Serve(proxyListener)
}

func (s *SSHServer) Stop() error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	return s.srv.Shutdown(ctx)
}



func NewSSHServer(){

}



func StartServer() {
	handler.Initial()
	conf := config.GetConf()
	hostKey := HostKey{Path: conf.HostKeyFile}
	logger.Debug("Loading host key")
	signer, err := hostKey.Load()
	if err != nil {
		logger.Fatal("Load host key error: ", err)
	}

	addr := net.JoinHostPort(conf.BindHost, conf.SSHPort)
	logger.Infof("Start SSH server at %s", addr)
	sshServer = &ssh.Server{
		LocalPortForwardingCallback: func(ctx ssh.Context, destinationHost string, destinationPort uint32) bool {
			return true
		},
		Addr:                       addr,
		KeyboardInteractiveHandler: auth.CheckMFA,
		PasswordHandler:            auth.CheckUserPassword,
		PublicKeyHandler:           auth.CheckUserPublicKey,
		NextAuthMethodsHandler:     auth.MFAAuthMethods,
		HostSigners:                []ssh.Signer{signer},
		Handler:                    handler.SessionHandler,
		SubsystemHandlers: map[string]ssh.SubsystemHandler{
			sshSubSystemSFTP: handler.SftpHandler,
		},
		ChannelHandlers: map[string]ssh.ChannelHandler{
			sshChannelSession: ssh.DefaultSessionHandler,
			sshChannelDirectTCPIP: func(srv *ssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx ssh.Context) {
				ssh.DirectTCPIPHandler(srv, conn, newChan, ctx)
			},
		},
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal(err)
	}
	proxyListener := &proxyproto.Listener{Listener: ln}
	logger.Fatal(sshServer.Serve(proxyListener))
}

func StopServer() {
	err := sshServer.Close()
	if err != nil {
		logger.Errorf("SSH server close failed: %s", err.Error())
	}
	logger.Info("Close ssh server")
}
