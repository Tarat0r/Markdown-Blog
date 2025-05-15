#!/bin/bash

# TODO: fix
kubectl apply -f secret.yaml
kubectl apply -f db-deployment.yaml
kubectl apply -f db-service.yaml
kubectl apply -f api-deployment.yaml
kubectl apply -f api-service.yaml
kubectl apply -f frontend-deployment.yaml
kubectl apply -f frontend-service.yaml