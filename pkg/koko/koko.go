package koko

import (
	"context"
	"fmt"

	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jumpserver/koko/pkg/config"
	"github.com/jumpserver/koko/pkg/exchange"
	"github.com/jumpserver/koko/pkg/httpd"
	"github.com/jumpserver/koko/pkg/i18n"
	"github.com/jumpserver/koko/pkg/logger"
	"github.com/jumpserver/koko/pkg/sshd"

	"github.com/jumpserver/koko/pkg/jms-sdk-go/service"


)

var Version = "unknown"

type Application struct {
	webServer *httpd.Server
	sshServer *sshd.Server
	jmsService *service.JMService
	roomManger exchange.RoomManager
	conf *config.Config
}

const (
	timeFormat = "2006-01-02 15:04:05"
	startMsg = `%s
KoKo Version %s, more see https://www.jumpserver.org
Quit the server with CONTROL-C.
`
)

func (a *Application) Start() {
	fmt.Printf(startMsg, time.Now().Format(timeFormat), Version)
	go a.webServer.Start()
	go a.sshServer.Start()
}

func (a *Application) Stop() {
	a.sshServer.Stop()
	a.webServer.Stop()
	logger.Info("Quit The KoKo")
}

func RunForever(confPath string) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	config.Initial(confPath)
	bootstrap(ctx)
	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	app := NewApp()
	app.Start()
	<-gracefulStop
	cancelFunc()
	app.Stop()
}

func bootstrap(ctx context.Context) {
	i18n.Initial()
	logger.Initial()
	//service.Initial(ctx)
	//exchange.Initial(ctx)
	Initial()
}

func NewApp() *Application {
	webSrv := httpd.NewServer()
	sshSrv := sshd.NewSSHServer(nil)
	return &Application{
		webServer: webSrv,
		sshServer: sshSrv,
	}
}
