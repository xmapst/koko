package koko

import (
	"github.com/jumpserver/koko/pkg/jms-sdk-go/model"
	"github.com/jumpserver/koko/pkg/jms-sdk-go/service"
)


type Application struct {
	terminalConf *model.TerminalConfig
	jmsService   *service.JMService
}

