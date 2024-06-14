package cfn

import "sync"

func NewMock(cfn cfnClient) *App {
	return &App{
		cfn:   cfn,
		cache: &sync.Map{},
	}
}
