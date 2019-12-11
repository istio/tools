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
from bs4 import BeautifulSoup
from urllib import request
import os
import wget
import datetime

cwd = os.getcwd()
perf_data_path = cwd + "/perf_data/"
current_release = os.getenv('CUR_RELEASE')
today = datetime.date.today()


def download_benchmark_csv():
    if not os.path.exists(perf_data_path):
        os.makedirs(perf_data_path)

    url_prefix = "https://gcsweb.istio.io/gcs/"
    gcs_bucket_name = "istio-build/perf"
    url = url_prefix + gcs_bucket_name
    page = request.urlopen(url)
    current_release_names = []
    master_release_names = []
    soup = BeautifulSoup(page, 'html.parser')
    for link in soup.find_all('a'):
        href_str = link.get('href')
        if href_str == "/gcs/istio-build/":
            continue
        download_prefix = "https://storage.googleapis.com/"
        for day_interval in list(range(1, 10)):
            prev_date = today - datetime.timedelta(day_interval)
            release_name = href_str.split("/")[4][15:]
            filename = release_name + ".csv"
            d = prev_date.strftime("%Y%m%d")
            if d in release_name and current_release in release_name:
                current_release_names.append(release_name)
                if not check_exist(filename):
                    download_url = download_prefix + href_str[5:] + "benchmark.csv"
                    wget.download(download_url, perf_data_path + release_name + ".csv")
            if d in release_name and "master" in release_name:
                master_release_names.append(release_name)
                if not check_exist(filename):
                    download_url = download_prefix + href_str[5:] + "benchmark.csv"
                    wget.download(download_url, perf_data_path + release_name + ".csv")
        delete_outdated_files(current_release_names + master_release_names)
    return current_release_names, master_release_names


def delete_outdated_files(release_names):
    filenames = []
    for release in release_names:
        filenames.append(release + ".csv")
    for f in os.listdir(perf_data_path):
        if f not in filenames:
            os.remove(perf_data_path + f)


def check_exist(filename):
    for f in os.listdir(perf_data_path):
        if f == filename:
            return True
    return False
