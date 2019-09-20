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

from unittest import mock

from . import mesh


def test_context_should_call_functions():
    set_up = mock.MagicMock()
    tear_down = mock.MagicMock()
    ingress_url = 'http://example.com'
    get_ingress_url = mock.MagicMock(return_value=ingress_url)
    env = mesh.Environment('', set_up, tear_down, get_ingress_url)
    with env.context() as url:
        set_up.assert_called_once_with()
        get_ingress_url.assert_called_once_with()
        assert url == ingress_url
    tear_down.assert_called_once_with()
