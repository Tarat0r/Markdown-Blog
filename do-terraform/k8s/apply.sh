#!/bin/bash

# TODO: fix
kubectl apply -n mdblog -f api-deploy.yaml
kubectl apply -n mdblog -f client.yaml
kubectl apply -n mdblog -f postgres.yaml

kubectl apply -n mdblog -f .env

# kubectl apply -f secret.yaml
# kubectl apply -f db-deployment.yaml
# kubectl apply -f api-service.yaml
# kubectl apply -f frontend-deployment.yaml
# kubectl apply -f frontend-service.yaml
