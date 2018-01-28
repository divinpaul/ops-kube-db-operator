package secret_test

import (
	"testing"

	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/secret"
)

func TestDBSecretMapShouldGenerateUrl(t *testing.T) {
	secret := &secret.DBSecret{
		Host:         "my-special-db.kube-system",
		Port:         "5432",
		Password:     "my-password",
		Username:     "bilbo",
		DatabaseName: "postgres",
	}
	mappedSecret := secret.Map()

	url := mappedSecret["DATABASE_URL"]

	if expectedUrl := "postgresql://bilbo:my-password@my-special-db.kube-system:5432/postgres?sslmode=require"; expectedUrl != url {
		t.Errorf("expected url %s, got: %s", expectedUrl, url)
	}
}
