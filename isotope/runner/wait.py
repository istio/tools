"""Functions which block until certain conditions."""

import collections
import datetime
import logging
import subprocess
import time
from typing import Callable, List

from . import consts, sh

RETRY_INTERVAL = datetime.timedelta(seconds=5)


def until(predicate: Callable[[], bool],
          retry_interval_seconds: int = RETRY_INTERVAL.seconds) -> None:
    """Calls predicate every RETRY_INTERVAL until it returns True."""
    while not predicate():
        time.sleep(retry_interval_seconds)


def until_output(args: List[str]) -> str:
    output = None
    while output is None:
        stdout = sh.run(args).stdout
        if stdout:
            output = stdout
        else:
            time.sleep(RETRY_INTERVAL.seconds)
    return output


def _until_rollouts_complete(resource_type: str, namespace: str) -> None:
    proc = sh.run_kubectl(
        [
            '--namespace', namespace, 'get', resource_type, '-o',
            'jsonpath={.items[*].metadata.name}'
        ],
        check=True)
    resources = collections.deque(proc.stdout.split(' '))
    logging.info('waiting for %ss in %s (%s) to rollout', resource_type,
                 namespace, ', '.join(resources))
    while len(resources) > 0:
        resource = resources.popleft()
        try:
            # kubectl blocks until ready.
            sh.run_kubectl(
                [
                    '--namespace', namespace, 'rollout', 'status',
                    resource_type, resource
                ],
                check=True)
        except subprocess.CalledProcessError as e:
            msg = 'failed to check rollout status of {}'.format(resource)
            if 'watch closed' in e.stderr:
                logging.debug('%s; retrying later', msg)
                resources.append(resource)
            else:
                logging.error(msg)


def until_deployments_are_ready(
        namespace: str = consts.DEFAULT_NAMESPACE) -> None:
    """Blocks until namespace's deployments' rollout statuses are complete."""
    _until_rollouts_complete('deployment', namespace)


def until_stateful_sets_are_ready(
        namespace: str = consts.DEFAULT_NAMESPACE) -> None:
    """Blocks until namespace's statefulsets' rollout statuses are complete."""
    _until_rollouts_complete('statefulset', namespace)


def until_prometheus_has_scraped() -> None:
    logging.info('allowing Prometheus time to scrape final metrics')
    # Add 5 seconds for more confidence that responses to "/metrics" complete.
    time.sleep(consts.PROMETHEUS_SCRAPE_INTERVAL.seconds + 5)


def until_namespace_is_deleted(
        namespace: str = consts.DEFAULT_NAMESPACE) -> None:
    """Blocks until `kubectl get namespace` returns an error."""
    until(lambda: _namespace_is_deleted(namespace))


def _namespace_is_deleted(namespace: str = consts.DEFAULT_NAMESPACE) -> bool:
    proc = sh.run_kubectl(['get', 'namespace', namespace])
    return proc.returncode != 0


def until_service_graph_is_ready() -> None:
    """Blocks until each node in the service graph reports readiness."""
    until(_service_graph_is_ready)


def _service_graph_is_ready() -> bool:
    proc = sh.run_kubectl(
        [
            '--namespace', consts.SERVICE_GRAPH_NAMESPACE, 'get', 'pods',
            '--selector', consts.SERVICE_GRAPH_SERVICE_SELECTOR, '-o',
            'jsonpath={.items[*].status.conditions[?(@.type=="Ready")].status}'
        ],
        check=True)
    out = proc.stdout
    all_services_ready = out != '' and 'False' not in out
    return all_services_ready
