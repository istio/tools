import datetime
import calendar
import requests
import collections
import os
import json
import argparse


class Prom(object):
    # url: base url for prometheus
    #

    def __init__(self, url, nseconds, end=None, host=None, start=None):
        self.url = url
        if start is None:
            end = end or 0
            self.end = calendar.timegm(
                datetime.datetime.utcnow().utctimetuple()) - end
            self.start = self.end - nseconds
        else:
            self.start = start
            self.end = start + nseconds

        self.headers = {}
        if host is not None:
            self.headers["Host"] = host

    def fetch(self, query, groupby=None, xform=None):
        resp = requests.get(self.url + "/api/v1/query_range", {
            "query": query,
            "start": self.start,
            "end": self.end,
            "step": "15"
        }, headers=self.headers)

        if not resp.ok:
            raise Exception(str(resp))

        data = resp.json()
        return computeMinMaxAvg(data, groupby=groupby, xform=xform)

    def fetch_cpu_by_container(self):
        return self.fetch(
            'rate(container_cpu_usage_seconds_total{container_name=~"mixer|policy|discovery|istio-proxy|captured|uncaptured"}[1m])',
            metric_by_deployment_by_container,
            to_miliCpus)

    def fetch_memory_by_container(self):
        return self.fetch(
            'container_memory_usage_bytes{container_name=~"mixer|policy|discovery|istio-proxy|captured|uncaptured"}',
            metric_by_deployment_by_container,
            to_megaBytes)

    def fetch_cpu_and_mem(self):
        out = flatten(self.fetch_cpu_by_container(), "cpu_mili")
        out.update(flatten(self.fetch_memory_by_container(), "mem_MB"))
        return out


def flatten(data, metric):
    res = {}
    for group, summary in data.items():
        # remove - and istio- from group
        grp = group.replace("istio-", "")
        grp = grp.replace("-", "_")
        grp = grp.replace("/", "_")
        res[metric + "_min_" + grp] = summary[0]
        res[metric + "_avg_" + grp] = summary[1]
        res[metric + "_max_" + grp] = summary[2]

    return res

# convert float bytes to in megabytes


def to_megaBytes(m):
    return int(m / (1024 * 1024))

# convert float cpus to int mili cpus


def to_miliCpus(c):
    return int(c * 1000.0)

DEPL_MAP = {
    "fortioserver": "fortioserver_deployment",
    "fortioclient": "fortio_deployment"
}

# returns deployment_name/container_name


def metric_by_deployment_by_container(metric):
    depl = metric_by_deployment(metric)
    if depl is None:
        return None

    mapped_name = depl
    if depl in DEPL_MAP:
        mapped_name = DEPL_MAP[depl]
    return mapped_name + "/" + metric['container_name']

# These deployments have columns in the table, so only these are watched.
Watched_Deployments = set(["istio-pilot", "istio-telemetry",
                           "istio-policy", "fortioserver", "fortioclient"])

# returns deployment_name


def metric_by_deployment(metric):
    depl = metric['pod_name'].rsplit('-', 2)[0]
    if depl not in Watched_Deployments:
        return None

    return depl


def computeMinMaxAvg(d, groupby=None, xform=None):
    if d['status'] != "success":
        raise Exception("command not successful: " + d['status'] + str(d))

    if d['data']['resultType'] != "matrix":
        raise Exception("resultType not matrix: " + d['data']['resultType'])

    """
    for res in d['data']['result']:
        values = [float(v[1]) for v in res['values']]
        res['values'] = ( min(values), sum(values)/len(values), max(values), len(values))
    """

    ret = collections.defaultdict(list)

    for result in d['data']['result']:
        group = result['metric']['name']
        if groupby is not None:
            group = groupby(result['metric'])
            if group is None:
                continue

        ret[group].append(result)

    summary = {}

    for group, lst in ret.items():
        values = [float(v[1]) for v in lst[0]['values']]
        for l in lst[1:]:
            v = l['values']
            for idx in range(len(values)):
                values[idx] += float(v[idx][1])

        s = (min(values), sum(values) / len(values), max(values), len(values))
        if xform is not None:
            s = (xform(s[0]), xform(s[1]), xform(s[2]), s[3])

        summary[group] = s

    return summary


def main(argv):
    args = getParser().parse_args(argv)
    p = Prom(args.url, args.nseconds, end=args.end, host=args.host)
    out = p.fetch_cpu_and_mem()
    print json.dumps(out)


def getParser():
    parser = argparse.ArgumentParser(
        "Fetch cpu and memory stats from prometheus")
    parser.add_argument("url", help="prometheus base url")
    parser.add_argument(
        "nseconds", help="duration in seconds of the extract", type=int)
    parser.add_argument(
        "--end", help="relative time in seconds from now to end collection", type=int, default=0)
    parser.add_argument(
        "--host", help="host header when collection is thru ingress", default=None)
    return parser

if __name__ == "__main__":
    import sys
    sys.exit(main(sys.argv[1:]))
