package sshd

import (
	"net"

	"github.com/gliderlabs/ssh"
	"github.com/pires/go-proxyproto"

	"github.com/jumpserver/koko/pkg/auth"
	"github.com/jumpserver/koko/pkg/config"
	"github.com/jumpserver/koko/pkg/handler"
	"github.com/jumpserver/koko/pkg/logger"
)

type Server struct {
	sshServer *ssh.Server
}

func (s *Server) Start() {
	logger.Infof("Start SSH server at %s", s.sshServer.Addr)
	ln, err := net.Listen("tcp", s.sshServer.Addr)
	if err != nil {
		logger.Fatal(err)
	}
	proxyListener := &proxyproto.Listener{Listener: ln}
	logger.Fatal(s.sshServer.Serve(proxyListener))
}

func (s *Server) Stop() {
	err := s.sshServer.Close()
	if err != nil {
		logger.Errorf("SSH server close failed: %s", err.Error())
	}
	logger.Info("Close ssh server")
}

func NewServer() *Server {
	handler.Initial()
	conf := config.Conf
	terminalConf :=conf.GetTerminalConf()
	signer, err := terminalConf.LoadHostKey()
	if err != nil {
		logger.Fatal("Load host key error: ", err)
	}
	addr := net.JoinHostPort(conf.BindHost, conf.SSHPort)
	sshServer := &ssh.Server{
		Addr:                       addr,
		KeyboardInteractiveHandler: auth.CheckMFA,
		PasswordHandler:            auth.CheckUserPassword,
		PublicKeyHandler:           auth.CheckUserPublicKey,
		NextAuthMethodsHandler:     auth.MFAAuthMethods,
		HostSigners:                []ssh.Signer{signer},
		Handler:                    handler.SessionHandler,
		SubsystemHandlers: map[string]ssh.SubsystemHandler{
			"sftp": handler.SftpHandler,
		},
	}
	return &Server{sshServer: sshServer}
}
