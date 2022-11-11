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

from django.urls import re_path
from . import views

urlpatterns = [
    re_path('cpu_vs_conn/', views.cpu_vs_conn, name="cpu_conn"),
    re_path('cpu_vs_qps/', views.cpu_vs_qps, name="cpu_qps"),
    re_path('mem_vs_conn/', views.mem_vs_conn, name="mem_conn"),
    re_path('mem_vs_qps/', views.mem_vs_qps, name="mem_qps"),
    re_path('latency_vs_conn/', views.latency_vs_conn, name="latency_conn"),
    re_path('latency_vs_qps/', views.latency_vs_qps, name="latency_qps"),
    re_path('flame_graph/', views.flame_graph, name="flame_graph"),
    re_path('micro_benchmarks/', views.micro_benchmarks, name="micro_benchmarks"),
]
