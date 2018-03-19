package rds

import (
	"testing"

	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/database"
	"github.com/aws/aws-sdk-go/aws"
	awsrds "github.com/aws/aws-sdk-go/service/rds"
	"github.com/stretchr/testify/assert"
)

func TestRDSToModel_HappyPath(t *testing.T) {
	s := "test"
	sgs := []*string{&s}
	c := NewRDSTransformerConfig(&s, sgs)
	bee := NewBumblebee(c)

	db, err := bee.RDSToModel(getRDSInstance())
	assert.NotNil(t, db)
	assert.Nil(t, err)

	assert.Equal(t, db.Host, "somedatabase.com")
	assert.Equal(t, db.Port, int64(5432))
	assert.Equal(t, db.Status, database.StatusAvailable)

}

func TestRDSToModel_EndpointNil(t *testing.T) {
	s := "test"
	sgs := []*string{&s}
	c := NewRDSTransformerConfig(&s, sgs)
	bee := NewBumblebee(c)

	i := getRDSInstance()
	i.Endpoint = nil
	db, err := bee.RDSToModel(i)

	assert.NotNil(t, db)
	assert.Nil(t, err)

	assert.Equal(t, db.Host, "")
	assert.Equal(t, db.Port, int64(0))

}

func getRDSInstance() *awsrds.DBInstance {
	return &awsrds.DBInstance{
		Endpoint: &awsrds.Endpoint{
			Address: aws.String("somedatabase.com"),
			Port:    aws.Int64(5432),
		},
		DBInstanceIdentifier: aws.String("test-test-test"),
		AllocatedStorage:     aws.Int64(5),
		DBInstanceStatus:     aws.String("available"),
	}
}
