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

// CreateInstanceInput represents input data for RDS instance creation
type CreateInstanceInput struct {
	InstanceName   string
	Storage        int64
	Size           string
	Backups        bool
	MultiAZ        bool
	MasterUsername string
	MasterPassword string
	SubnetGroup    string
	SecurityGroups []*string
	Tags           map[string]string
}

// CreateInstanceInput represents output data from RDS instance creation
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

// DBInstanceManager provides abstraction for interacting with RDSAPI
type DBInstanceManager struct {
	client rdsiface.RDSAPI
}

// Create will create a new RDS instance if not already existing
func (a *DBInstanceManager) Create(input *CreateInstanceInput) (*CreateInstanceOutput, error) {
	var backupRetention int64

	if input.Backups {
		backupRetention = 35
	}

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
		BackupRetentionPeriod:      aws.Int64(backupRetention),
		PreferredMaintenanceWindow: aws.String("Sat:14:30-Sat:15:30"), // Sun 01:30-02:30 AEDT
		PreferredBackupWindow:      aws.String("13:30-14:30"),         // Sun 00:30-01:30 AEDT
		MasterUserPassword:         aws.String(input.MasterPassword),
		MasterUsername:             aws.String(input.MasterUsername),
		DBSubnetGroupName:          aws.String(input.SubnetGroup),
		Tags:                       mapToTags(input.Tags),
		VpcSecurityGroupIds:        input.SecurityGroups,
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case awsrds.ErrCodeDBInstanceAlreadyExistsFault:
				return a.getInstance(aws.String(input.InstanceName), true)
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

	return a.getInstance(db.DBInstance.DBInstanceIdentifier, false)
}

func (a *DBInstanceManager) getInstance(dbInstanceIdentifier *string, alreadyExists bool) (*CreateInstanceOutput, error) {
	output, err := a.client.DescribeDBInstances(&awsrds.DescribeDBInstancesInput{
		DBInstanceIdentifier: dbInstanceIdentifier,
	})

	if err != nil {
		return nil, err
	}

	dbInstance := output.DBInstances[0]
	return toOutput(dbInstance, alreadyExists), nil
}

func toOutput(dbInstance *awsrds.DBInstance, alreadyExists bool) *CreateInstanceOutput {
	return &CreateInstanceOutput{
		AlreadyExists: alreadyExists,
		Name:          aws.StringValue(dbInstance.DBInstanceIdentifier),
		ARN:           aws.StringValue(dbInstance.DBInstanceArn),
		Address:       aws.StringValue(dbInstance.Endpoint.Address),
		Port:          aws.Int64Value(dbInstance.Endpoint.Port),
	}
}

// NewDBInstanceManager return new instance of DBInstanceManager for interacting with RDSAPI
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
