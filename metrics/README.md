# Verifying Test Scenarios

This provides a framework to monitor the long running scenarios through Prometheus queries.

## Setup

Ensure you have all dependencies with `pip3 install -r requirements.txt`.

## Running Tests

`./check_metrics.py`

## Writing Tests

Scenarios that need to be tested should be added to [test_cases.py](test_cases.py).

Note that the `queries` variable should contain the list of all tests.

For each thing we want to test, we create a Query object. Here is an example:

```python
Query(
    'Total requests', # Description of the query
    'sum(rate(istio_requests_total{}[10m]))', # Prometheus query
    Alarm(
        lambda qps: qps < 30, # This test will fail if this function returns True
        'Expect at least 30qps' # Failure message
    )
)
```

This example queries for the number of queries per second in total over a 10 minute period. It then asserts there are at least 30 qps.

Currently, the Prometheus query must return a single result, so `sum` is useful.

See the [Prometheus docs](https://prometheus.io/docs/prometheus/latest/querying/basics/) for more info.
