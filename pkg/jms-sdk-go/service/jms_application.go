package service

import (
	"fmt"

	"github.com/jumpserver/koko/pkg/jms-sdk-go/model"
)

func (s *JMService) GetMySQLApplicationById(appId string) (app model.DatabaseApplication, err error) {
	err = s.getApplicationById(appId, &app)
	return
}

func (s *JMService) GetK8sApplicationById(appId string) (app model.K8sApplication, err error) {
	err = s.getApplicationById(appId, &app)
	return
}

func (s *JMService) getApplicationById(appId string, res interface{}) error {
	reqUrl := fmt.Sprintf(ApplicationDetailURL, appId)
	_, err := s.authClient.Get(reqUrl, res)
	return err
}
