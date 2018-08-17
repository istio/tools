import textwrap
from unittest import mock

import pytest

from . import entrypoint


def test_extract_url_should_return_service_with_is_entrypoint() -> None:
    contents = textwrap.dedent("""\
        services:
        - name: b
        - name: a
          isEntrypoint: true
    """)

    expected = 'http://a.service-graph.svc.cluster.local:8080'
    with mock.patch('builtins.open', mock.mock_open(read_data=contents)):
        actual = entrypoint.extract_url('fake-file.yaml')

    assert expected == actual


def test_extract_url_should_fail_with_no_entrypoints() -> None:
    contents = textwrap.dedent("""\
        services:
        - name: b
        - name: a
    """)

    with mock.patch('builtins.open', mock.mock_open(read_data=contents)):
        with pytest.raises(ValueError):
            entrypoint.extract_url('fake-file.yaml')


def test_extract_url_should_fail_with_multiple_entrypoints() -> None:
    contents = textwrap.dedent("""\
        services:
        - name: b
          isEntrypoint: true
        - name: a
          isEntrypoint: true
    """)

    with mock.patch('builtins.open', mock.mock_open(read_data=contents)):
        with pytest.raises(ValueError):
            entrypoint.extract_url('fake-file.yaml')
