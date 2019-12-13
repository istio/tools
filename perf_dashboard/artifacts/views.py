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
from django.shortcuts import render
import pandas as pd
from helpers import download
import os


# Create your views here.
def artifact(request):
    cur_release_names, master_release_names = download.download_benchmark_csv(30)
    gcs_prefix = "https://gcsweb.istio.io/gcs/istio-build/perf/benchmark_data-"
    cur_release_bundle = [[]] * len(cur_release_names)
    master_release_bundle = [[]] * len(master_release_names)
    for i in range(len(cur_release_names)):
        cur_release_bundle[i] = [0] * 2
        cur_release_bundle[i][0] = cur_release_names[i]
        cur_release_bundle[i][1] = gcs_prefix + cur_release_names[i]

    for i in range(len(master_release_names)):
        master_release_bundle[i] = [0] * 2
        master_release_bundle[i][0] = master_release_names[i]
        master_release_bundle[i][1] = gcs_prefix + master_release_names[i]

    context = {'cur_release_bundle': cur_release_bundle,
               'master_release_bundle': master_release_bundle}

    return render(request, "artifact.html", context=context)
