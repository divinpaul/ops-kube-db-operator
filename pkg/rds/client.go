package rds

import (
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/database"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awsrds "github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
)

type RDSClient struct {
	client rdsiface.RDSAPI
	RDSTransformer
}

func NewRDSImpure(rdsclient rdsiface.RDSAPI, t RDSTransformer) *RDSClient {
	return &RDSClient{
		client:         rdsclient,
		RDSTransformer: t,
	}
}

func (r *RDSClient) CreateDB(req *database.Request, masterCreds *database.Credential) (*database.Database, error) {

	i, err := r.ModelToRDS(req, masterCreds)
	if err != nil {
		return nil, err
	}

	db, err := r.client.CreateDBInstance(i)
	if err != nil {
		return nil, err
	}

	modelDB, err := r.RDSToModel(db.DBInstance)
	if err != nil {
		return nil, err
	}
	return modelDB, nil
}

func (r *RDSClient) GetDB(dbID database.DatabaseID) (*database.Database, error) {
	dbInput := &awsrds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(string(dbID)),
	}
	dbOutput, err := r.client.DescribeDBInstances(dbInput)
	if err != nil {
		awsError, ok := err.(awserr.Error)
		if ok && awsError.Code() == awsrds.ErrCodeDBInstanceNotFoundFault {
			return nil, nil
		}
		return nil, err
	}

	return r.RDSToModel(dbOutput.DBInstances[0])
}
