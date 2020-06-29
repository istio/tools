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

current_release = [os.getenv('CUR_RELEASE')]


# Create your views here.
def artifact(request):
    cur_href_link, cur_release_names, cur_release_dates, master_href_link, master_release_names, master_release_dates = download.download_benchmark_csv(60)
    cur_release_bundle = get_artifacts_release_bundle(cur_release_dates, cur_release_names, cur_href_link)
    master_release_bundle = get_artifacts_release_bundle(master_release_dates, master_release_names, master_href_link)

    context = {'current_release': current_release,
               'cur_release_bundle': cur_release_bundle,
               'master_release_bundle': master_release_bundle}

    return render(request, "artifact.html", context=context)


def get_artifacts_release_bundle(release_dates, release_names, href_link):
    release_bundle = [[]] * len(release_names)
    gcs_prefix = "https://gcsweb.istio.io/"
    for i in range(len(release_names)):
        release_bundle[i] = [0] * 3
        release_bundle[i][0] = release_dates[i]
        release_bundle[i][1] = release_names[i]
        release_bundle[i][2] = gcs_prefix + href_link[i]
    return release_bundle
