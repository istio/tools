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

import os
import datetime
from google.cloud import storage


cwd = os.getcwd()
perf_data_path = cwd + "/perf_data/"

load_generator_type = "fortio"


def download_benchmark_csv(download_dataset_days, current_release, project_id, bucket_name):

    if not current_release:
        current_release = os.getenv('CUR_RELEASE')
    if not project_id:
        project_id = os.getenv('PROJECT_ID')
    if not bucket_name:
        bucket_name = os.getenv('BUCKET_NAME')
    if not download_dataset_days:
        download_dataset_days = os.getenv('DOWNLOAD_DATASET_DAYS')

    bucket_prefix = os.getenv('BUCKET_PREFIX')
    bucket_delimiter = os.getenv('BUCKET_DELIMITER')

    print(current_release, project_id, bucket_name, bucket_prefix, bucket_delimiter, download_dataset_days)

    if not os.path.exists(perf_data_path):
        os.makedirs(perf_data_path)

    download_dateset = get_download_dateset(int(download_dataset_days))
    storage_client = storage.Client(project_id)

    bucket = storage_client.bucket(bucket_name)

    blobs = storage_client.list_blobs(bucket_name, prefix=bucket_prefix, delimiter=bucket_delimiter)
    print(list(blobs))
    prefixes = blobs.prefixes
    print(prefixes)

    cur_href_links = []
    cur_release_names = []
    cur_release_dates = []
    master_href_links = []
    master_release_names = []
    master_release_dates = []
    process_prefixes(download_dateset, bucket, prefixes, cur_href_links, cur_release_names, cur_release_dates,
                     master_href_links, master_release_names, master_release_dates, current_release)

    delete_outdated_files(download_dateset)
    return cur_href_links, cur_release_names, cur_release_dates, master_href_links, master_release_names, master_release_dates


def process_prefixes(download_dateset, bucket, prefixes, cur_href_links, cur_release_names, cur_release_dates,
                     master_href_links, master_release_names, master_release_dates, current_release):

    for prefix in prefixes:
        print(f"{prefix}")
        if load_generator_type in prefix:
            # an example benchmark_test_id would be like:
            # "20200525_fortio_master_1.7-alpha.d0e07f6e430fd99554ccc3aee3be8a730cd8a226"
            benchmark_test_id = prefix.split('/')[1]
            if current_release.split("-")[1] in benchmark_test_id or "master" in benchmark_test_id:
                test_date, test_load_gen_type, test_branch, release_name = parse_perf_href_str(benchmark_test_id)
                print(f"date: {test_date}, test_branch: {test_branch}, release_name: {release_name}")
                if test_date in download_dateset:
                    download_filename = "benchmark.csv"
                    dump_filename = benchmark_test_id + "_" + download_filename
                    dump_to_filepath = perf_data_path + dump_filename
                    is_exist = check_exist(dump_filename)

                    # Make the API the same as previously so view.py parsing works.
                    fake_prefix = "/././" + prefix
                    if test_branch == "master":
                        master_href_links.insert(0, fake_prefix)
                        master_release_names.insert(0, release_name)
                        master_release_dates.insert(0, test_date)
                    else:
                        cur_href_links.insert(0, fake_prefix)
                        cur_release_names.insert(0, release_name)
                        cur_release_dates.insert(0, test_date)
                    try:
                        if is_exist:
                            continue
                        blob_id = prefix + download_filename
                        blob = bucket.blob(blob_id)
                        blob.download_to_filename(dump_to_filepath)
                        print(f"downloaded: {blob_id}")
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


def get_download_dateset(download_dataset_days):
    download_dateset = set()
    today = datetime.date.today() + datetime.timedelta(days=1)
    for day_interval in list(range(1, download_dataset_days)):
        prev_date = (today - datetime.timedelta(day_interval)).strftime("%Y%m%d")
        download_dateset.add(prev_date)
    return download_dateset


def delete_outdated_files(download_dateset):
    filenames = ['master_temp.csv', 'cur_temp.csv']
    for f in os.listdir(perf_data_path):
        if f in filenames:
            continue
        f_prefix = f.split("_")[0]
        if f_prefix not in download_dateset:
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
