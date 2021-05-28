package sshd

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/pires/go-proxyproto"
	gossh "golang.org/x/crypto/ssh"
)


const (
	sshChannelSession     = "session"
	sshChannelDirectTCPIP = "direct-tcpip"
	sshSubSystemSFTP      = "sftp"
)

type Server struct {
	Srv *ssh.Server
}

func (s *Server) Start() {
	ln, err := net.Listen("tcp", s.Srv.Addr)
	if err != nil {
		log.Fatalln(err)
	}
	proxyListener := &proxyproto.Listener{Listener: ln}
	log.Println(s.Srv.Serve(proxyListener))
}

func (s *Server) Stop()  {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	log.Println(s.Srv.Shutdown(ctx))
}

type SSHHandler interface {
	GetAddr() string
	GetSigner() ssh.Signer
	KeyboardInteractiveAuth(ctx ssh.Context, challenger gossh.KeyboardInteractiveChallenge) AuthStatus
	PasswordAuth(ctx ssh.Context, password string) AuthStatus
	PublicKeyHandler(ctx ssh.Context, key ssh.PublicKey) AuthStatus
	NextAuthMethodsHandler(ctx ssh.Context) []string
	SessionHandler(ssh.Session)
	SFTPHandler(ssh.Session)
	LocalPortForwardingPermission(ctx ssh.Context, destinationHost string, destinationPort uint32) bool
	DirectTCPIPChannelHandler(srv *ssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx ssh.Context)
}

type AuthStatus ssh.AuthResult

const (
	AuthFailed              = AuthStatus(ssh.AuthFailed)
	AuthSuccessful          = AuthStatus(ssh.AuthSuccessful)
	AuthPartiallySuccessful = AuthStatus(ssh.AuthPartiallySuccessful)
)

func NewSSHServer(handler SSHHandler) *Server {
	srv := &ssh.Server{
		LocalPortForwardingCallback: func(ctx ssh.Context, destinationHost string, destinationPort uint32) bool {
			return handler.LocalPortForwardingPermission(ctx, destinationHost, destinationPort)
		},
		Addr: handler.GetAddr(),
		KeyboardInteractiveHandler: func(ctx ssh.Context, challenger gossh.KeyboardInteractiveChallenge) ssh.AuthResult {
			//auth.CheckMFA()
			return ssh.AuthResult(handler.KeyboardInteractiveAuth(ctx, challenger))
		},
		PasswordHandler: func(ctx ssh.Context, password string) ssh.AuthResult {
			//auth.CheckUserPassword
			return ssh.AuthResult(handler.PasswordAuth(ctx, password))
		},
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) ssh.AuthResult {
			//auth.CheckUserPublicKey
			return ssh.AuthResult(handler.PublicKeyHandler(ctx, key))
		},
		NextAuthMethodsHandler: func(ctx ssh.Context) []string {
			//auth.MFAAuthMethods
			return handler.NextAuthMethodsHandler(ctx)
		},
		HostSigners: []ssh.Signer{handler.GetSigner()},
		Handler:     func(s ssh.Session) { handler.SessionHandler(s) },
		SubsystemHandlers: map[string]ssh.SubsystemHandler{
			sshSubSystemSFTP: func(s ssh.Session) { handler.SFTPHandler(s) },
		},
		ChannelHandlers: map[string]ssh.ChannelHandler{
			sshChannelSession: ssh.DefaultSessionHandler,
			sshChannelDirectTCPIP: func(srv *ssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx ssh.Context) {
				handler.DirectTCPIPChannelHandler(srv, conn, newChan, ctx)
			},
		},
	}
	return &Server{srv}
}

//func StartServer() {
//	handler.Initial()
//	conf := config.GetConf()
//	hostKey := HostKey{Path: conf.HostKeyFile}
//	logger.Debug("Loading host key")
//	signer, err := hostKey.Load()
//	if err != nil {
//		logger.Fatal("Load host key error: ", err)
//	}
//
//	addr := net.JoinHostPort(conf.BindHost, conf.SSHPort)
//	logger.Infof("Start SSH server at %s", addr)
//	sshServer = &ssh.Server{
//		LocalPortForwardingCallback: func(ctx ssh.Context, destinationHost string, destinationPort uint32) bool {
//			return true
//		},
//		Addr:                       addr,
//		KeyboardInteractiveHandler: auth.CheckMFA,
//		PasswordHandler:            auth.CheckUserPassword,
//		PublicKeyHandler:           auth.CheckUserPublicKey,
//		NextAuthMethodsHandler:     auth.MFAAuthMethods,
//		HostSigners:                []ssh.Signer{signer},
//		Handler:                    handler.SessionHandler,
//		SubsystemHandlers: map[string]ssh.SubsystemHandler{
//			sshSubSystemSFTP: handler.SftpHandler,
//		},
//		ChannelHandlers: map[string]ssh.ChannelHandler{
//			sshChannelSession: ssh.DefaultSessionHandler,
//			sshChannelDirectTCPIP: func(srv *ssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx ssh.Context) {
//				ssh.DirectTCPIPHandler(srv, conn, newChan, ctx)
//			},
//		},
//	}
//	ln, err := net.Listen("tcp", addr)
//	if err != nil {
//		logger.Fatal(err)
//	}
//	proxyListener := &proxyproto.Listener{Listener: ln}
//	logger.Fatal(sshServer.Serve(proxyListener))
//}
//
//func StopServer() {
//	err := sshServer.Close()
//	if err != nil {
//		logger.Errorf("SSH server close failed: %s", err.Error())
//	}
//	logger.Info("Close ssh server")
//}
