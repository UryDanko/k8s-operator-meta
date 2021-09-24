#!/bin/bash

kubectl port-forward --namespace metacontroller $(kubectl get pod --namespace metacontroller --selector="app=sandbox-controller" --output jsonpath='{.items[0].metadata.name}') 40000:40000
