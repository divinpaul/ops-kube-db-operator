#!/usr/bin/env bash

db_name=test-db
tribe_namespace=platform-enablement
kubectl get secrets -n kube-system ${db_name}-master -o yaml --ignore-not-found=true
kubectl get secrets,cm,postgresdb -n ${tribe_namespace} |grep ${db_name}
kubectl get secrets,cm,service,deploy ${db_name}-metrics-exporter -n ${tribe_namespace}-shadow --ignore-not-found=true
kubectl get postgresdb ${db_name} -o yaml -n ${tribe_namespace} --ignore-not-found=true

kubectl get secret,cm,service,deploy ${db_name}-metrics-exporter -n ${tribe_namespace} --ignore-not-found=true