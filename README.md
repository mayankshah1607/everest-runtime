# Everest SDK PoC

This repo is a PoC for the Everest Plugin SDK. Also contains an implementation for ClickHouse as an example.

The CH plugin supports:
- sharding
- replication
- provisioning `clickhouse-keeper`
- using existing zookeeper clusters

## Quick start.

1. Apply CRDs:
```bash
kubectl apply -f config/crds/bases
```

2. Install ClickHouse operator:
```bash
kubectl apply -f https://raw.githubusercontent.com/Altinity/clickhouse-operator/master/deploy/operator/clickhouse-operator-install-bundle.yaml
```

3. Run plugin locally:
```bash
go run main.go
```

> Make sure your $KUBECONFIG points to a running cluster.
