package rds

//go:generate mockgen -source=$GOFILE -destination=../mocks/mock_aws_transformer.go -package=mocks

import (
	"fmt"

	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/database"
	"github.com/aws/aws-sdk-go/aws"
	awsrds "github.com/aws/aws-sdk-go/service/rds"
)

type RDSTransformer interface {
	RDSToModel(db *awsrds.DBInstance) (*database.Database, error)
	ModelToRDS(req *database.Request, master *database.Credential) (*awsrds.CreateDBInstanceInput, error)
}

type bumblebee struct {
	*rdsConfig
}

type rdsConfig struct {
	dbSubnetGroup    *string
	dbSecurityGroups []*string
}

func NewRDSTransformerConfig(subnetGroup *string, sgs []*string) *rdsConfig {

	return &rdsConfig{
		dbSubnetGroup:    subnetGroup,
		dbSecurityGroups: sgs,
	}
}

func NewBumblebee(c *rdsConfig) RDSTransformer {
	return &bumblebee{c}
}

func (b *bumblebee) RDSToModel(db *awsrds.DBInstance) (*database.Database, error) {
	modelDB := &database.Database{
		ID:      database.DatabaseID(*db.DBInstanceIdentifier),
		Storage: *db.AllocatedStorage,
		Status:  awsStatusMatcher(*db.DBInstanceStatus),
	}

	if db.Endpoint != nil {
		modelDB.Host = *db.Endpoint.Address
		modelDB.Port = *db.Endpoint.Port
	}

	return modelDB, nil
}

func (b *bumblebee) ModelToRDS(req *database.Request, master *database.Credential) (*awsrds.CreateDBInstanceInput, error) {

	class, err := getInstanceClassForSize(&req.Size)
	if err != nil {
		return nil, err
	}
	input := &awsrds.CreateDBInstanceInput{
		DBInstanceIdentifier:       aws.String(string(req.ID)),
		DBInstanceClass:            class,
		MultiAZ:                    aws.Bool(req.HA),
		Tags:                       mapToAWSTags(req.Metadata),
		AllocatedStorage:           aws.Int64(req.Storage),
		CopyTagsToSnapshot:         aws.Bool(true),
		Engine:                     aws.String("postgres"),
		EngineVersion:              aws.String("9.6.5"),
		Port:                       aws.Int64(5432),
		StorageEncrypted:           aws.Bool(true),
		StorageType:                aws.String("gp2"),
		BackupRetentionPeriod:      aws.Int64(35),
		PreferredMaintenanceWindow: aws.String("Sat:14:30-Sat:15:30"), // Sun 01:30-02:30 AEDT
		PreferredBackupWindow:      aws.String("13:30-14:30"),         // Sun 00:30-01:30 AEDT
		MasterUserPassword:         aws.String(string(master.Password)),
		MasterUsername:             aws.String(master.Username),
		DBSubnetGroupName:          b.dbSubnetGroup,
		VpcSecurityGroupIds:        b.dbSecurityGroups,
	}
	err = input.Validate()
	if err != nil {
		return nil, err
	}
	return input, nil
}

/**
available: available,
Unavailable: backing-up,configuring-enhanced-monitoring,starting,
			 modifying,rebooting,renaming,resetting-master-credentials,
			 stopping,creating,deleting,maintenance
error: 	restore-error,failed,inaccessible-encryption-credentials
		incompatible-parameters,,incompatible-option-group,
		incompatible-credentials,incompatible-network,incompatible-restore
*/
func awsStatusMatcher(status string) database.Status {
	switch status {
	case "available":
		return database.StatusAvailable
	case "restore-error":
	case "failed":
	case "inaccessible-encryption-credential":
	case "incompatible-parameters":
	case "incompatible-option-group":
	case "incompatible-credentials":
	case "incompatible-network":
	case "incompatible-restore":
		return database.StatusErrored
	default:
		return database.StatusUnavailable
	}
	return 2
}

func mapToAWSTags(m map[string]string) []*awsrds.Tag {
	var tags []*awsrds.Tag

	for k, v := range m {
		tags = append(tags, &awsrds.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}

	return tags
}

func GetSizeForInstanceClass(class string) (*database.Size, error) {
	for k, v := range getMap() {
		if v == class {
			return &k, nil
		}
	}
	return nil, fmt.Errorf("cannot find instance class for size: %s", class)
}

func getInstanceClassForSize(size *database.Size) (*string, error) {
	c := getMap()
	if v, ok := c[*size]; ok {
		return &v, nil
	}
	return nil, fmt.Errorf("unsupported database size: %v", size)
}

func getMap() map[database.Size]string {
	return map[database.Size]string{
		database.SizeXSmall:  "db.t2.small",
		database.SizeSmall:   "db.t2.medium",
		database.SizeMedium:  "db.t2.xlarge",
		database.SizeLarge:   "db.m4.large",
		database.SizeXLarge:  "db.m4.2xlarge",
		database.SizeXXLarge: "db.m4.4xlarge",
		database.SizeMassive: "db.m4.16xlarge",
	}
}
