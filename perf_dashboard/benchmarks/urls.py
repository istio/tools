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

from django.conf.urls import url
from . import views

urlpatterns = [
    url('configs_measurements/', views.benchmarks_overview, name="benchmarks_overview"),
    url('cpu_vs_connection/', views.cpu_vs_conn, name="cpu_conn"),
    url('cpu_vs_qps/', views.cpu_vs_qps, name="cpu_qps"),
    url('mem_vs_connection/', views.mem_vs_conn, name="mem_conn"),
    url('mem_vs_qps/', views.mem_vs_qps, name="mem_qps"),
    url('latency_vs_connection/', views.latency_vs_conn, name="latency_conn"),
    url('latency_vs_qps/', views.latency_vs_qps, name="latency_qps"),
    url('flame_graph/', views.flame_graph, name="flame_graph"),
    url('micro_benchmarks/', views.micro_benchmarks, name="micro_benchmarks"),
]
