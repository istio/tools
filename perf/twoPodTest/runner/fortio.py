from __future__ import print_function
import json
import os
import tempfile
import subprocess
import requests
import argparse
from datetime import datetime
import calendar
import pytz

import prom

"""
    returns data in a single line format
    Labels, StartTime, RequestedQPS, ActualQPS, NumThreads,
    min, max, p50, p75, p90, p99, p999
"""


def convertData(data):
    obj = {}

    for key in "Labels,StartTime,RequestedQPS,ActualQPS,NumThreads,RunType,ActualDuration".split(","):
        if key == "RequestedQPS" and data[key] == "max":
            obj[key] = 99999999
            continue
        if key in ["RequestedQPS", "ActualQPS"]:
            obj[key] = int(round(float(data[key])))
            continue
        if key == "ActualDuration":
            obj[key] = int(data[key] / 10 ** 9)
            continue

        if key == "Labels":
            labels = data[key]
            obj["mixer"] = True
            if "nomixer_" in labels:
                obj["mixer"] = False
            obj["mixer_cache"] = True
            if "nomixercache_" in labels:
                obj["mixer_cache"] = False

            obj["serversidecar"] = True
            obj["clientsidecar"] = True
            if "nosidecars" in labels:
                obj["serversidecar"] = False
                obj["clientsidecar"] = False

            if "serversidecar" in labels:
                obj["serversidecar"] = True
                obj["clientsidecar"] = False

            obj["proxyaccesslog"] = True
            if "nolog" in labels:
                obj["proxyaccesslog"] = False

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
        success = data["RetCodes"]["200"]

    obj["errorPercent"] = 100 * \
        (data["Sizes"]["Count"] - success) / data["Sizes"]["Count"]
    obj["Payload"] = int(data['Sizes']['Avg'])
    return obj


def fetch(url):
    data = None
    if url.startswith("http"):
        d = requests.get(url)
        data = d.json()
    else:
        data = json.load(open(url))

    return convertData(data)


def converDataToList(txt):
    idx = 0
    lines = []

    href = 'a href="'
    while True:
        idx = txt.find(href, idx)
        if idx == -1:
            break
        startRef = idx + len(href)
        end = txt.find('"', startRef)
        lines.append(txt[startRef:end])
        idx += 1

    return lines


def syncFortio(url, table, selector=None):
    dataurl = url + "/data/"
    data = requests.get(dataurl)
    fd, datafile = tempfile.mkstemp()
    out = os.fdopen(fd, "wt")
    cnt = 0

    for fl in converDataToList(data.text):
        gd = fetch(dataurl + fl)
        st = gd['StartTime']
        if selector is not None:
            if selector.startswith("^"):
                if not st.startswith(selector[1:]):
                    continue
            elif selector not in gd["Labels"]:
                continue

        sd = datetime.strptime(st[:19], "%Y-%m-%dT%H:%M:%S")
        print("Fetching prometheus metrics for", sd, end=' ')

        if gd['errorPercent'] > 10:
            print("... Run resulted in", gd['errorPercent'], "% errors")
            continue
        # give 30s after start of test
        prom_start = calendar.timegm(sd.utctimetuple()) + 30
        p = prom.Prom("http://localhost:9090", 120, start=prom_start)
        prom_metrics = p.fetch_cpu_and_mem()
        if len(prom_metrics) == 0:
            print("... Not found")
            continue
        else:
            print("")

        gd.update(prom_metrics)
        out.write(json.dumps(gd) + "\n")
        cnt += 1

    out.close()
    print("Wrote {} records to {}".format(cnt, datafile))

    # p = subprocess.Popen("bq insert {table} {datafile}".format(
    #     table=table, datafile=datafile).split())
    # ret = p.wait()
    # print(p.stdout)
    # print(p.stderr)
    return 0


def main(argv):
    args = getParser().parse_args(argv)
    return syncFortio(args.url, args.table, args.selector)


def getParser():
    parser = argparse.ArgumentParser("Fetch and upload results to bigQuery")
    parser.add_argument(
        "--table", help="Name of the BigQuery table to send results to", default="istio_perf_01.perf")
    parser.add_argument("--selector", help="timestamps to match for import")
    parser.add_argument("url", help="url to fetch fortio json results from")
    return parser

if __name__ == "__main__":
    import sys
    sys.exit(main(sys.argv[1:]))
