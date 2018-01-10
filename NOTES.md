# Onboarding Notes

## Questions
* access to database for sql operations??
  - use [access-db pgweb script](./bin/access-db) will provide access to pgweb interface on [http://localhost:8081](http://localhost:8081)
  - create a pod that exits to run a single psql script
* postgresdb crd as part of pipeline?? First time will fail. If you delete the crd, db will get deleted so - danger Will Robinson.
* how to migrate data from another aws account/rds instance
  - when create from snapshot feature appears, this may be able to access snaps in a remote account

## Upcoming features

* create db from snapshot
* manual backup

## For PE

* tag rds instances with name, namespace and crd version (ie v1alpha1) to be used for migrations
