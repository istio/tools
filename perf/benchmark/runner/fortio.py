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

from __future__ import print_function
import json
import os
import requests
from datetime import datetime
import calendar
import csv
import argparse
import subprocess
import tempfile
import prom

"""
    returns data in a single line format
    Labels, StartTime, RequestedQPS, ActualQPS, NumThreads,
    min, max, p50, p75, p90, p99, p999
"""


def convert_data(data):
    obj = {}

    # These keys are generated from fortio default json file
    for key in "Labels,StartTime,RequestedQPS,ActualQPS,NumThreads,RunType,ActualDuration".split(
            ","):
        if key == "RequestedQPS" and data[key] == "max":
            obj[key] = 99999999
            continue
        if key in ["RequestedQPS", "ActualQPS"]:
            obj[key] = int(round(float(data[key])))
            continue
        if key == "ActualDuration":
            obj[key] = int(data[key] / 10 ** 9)
            continue
        # fill out other data key to obj key
        obj[key] = data[key]

    h = data["DurationHistogram"]
    obj["min"] = int(h["Min"] * 10 ** 6)
    obj["max"] = int(h["Max"] * 10 ** 6)

    p = h["Percentiles"]

    for pp in p:
        obj["p" + str(pp["Percentile"]).replace(".", "")
            ] = int(pp["Value"] * 10 ** 6)

    success = 0
    if '200' in data["RetCodes"]:
        success = int(data["RetCodes"]["200"])

    obj["errorPercent"] = 100 * \
        (int(data["Sizes"]["Count"]) - success) / int(data["Sizes"]["Count"])
    obj["Payload"] = int(data['Sizes']['Avg'])
    return obj


def fetch(url):
    data = None
    if url.startswith("http"):
        try:
            d = requests.get(url)
            if d.status_code != 200:
                return None
            # Add debugging info for JSON parsing error in perf pipeline (nighthawk)
            print("fetching data from fortioclient")
            print(d)
            data = d.json()
        except Exception:
            print("Error while fetching from " + url)
            raise
    else:
        data = json.load(open(url))

    return convert_data(data)


def convert_data_to_list(txt):
    idx = 0
    lines = []

    marker = '<option value="'
    # marker = 'a href="' # This used to be the marker in older version of
    # fortio
    while True:
        idx = txt.find(marker, idx)
        if idx == -1:
            break
        startRef = idx + len(marker)
        end = txt.find('"', startRef)
        lines.append(txt[startRef:end])
        idx += 1
    return lines


# number of seconds to skip after test begins.
METRICS_START_SKIP_DURATION = 62
# number of seconds to skip before test ends.
METRICS_END_SKIP_DURATION = 30
# number of seconds to summarize during test
METRICS_SUMMARY_DURATION = 180


def sync_fortio(url, table, selector=None, promUrl="", csv=None, csv_output=""):
    listurl = url + "/fortio/data/"
    listdata = requests.Response()
    try:
        listdata = requests.get(listurl)
    except requests.exceptions.RequestException as e:
        # TODO handling connection refused issue after logging available
        print(e)
        sys.exit(1)
    fd, datafile = tempfile.mkstemp(suffix=".json")
    out = os.fdopen(fd, "wt")
    stats = []
    cnt = 0

    dataurl = url + "/data/"
    data = []
    for fl in convert_data_to_list(listdata.text):
        gd = fetch(dataurl + fl)
        if gd is None:
            continue
        st = gd['StartTime']
        if selector is not None:
            if selector.startswith("^"):
                if not st.startswith(selector[1:]):
                    continue
            elif selector not in gd["Labels"]:
                continue

        if promUrl:
            sd = datetime.strptime(st[:19], "%Y-%m-%dT%H:%M:%S")
            print("Fetching prometheus metrics for", sd, gd["Labels"])
            if gd['errorPercent'] > 10:
                print("... Run resulted in", gd['errorPercent'], "% errors")
                continue
            min_duration = METRICS_START_SKIP_DURATION + METRICS_END_SKIP_DURATION
            if min_duration > gd['ActualDuration']:
                print("... {} duration={}s is less than minimum {}s".format(
                    gd["Labels"], gd['ActualDuration'], min_duration))
                continue
            prom_start = calendar.timegm(
                sd.utctimetuple()) + METRICS_START_SKIP_DURATION
            duration = min(gd['ActualDuration'] - min_duration,
                           METRICS_SUMMARY_DURATION)
            p = prom.Prom(promUrl, duration, start=prom_start)
            prom_metrics = p.fetch_istio_proxy_cpu_and_mem()
            if not prom_metrics:
                print("... Not found")
                continue
            else:
                print("")

            gd.update(prom_metrics)

        data.append(gd)
        out.write(json.dumps(gd) + "\n")
        stats.append(gd)
        cnt += 1

    out.close()
    print("Wrote {} json records to {}".format(cnt, datafile))

    if csv is not None:
        write_csv(csv, data, csv_output)

    if table:
        return write_table(table, datafile)

    return 0


def write_csv(keys, data, csv_output):
    if csv_output is None or csv_output == "":
        fd, csv_output = tempfile.mkstemp(suffix=".csv")
        out = os.fdopen(fd, "wt")
    else:
        out = open(csv_output, "w+")

    lst = keys.split(',')
    out.write(keys + "\n")

    for gd in data:
        row = []
        for key in lst:
            row.append(str(gd.get(key, '-')))

        out.write(','.join(row) + "\n")

    out.close()
    print("Wrote {} csv records to {}".format(len(data), csv_output))


def write_table(table, datafile):
    print("table: %s, datafile: %s" % (table, datafile))
    p = subprocess.Popen("bq insert {table} {datafile}".format(
        table=table, datafile=datafile).split())
    ret = p.wait()
    print(p.stdout)
    print(p.stderr)
    return ret


def main(argv):
    args = get_parser().parse_args(argv)
    return sync_fortio(
        args.url,
        args.table,
        args.selector,
        args.prometheus,
        args.csv,
        args.csv_output)


def get_parser():
    parser = argparse.ArgumentParser("Fetch and upload results to bigQuery")
    parser.add_argument(
        "--table",
        help="Name of the BigQuery table to send results to, like istio_perf_01.perf",
        default=None)
    parser.add_argument(
        "--selector",
        help="timestamps to match for import")
    parser.add_argument(
        "--csv",
        help="columns in the csv file",
        default="StartTime,ActualDuration,Labels,NumThreads,ActualQPS,p50,p90,p99,"
                "cpu_mili_avg_istio_proxy_fortioclient,cpu_mili_avg_istio_proxy_fortioserver,"
                "cpu_mili_avg_istio_proxy_istio-ingressgateway,mem_Mi_avg_istio_proxy_fortioclient,"
                "mem_Mi_avg_istio_proxy_fortioserver,mem_Mi_avg_istio_proxy_istio-ingressgateway")
    parser.add_argument(
        "--csv_output",
        help="output path of csv file")
    parser.add_argument(
        "url",
        help="url to fetch fortio json results from")
    parser.add_argument(
        "--prometheus",
        help="url to fetch prometheus results from. if blank, will only output Fortio metrics.",
        default="")
    return parser


if __name__ == "__main__":
    import sys
    sys.exit(main(sys.argv[1:]))
