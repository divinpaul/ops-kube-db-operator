# Accessing the Instance

To verify access to the RDS Instance is configured correctly one can use a tool like [`pgweb`](https://github.com/sosedoff/pgweb) to get a web ui administration view of the database. Below is an example of a script to help you get access to the DB:

```bash
### access.sh
#!/usr/bin/env bash

###
# TODO: update `--db` flag to represent the actual db name once that is changed in the code
###

NAMESPACE=$1
NAME=$2

DB_AUTH=`kubectl get -n $NAMESPACE secrets $NAME -o json`
DB_USER=`echo $DB_AUTH | jq -r '.data.username'|base64 --decode`
DB_PASSWORD=`echo $DB_AUTH | jq -r '.data.password'|base64 --decode`
DB_PASSWORD=`echo $DB_AUTH | jq -r '.data.host'|base64 --decode`
POD_NAME="pgweb-${USER//[.]/-}"

function cleanup {
  kubectl delete pod $POD_NAME
}
trap cleanup EXIT

$PGWEB_VERSION=0.9.10
kubectl run -n $NAMESPACE $POD_NAME \
    --generator run-pod/v1 \
    --pod-running-timeout=1h \
    --image=sosedoff/pgweb:$PGWEB_VERSION -- \
    pgweb --host $DB_HOST \
        --user $DB_USER \
        --db postgres \
        --pass "$DB_PASSWORD"

# WAIT FOR POD TO BE READY
while true
do
    sleep 1
    STATUS=`kubectl get pod $POD_NAME -o template --template={{.status.phase}}`
    echo "Waiting for pod to be ready: $STATUS ..."
    if [[ $STATUS == *"Running"* ]]; then
        break
    fi
done

echo "Access on http://localhost:8081"
kubectl port-forward $POD_NAME 8081
```

An example usage:

```bash
# call the script with the appropriate namespace and db
./access.sh some-namespace some-postgres-db

# you should now be able to access http://localhost:8081 and have an interface into the DB
```
