package core_test

import (
	"errors"
	"testing"

	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/core"
)

func TestCreateDBIfNotExists(t *testing.T) {
	expectedDB := &core.Database{}
	core.Set(core.Functions{
		CreateDBInstance: core.CreateDBInstance(func(req core.CreateDatabaseRequest) (*core.Database, error) {
			return expectedDB, nil
		}),
		DescribeDBInstance: core.DescribeDBInstance(func(name string) (*core.Database, error) {
			return nil, nil
		}),
	})

	req := core.CreateDatabaseRequest{
		Name: "test",
	}

	db, err := core.CreateDatabaseIfNotExists(req)

	if err != nil {
		t.Errorf("error", err)
	}

	if db != expectedDB {
		t.Errorf("Not expected db.")
	}
}
func TestCreateDBIfExists(t *testing.T) {
	expectedDB := &core.Database{}
	core.Set(core.Functions{
		CreateDBInstance: core.CreateDBInstance(func(req core.CreateDatabaseRequest) (*core.Database, error) {
			return nil, nil
		}),
		DescribeDBInstance: core.DescribeDBInstance(func(name string) (*core.Database, error) {
			return expectedDB, nil
		}),
	})

	req := core.CreateDatabaseRequest{
		Name: "test",
	}

	db, err := core.CreateDatabaseIfNotExists(req)

	if err != nil {
		t.Errorf("error", err)
	}

	if db != expectedDB {
		t.Errorf("Not expected db.")
	}
}

func TestCreateDBIfExistsError(t *testing.T) {
	core.Set(core.Functions{
		CreateDBInstance: core.CreateDBInstance(func(req core.CreateDatabaseRequest) (*core.Database, error) {
			return nil, errors.New("test create error")
		}),
		DescribeDBInstance: core.DescribeDBInstance(func(name string) (*core.Database, error) {
			return nil, nil
		}),
	})

	req := core.CreateDatabaseRequest{
		Name: "test",
	}

	db, err := core.CreateDatabaseIfNotExists(req)

	if err == nil {
		t.Errorf("error", err)
	}

	if db != nil {
		t.Errorf("Not expected db.")
	}
}
