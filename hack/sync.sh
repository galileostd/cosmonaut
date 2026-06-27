#!/bin/bash

kubectl apply -f - << 'EOF'
apiVersion: cosmonaut.galileostd.io/v1
kind: CosmoComponent
metadata:
  name: spark
  namespace: cosmonaut
spec:
  plugin: spark
  type: processing
  endpoint: "cosmonaut-plugin-spark.cosmonaut.svc.cluster.local:50051"
  healthCheckIntervalSeconds: 60
  config:
    namespace: "cosmonaut"
    service_account: "spark"
---
apiVersion: cosmonaut.galileostd.io/v1
kind: CosmoComponent
metadata:
  name: polaris
  namespace: cosmonaut
spec:
  plugin: polaris
  type: catalog
  endpoint: "cosmonaut-plugin-polaris.cosmonaut.svc.cluster.local:50051"
  healthCheckIntervalSeconds: 60
  config:
    endpoint: "http://polaris.cosmonaut.svc.cluster.local:8181"
    catalog: "POLARIS"
    token: ""
---
apiVersion: cosmonaut.galileostd.io/v1
kind: CosmoComponent
metadata:
  name: trino
  namespace: cosmonaut
spec:
  plugin: trino
  type: query-engine
  endpoint: "cosmonaut-plugin-trino.cosmonaut.svc.cluster.local:50051"
  healthCheckIntervalSeconds: 60
  config:
    endpoint: "http://trino.cosmonaut.svc.cluster.local:8080"
    catalog: "iceberg"
EOF

TOKEN=$(curl -s -X POST http://localhost:8181/api/catalog/v1/oauth/tokens \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&client_id=root&client_secret=polaris&scope=PRINCIPAL_ROLE%3AALL" \
  | jq -r '.access_token')

kubectl patch cosmocomponent polaris -n cosmonaut --type=merge -p "{\"spec\":{\"config\":{\"token\":\"${TOKEN}\"}}}"

sleep 65
curl -s http://localhost:8080/api/v1/components | jq '.items[] | {name, health, message}'