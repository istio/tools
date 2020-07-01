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


# TODO: add load_gen_type as a param
def download_benchmark_csv(days):
    if not os.path.exists(perf_data_path):
        os.makedirs(perf_data_path)

    url_prefix = "https://gcsweb.istio.io/gcs/"
    gcs_bucket_name = "istio-build/perf"
    url = url_prefix + gcs_bucket_name
    try:
        page = request.urlopen(url)
    except Exception as e:
        print(e)
        exit(1)

    cur_href_links = []
    cur_release_names = []
    cur_release_dates = []
    master_href_links = []
    master_release_names = []
    master_release_dates = []
    soup = BeautifulSoup(page, 'html.parser')
    download_dateset = get_download_dateset(days)

    for link in soup.find_all('a'):
        href_str = link.get('href')
        if href_str == "/gcs/istio-build/":
            continue
        # TODO:
        #  - add "nighthawk" check later like: or "nighthawk" in href_str
        #  - maybe only need one href_link list (refactor later when other pages being fixed)
        if "fortio" in href_str:
            href_parts = href_str.split("/")
            # an example benchmark_test_id would be like:
            # "20200525_fortio_master_1.7-alpha.d0e07f6e430fd99554ccc3aee3be8a730cd8a226"
            benchmark_test_id = href_parts[4]
            test_date, test_load_gen_type, test_branch, release_name = parse_perf_href_str(benchmark_test_id)
            if test_date in download_dateset:
                download_prefix = "https://storage.googleapis.com/istio-build/perf/"
                download_filename = "benchmark.csv"
                download_url = download_prefix + benchmark_test_id + "/" + download_filename
                # TODO: this will make all filename to be empty but without this check will slow down the page rendering
                # local_filename = benchmark_test_id + "_" + download_filename
                # if check_exist(local_filename):
                #     continue
                dump_to_filepath = perf_data_path + benchmark_test_id + "_" + download_filename
                if test_branch == "master":
                    master_href_links.insert(0, href_str)
                    master_release_names.insert(0, release_name)
                    master_release_dates.insert(0, test_date)
                else:
                    cur_href_links.insert(0, href_str)
                    cur_release_names.insert(0, release_name)
                    cur_release_dates.insert(0, test_date)
                try:
                    wget.download(download_url, dump_to_filepath)
                except Exception as e:
                    if test_branch == "master":
                        master_href_links.pop(0)
                        master_release_names.pop(0)
                        master_release_dates.pop(0)
                    else:
                        cur_href_links.pop(0)
                        cur_release_names.pop(0)
                        cur_release_dates.pop(0)
                    print(e)
        else:
            continue
    return cur_href_links, cur_release_names, cur_release_dates, master_href_links, master_release_names, master_release_dates


def get_download_dateset(days):
    download_dateset = set()
    today = datetime.date.today()
    for day_interval in list(range(1, days)):
        prev_date = (today - datetime.timedelta(day_interval)).strftime("%Y%m%d")
        download_dateset.add(prev_date)
    return download_dateset


def delete_outdated_files(release_names):
    filenames = ['master_temp.csv', 'cur_temp.csv']
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


def parse_perf_href_str(benchmark_test_id):
    # TODO:
    #   - can make this to be env var: LOAD_GEN_TYPE for switching between fortio and nighthawk
    #   - extract test_parts to a class when pipeline label is stable
    test_parts = benchmark_test_id.split("_")
    test_date = test_parts[0]
    test_load_gen_type = test_parts[1]
    test_branch = test_parts[2]
    release_name = test_parts[3]
    return test_date, test_load_gen_type, test_branch, release_name
