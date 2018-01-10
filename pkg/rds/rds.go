// Package rds hides and simplifies the AWS SDK for creating RDS instances
// It provides a blocking API for creating and deleting DB Instances
// This centralised the logic and exposes an interface for mocking elsewhere.
package rds

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	awsrds "github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
)

type CreateInstanceInput struct {
	InstanceName   string
	Storage        int64
	Size           string
	Backups        bool
	MultiAZ        bool
	MasterUsername string
	MasterPassword string
	Tags           map[string]string
}

type CreateInstanceOutput struct {
	Name          string
	ARN           string
	Address       string
	Port          int64
	AlreadyExists bool
}

// DBInstanceCreator is an interface for creating a RDS Instance that is
// blocking
type DBInstanceCreator interface {
	// Create will create an RDS Instance and block until
	// the creation is successfully completed, or return an error if
	// something goes wrong. Create will also return whether the instance
	// was already created previously as a bool rather than in an error
	// for easy logic if the instance already exists. This means it will be
	// idempotent.
	Create(*CreateInstanceInput) (*CreateInstanceOutput, error)
}

type DBInstanceManager struct {
	client rdsiface.RDSAPI
}

func (a *DBInstanceManager) Create(input *CreateInstanceInput) (*CreateInstanceOutput, error) {
	db, err := a.client.CreateDBInstance(&awsrds.CreateDBInstanceInput{
		DBInstanceIdentifier:       aws.String(input.InstanceName),
		DBInstanceClass:            aws.String(input.Size),
		CopyTagsToSnapshot:         aws.Bool(true),
		Engine:                     aws.String("postgres"),
		EngineVersion:              aws.String("9.6.5"),
		Port:                       aws.Int64(5432),
		AllocatedStorage:           aws.Int64(input.Storage),
		StorageEncrypted:           aws.Bool(true),
		StorageType:                aws.String("gp2"),
		MultiAZ:                    aws.Bool(input.MultiAZ),
		BackupRetentionPeriod:      aws.Int64(0),
		PreferredMaintenanceWindow: aws.String("Sat:14:30-Sat:15:30"), // Sun 01:30-02:30 AEDT
		PreferredBackupWindow:      aws.String("Sat:13:30-Sat:14:30"), // Sun 00:30-01:30 AEDT
		MasterUserPassword:         aws.String(input.MasterPassword),
		MasterUsername:             aws.String(input.MasterUsername),
		Tags:                       mapToTags(input.Tags),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case awsrds.ErrCodeDBInstanceAlreadyExistsFault:
				output, derr := a.client.DescribeDBInstances(&awsrds.DescribeDBInstancesInput{
					DBInstanceIdentifier: db.DBInstance.DBInstanceIdentifier,
				})

				// TODO: Confirm the DB has our tags and we managed it, otherwise return an error

				if derr != nil {
					return nil, derr
				}

				return &CreateInstanceOutput{
					AlreadyExists: true,
					Name:          aws.StringValue(output.DBInstances[0].DBInstanceIdentifier),
					ARN:           aws.StringValue(output.DBInstances[0].DBInstanceArn),
					Address:       aws.StringValue(output.DBInstances[0].DBInstanceArn),
					Port:          aws.Int64Value(output.DBInstances[0].DbInstancePort),
				}, nil
			}
		}

		return nil, err
	}

	err = a.client.WaitUntilDBInstanceAvailable(&awsrds.DescribeDBInstancesInput{
		DBInstanceIdentifier: db.DBInstance.DBInstanceIdentifier,
	})

	if err != nil {
		return nil, err
	}

	return &CreateInstanceOutput{
		AlreadyExists: false,
		Name:          aws.StringValue(db.DBInstance.DBInstanceIdentifier),
		ARN:           aws.StringValue(db.DBInstance.DBInstanceArn),
		Address:       aws.StringValue(db.DBInstance.DBInstanceArn),
		Port:          aws.Int64Value(db.DBInstance.DbInstancePort),
	}, nil
}

func NewDBInstanceManager() *DBInstanceManager {
	session, _ := session.NewSession(aws.NewConfig())
	manager := &DBInstanceManager{client: awsrds.New(session)}
	return manager
}

func mapToTags(m map[string]string) []*awsrds.Tag {
	tags := []*awsrds.Tag{}

	for k, v := range m {
		tags = append(tags, &awsrds.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}

	return tags
}
