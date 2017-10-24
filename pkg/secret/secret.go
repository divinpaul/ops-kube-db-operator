package secret

import (
	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

// Secret represents a Kubernetes Secret resource
type Secret struct {
	obj    *apiv1.Secret
	client *kubernetes.Clientset
	exists bool
	ns     string
}

// New creates a new Secret object if it does not exist
func New(client *kubernetes.Clientset, namespace string, name string) *Secret {
	s := &Secret{
		exists: true,
		ns:     namespace,
		client: client,
	}

	var err error
	s.obj, err = s.client.Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// create new

			s.obj = &apiv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      name,
				},
				StringData: map[string]string{
					"this": "is-sparta",
				},
			}
			s.exists = false
		} else {
			// something is borked
			log.Infof("Error while getting secret %s/%s: %v", namespace, name, err)
		}
	}

	return s
}

// Save updates a secret when it exists, creates a new one if it doesnt
func (s *Secret) Save() (err error) {
	var obj *apiv1.Secret
	if s.exists {
		log.Infof("Updating secret %s/%s", s.ns, s.obj.ObjectMeta.Name)
		obj, err = s.client.Secrets(s.ns).Update(s.obj)
	} else {
		log.Infof("Creating secret %s/%s", s.ns, s.obj.ObjectMeta.Name)
		obj, err = s.client.Secrets(s.ns).Create(s.obj)
	}
	if err != nil {
		log.Errorf("Error saving secret %s/%s: %v", s.ns, s.obj.ObjectMeta.Name, err)
		return
	}

	log.Infof("Saved secret %s/%s", s.ns, s.obj.ObjectMeta.Name)
	s.obj = obj
	return
}

// SetData overwrites the contents of a secret
func (s *Secret) SetData(username, password string) *Secret {
	s.obj.StringData = map[string]string{
		"username": username,
		"password": password,
	}
	return s
}

// Delete deletes a secret from Kubernetes
func (s *Secret) Delete() error {
	if s.exists {
		err := s.client.Secrets(s.ns).Delete(s.obj.ObjectMeta.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	s.exists = false
	return nil
}
