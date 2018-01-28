package rds

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	awsrds "github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
)

type mockRDSApi struct {
	rdsiface.RDSAPI
	createFunc     func(*awsrds.CreateDBInstanceInput) (*awsrds.CreateDBInstanceOutput, error)
	describeFunc   func(input *awsrds.DescribeDBInstancesInput) (*awsrds.DescribeDBInstancesOutput, error)
	waitFunc       func(input *awsrds.DescribeDBInstancesInput) error
	CreateCalled   bool
	DescribeCalled bool
	WaitCalled     bool
}

func (r *mockRDSApi) CreateDBInstance(input *awsrds.CreateDBInstanceInput) (*awsrds.CreateDBInstanceOutput, error) {
	r.CreateCalled = true
	return r.createFunc(input)
}

func (r *mockRDSApi) DescribeDBInstances(input *awsrds.DescribeDBInstancesInput) (*awsrds.DescribeDBInstancesOutput, error) {
	r.DescribeCalled = true
	return r.describeFunc(input)
}

func (r *mockRDSApi) WaitUntilDBInstanceAvailable(input *awsrds.DescribeDBInstancesInput) error {
	r.WaitCalled = true
	return r.waitFunc(input)
}

func newMockRDSApi(createOutput *awsrds.CreateDBInstanceOutput, descOutput *awsrds.DescribeDBInstancesOutput) *mockRDSApi {
	return &mockRDSApi{
		CreateCalled:   false,
		DescribeCalled: false,
		WaitCalled:     false,
		createFunc: func(awsInput *awsrds.CreateDBInstanceInput) (*awsrds.CreateDBInstanceOutput, error) {
			return createOutput, nil
		},
		describeFunc: func(input *awsrds.DescribeDBInstancesInput) (*awsrds.DescribeDBInstancesOutput, error) {
			return descOutput, nil
		},
		waitFunc: func(_input *awsrds.DescribeDBInstancesInput) error {
			return nil
		},
	}
}

func TestSuccessfulCreate(t *testing.T) {
	// Given
	createOutput := &awsrds.CreateDBInstanceOutput{DBInstance: aDBInstance("dbName", "arn::123")}

	describeOutput := &awsrds.DescribeDBInstancesOutput{DBInstances: []*awsrds.DBInstance{
		aDBInstanceWithEndpoint("dbName", "arn::123", "db.awscom", 1234),
	}}

	mockRds := newMockRDSApi(createOutput, describeOutput)
	dbManager := &DBInstanceManager{client: mockRds}

	// When
	actual, err := dbManager.Create(&CreateInstanceInput{})

	// Then
	if err != nil {
		t.Error(err)
	}

	if !mockRds.CreateCalled {
		t.Errorf("CreateDBInstance method never called.")
	}

	if !mockRds.DescribeCalled {
		t.Errorf("DescribeDBInstance method never called.")
	}

	if !mockRds.WaitCalled {
		t.Errorf("Wait method never called.")
	}

	if actual.AlreadyExists {
		t.Errorf("AlreadyExists is true when it should be false: %v", actual)
	}

	if expected := "dbName"; actual.Name != expected {
		t.Errorf("Expected Name to be %s, got:%v", expected, actual.Name)
	}

	if expected := "arn::123"; actual.ARN != expected {
		t.Errorf("Expected ARN to be %s, got:%v", expected, actual.Name)
	}

	if expected := "db.awscom"; actual.Address != expected {
		t.Errorf("Expected Address to be %s, got:%v", expected, actual.Name)
	}

	if expected := int64(1234); actual.Port != expected {
		t.Errorf("Expected Port to be %v, got:%v", expected, actual.Name)
	}
}

func TestAlreadyExistCreate(t *testing.T) {

}

func TestErroredCreate(t *testing.T) {

}
func TestTimedOutCreate(t *testing.T) {

}

func aDBInstance(name, arn string) *awsrds.DBInstance {
	return &awsrds.DBInstance{
		DBInstanceIdentifier: aws.String(name),
		DBInstanceArn:        aws.String(arn),
	}
}

func aDBInstanceWithEndpoint(name, arn, endpointAddress string, port int64) *awsrds.DBInstance {
	return &awsrds.DBInstance{
		DBInstanceIdentifier: aws.String(name),
		DBInstanceArn:        aws.String(arn),
		Endpoint: &awsrds.Endpoint{
			Address: aws.String(endpointAddress),
			Port:    aws.Int64(port),
		},
	}
}
