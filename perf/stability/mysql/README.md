# MySQL Testing

TODO:

- Resources
- setup test, bash inline in the client container, no more manual script.
- STRICT without DR will fail.

This tests ensures Istio working with MySQL, with or without mTLS enabled.

The setup consists of two parts

- A MySQL server.
- A MySQL cmmand line client.

Both of them have Istio sidecar injected. We test the connectivity from client to server by sending
a few commands from the cli to the server.

## How To Run The Test

Install MySQL via Helm.

```bash
helm install --name mysql-test .
```

Disable mTLS first.

```bash
kubectl apply -f mtls-disabled.yaml
```

Verify client can talk to server whem mTLS is disabled.

```bash
# Ensure the MySQL service disable mTLS since the default PERMISSIVE mode does not work for MySQL.
kubectl apply -f mtls-disable.yaml
kubectl  exec   $(kubectl get pod -l app=mysql-client -o jsonpath='{.items[0].metadata.name}')  -- mysql -uroot -proot -h mysql-server  -P3306  -e 'show databases;'
```

Enable the mTLS and verify connectivity again.

```bash
kubectl delete -f mtls-disable.yaml
kubectl apply -f mtls-enabled.yaml
kubectl  exec   $(kubectl get pod -l app=mysql-client -o jsonpath='{.items[0].metadata.name}')  -- mysql -uroot -proot -h mysql-server  -P3306  -e 'show databases;'
```