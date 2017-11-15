# ConfigMap Settings

The following is an explanation of the parameters that can be configured in the configmap:

## aws.security.group.id

This is the security group id that will be linked to the RDS Instance. Ideally this will allow incoming connections on port `5432` from the security group attached to the Kubernetes Nodes.

## aws.subnet.group.name

This is the subnet group name for the subnets where the RDS Instances will be allocated.

## aws.kms.key.arn

This is the ARN for the KMS Key to enable data encryption at rest.

## aws.rds.postgres.default.port

This is the port where connection are expected.

## aws.rds.postgres.default.multiaz

Flag to enable multi az by default. Possible values areL `true`, `false`.

## aws.rds.postgres.default.kms.encryption

Wether to enable KMS Encryption by default

## aws.rds.postgres.default.instance.class

The default instance class to use if one is not specified.

## aws.rds.postgres.default.backup.window

The default backup window for snapshot backups, for example: `14:00-15:00`.

## aws.rds.postgres.default.maintenance.window

The default maintenance window, for example: `Sun:23:45-Mon:00:15`

## aws.rds.postgres.default.storage.type

The storage type to use. Possible values: `gp2`, `io1`, `standard`.

## aws.rds.postgres.default.storage.size

The default storage size for the db, in GB, for example `5` would mean 5GB.
