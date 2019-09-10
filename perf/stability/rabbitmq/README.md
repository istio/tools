# RabbitMq

This test runs an instance of RabbitMQ, as well as a client that sends messages and later tries to read them.

## Creating the template

The base template was generated with:

```bash
helm template stable/rabbitmq --name rabbitmq --set rabbitmq.password=istio --set rabbitmq.username=istio
```

Then, the `securityContext` was shifted to the container level rather than the pod level on line 281.
