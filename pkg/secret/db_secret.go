package secret

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

const (
	HOST     = "DB_HOST"
	PORT     = "DB_PORT"
	USER     = "DB_USER"
	PASSWORD = "DB_PASSWORD"
	NAME     = "DB_NAME"
	INSTANCE = "DB_INSTANCE"
	URL      = "DATABASE_URL"
)

type DBSecret struct {
	Name         string
	Namespace    string
	Host         string
	Port         string
	Password     string
	Username     string
	DatabaseName string
	InstanceName string

	client corev1.SecretsGetter
}

func (d *DBSecret) Save() error {
	obj, err := d.client.Secrets(d.Namespace).Get(d.Name, metav1.GetOptions{})

	if err != nil && errors.IsNotFound(err) {
		obj = &apiv1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: d.Namespace,
				Name:      d.Name,
			},
			StringData: d.Map(),
		}
		obj, err = d.client.Secrets(d.Namespace).Create(obj)

		return err
	} else if err != nil {
		return err
	}

	obj.StringData = d.Map()
	obj, err = d.client.Secrets(d.Namespace).Update(obj)

	return err
}

func (d *DBSecret) Delete() error {
	return d.client.Secrets(d.Namespace).Delete(d.Name, &metav1.DeleteOptions{})
}

// Setup stringer interface for printing
func (d *DBSecret) String() string {
	return fmt.Sprintf("Database secret %s - %s for: %s@%s:%s", d.Namespace, d.Name, d.Username, d.Host, d.Port)
}

// Centralise secret key structure
func (d *DBSecret) Map() map[string]string {
	return map[string]string{
		HOST:     d.Host,
		PORT:     d.Port,
		NAME:     d.DatabaseName,
		INSTANCE: d.InstanceName,
		USER:     d.Username,
		PASSWORD: d.Password,
		URL:      fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require", d.Username, d.Password, d.Host, d.Port, d.DatabaseName),
	}
}

func NewOrGet(client corev1.SecretsGetter, namespace, name string) (*DBSecret, error) {
	obj, err := client.Secrets(namespace).Get(name, metav1.GetOptions{})

	if err != nil && errors.IsNotFound(err) {
		return &DBSecret{Name: name, Namespace: namespace, client: client}, nil
	}

	if err != nil {
		return nil, err
	}

	dbSecret := &DBSecret{
		Name:         obj.Name,
		Namespace:    obj.Namespace,
		Host:         string(obj.Data[HOST]),
		Port:         string(obj.Data[PORT]),
		Username:     string(obj.Data[USER]),
		Password:     string(obj.Data[PASSWORD]),
		DatabaseName: string(obj.Data[NAME]),
		InstanceName: string(obj.Data[INSTANCE]),
		client:       client,
	}

	return dbSecret, nil
}
