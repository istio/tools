# MySQL Testing

This tests ensures Istio working with MySQL, with or without mTLS enabled.

The setup consists of two parts

- A MySQL server.
- A MySQL command line client.

Both of them have Istio sidecar injected. We test the connectivity from client to server by sending
a few commands from the cli to the server.

## How To Run The Test

```bash
make mysql
```
