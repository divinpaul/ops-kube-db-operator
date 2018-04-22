package k8s

import (
	"encoding/binary"
	"strconv"

	"fmt"

	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/database"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	HOST     = "DB_HOST"
	PORT     = "DB_PORT"
	USER     = "DB_USER"
	PASSWORD = "DB_PASSWORD"
	NAME     = "DB_NAME"
	URL      = "DATABASE_URL"
)

type StoreCreds struct {
	client kubernetes.Interface
}

func NewStoreCreds(client kubernetes.Interface) *StoreCreds {
	return &StoreCreds{
		client: client,
	}
}
func (k *StoreCreds) GetCred(credScope database.Scope, id database.CredentialID) (*database.Credential, error) {
	ns := string(credScope)
	secret, err := k.client.CoreV1().Secrets(ns).Get(string(id), metav1.GetOptions{})

	if err != nil && errors.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return transformSecretToCredential(*secret), nil
}

func (k *StoreCreds) UpdateCred(credential *database.Credential) error {
	ns := string(credential.Scope)
	_, err := k.client.CoreV1().Secrets(ns).Update(transformCredentialToSecret(credential, ns))
	return err
}

func (k *StoreCreds) CreateCred(credential *database.Credential) error {
	ns := string(credential.Scope)
	_, err := k.client.CoreV1().Secrets(ns).Create(transformCredentialToSecret(credential, ns))
	return err
}

func transformSecretToCredential(secret v1.Secret) *database.Credential {

	data := secret.Data
	port, _ := binary.Varint(data[PORT])
	cred := &database.Credential{
		Port:         port,
		Password:     database.Password(data[PASSWORD]),
		Username:     string(data[USER]),
		Host:         string(data[HOST]),
		DatabaseName: string(data[NAME]),
	}
	return cred
}

func transformCredentialToSecret(cred *database.Credential, ns string) *v1.Secret {

	var secret = map[string]string{
		HOST:     cred.Host,
		PASSWORD: string(cred.Password),
		USER:     cred.Username,
		PORT:     strconv.FormatInt(cred.Port, 10),
		NAME:     cred.DatabaseName,
		URL:      fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=require", cred.Username, cred.Password, cred.Host, strconv.FormatInt(cred.Port, 10), cred.DatabaseName),
	}

	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            string(cred.ID),
			Namespace:       ns,
			OwnerReferences: []metav1.OwnerReference{},
			Labels: map[string]string{
				"deployed-with": "ops-kube-db-operator",
				"db-name":       cred.DatabaseName,
			},
		},
		StringData: secret,
	}
}
