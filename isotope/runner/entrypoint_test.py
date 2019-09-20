# Copyright Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

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
