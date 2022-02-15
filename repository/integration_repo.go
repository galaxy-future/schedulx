package repository

import (
	"context"
	"sync"

	"github.com/galaxy-future/schedulx/repository/model/db"
)

type IntegrationRepo struct {
}

func (r *IntegrationRepo) Create(ctx context.Context, host string, account string, password string, _type string) error {
	return nil
}

func (r *IntegrationRepo) Delete(ctx context.Context, id int) error {
	return nil
}

func (r *IntegrationRepo) List(ctx context.Context, _type string, num int, size int) ([]db.Integration, int, error) {
	return nil, 0, nil
}

var integrationInstance *IntegrationRepo
var integrationRepoOnce sync.Once

func GetIntegrationInstance() *IntegrationRepo {
	integrationRepoOnce.Do(func() {
		integrationInstance = &IntegrationRepo{}
	})
	return integrationInstance
}
