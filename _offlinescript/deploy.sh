#!/usr/bin/env bash

docker build server -t wangbeyond/carpark:latest
docker push grabtalaria/talaria:latest

# delete the pod currently running
kubectl delete po $(kubectl get po -l app=carpark --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}')