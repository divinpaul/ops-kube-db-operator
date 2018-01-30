#!/usr/bin/env bash

db_name=test-db
tribe_namespace=platform-enablement
kubectl delete secrets -n kube-system ${db_name}-master --ignore-not-found=true
kubectl delete postgresdb -n ${tribe_namespace} --all --ignore-not-found=true
kubectl delete secrets -n ${tribe_namespace}  ${db_name}-admin --ignore-not-found=true
kubectl delete secret,cm,service,deploy ${db_name}-metrics-exporter -n ${tribe_namespace}-shadow --ignore-not-found=true