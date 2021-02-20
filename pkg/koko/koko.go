package koko

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jumpserver/koko/pkg/config"
	"github.com/jumpserver/koko/pkg/httpd"
	"github.com/jumpserver/koko/pkg/logger"
	"github.com/jumpserver/koko/pkg/sshd"
)

var Version = "unknown"

type KoKo struct {
	sshServer  *sshd.Server
	httpServer *httpd.Server
}

func (k *KoKo) Start() {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("Koko Version %s, more see https://www.jumpserver.org\n", Version)
	fmt.Println("Quit the server with CONTROL-C.")
	go k.sshServer.Start()
	go k.httpServer.Start()
}

func (k *KoKo) Stop() {
	k.sshServer.Stop()
	k.httpServer.Stop()
	logger.Info("Quit The KoKo")
}

func RunForever(confPath string) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	config.Initial(confPath)
	bootstrap(ctx)
	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sshdServer := sshd.NewServer()
	httpServer := httpd.NewServer()
	app := &KoKo{sshdServer, httpServer}
	app.Start()
	<-gracefulStop
	cancelFunc()
	app.Stop()
}

func bootstrap(ctx context.Context) {
	setupI18n()
	setupLogger()
	setupServiceAuth()
	setupExchange()
	setupTimingTasks(ctx)
	Initial()
}
