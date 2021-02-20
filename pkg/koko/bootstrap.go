package koko

import (
	"context"
	"github.com/jumpserver/koko/pkg/exchange"
	"github.com/jumpserver/koko/pkg/i18n"
	"github.com/jumpserver/koko/pkg/logger"
	"github.com/jumpserver/koko/pkg/service"
)

func setupI18n() {
	i18n.Initial()
}

func setupLogger() {
	logger.Initial()
}

func setupServiceAuth() {
	service.Initial()
}

func setupExchange() {
	exchange.Initial()
}

func setupTimingTasks(ctx context.Context) {

}
