package rds

import (
	"testing"

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

func newMockRDSApi() *mockRDSApi {
	return &mockRDSApi{
		CreateCalled:   false,
		DescribeCalled: false,
		WaitCalled:     false,
		createFunc: func(awsInput *awsrds.CreateDBInstanceInput) (*awsrds.CreateDBInstanceOutput, error) {
			return &awsrds.CreateDBInstanceOutput{}, nil
		},
		describeFunc: func(input *awsrds.DescribeDBInstancesInput) (*awsrds.DescribeDBInstancesOutput, error) {
			return &awsrds.DescribeDBInstancesOutput{}, nil
		},
		waitFunc: func(_input *awsrds.DescribeDBInstancesInput) error {
			return nil
		},
	}
}

func TestSuccessfulCreate(t *testing.T) {
	mockRds := newMockRDSApi()
	commander := &DBInstanceManager{client: mockRds}

	actual, err := commander.Create(&CreateInstanceInput{})

	if err != nil {
		t.Error(err)
	}

	if !mockRds.CreateCalled {
		t.Errorf("CreateDBInstance method never called.")
	}

	if actual.AlreadyExists {
		t.Errorf("AlreadyExists is true when it should be false: %s", actual)
	}

	if !mockRds.WaitCalled {
		t.Errorf("Wait method never called.")
	}
}
func TestAlreadyExistCreate(t *testing.T) {

}
func TestErroredCreate(t *testing.T) {

}
func TestTimedOutCreate(t *testing.T) {

}
