"""Abstractions for common shell calls."""

import contextlib
import logging
import subprocess
from typing import Dict, List, Union

from . import wait


def run_gcloud(args: List[str], check=False) -> subprocess.CompletedProcess:
    return run(['gcloud', *args], check=check)


def run_kubectl(args: List[str], check=False) -> subprocess.CompletedProcess:
    return run_with_k8s_api(['kubectl', *args], check=check)


def run_with_k8s_api(args: List[str],
                     check=False) -> subprocess.CompletedProcess:
    """Ensures the command succeeds against a responsive Kubernetes API."""
    proc = run(args)

    # Retry while the error is because of connection refusal.
    while 'getsockopt: connection refused' in proc.stderr:
        logging.debug('Kubernetes connection refused; retrying...')
        # Wait until `kubectl version` completes, indicating the
        # Kubernetes API is responsive.
        wait.until(
            lambda: run_kubectl(['version']).returncode == 0,
            retry_interval_seconds=5)
        proc = run(args)

    if check and proc.returncode != 0:
        logging.error('%s\n%s\n%s', proc, proc.stdout, proc.stderr)
        raise subprocess.CalledProcessError(
            proc.returncode, proc.args, output=proc.stdout, stderr=proc.stderr)

    return proc


def run(args: List[str], check=False,
        env: Dict[str, str] = None) -> subprocess.CompletedProcess:
    """Delegates to subprocess.run, capturing stdout and stderr.

    Args:
        args: the list of args, with the command as the first item
        check: if True, raises an exception if the command returns non-zero
        env: the environment variables to set during the command's runtime

    Returns:
        A completed process, with stdout and stderr decoded as UTF-8 strings.
    """
    logging.debug('%s', args)

    try:
        proc = subprocess.run(
            args,
            check=check,
            env=env,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE)
    except subprocess.CalledProcessError as e:
        _decode(e)
        if check:
            logging.error('%s\n%s\n%s', e, e.stdout, e.stderr)
        raise e

    _decode(proc)
    return proc


def _decode(
        proc: Union[subprocess.CompletedProcess, subprocess.CalledProcessError]
) -> None:
    if proc.stdout is not None:
        proc.stdout = proc.stdout.decode('utf-8').strip()
    if proc.stderr is not None:
        proc.stderr = proc.stderr.decode('utf-8').strip()
