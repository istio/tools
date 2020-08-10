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
import datetime
import calendar
import requests
import collections
import os
import json
import argparse
import logging
try:
    # Python 3
    import http.client as http_client
except ImportError:
    # Python 2
    import httplib as http_client

if os.environ.get("DEBUG", "0") != "0":
    http_client.HTTPConnection.debuglevel = 1
    logging.basicConfig()
    logging.getLogger().setLevel(logging.DEBUG)
    req_log = logging.getLogger('requests.packages.urllib3')
    req_log.setLevel(logging.DEBUG)
    req_log.propagate = True


def calculate_average(item, resource_type):
    data_points_list = item["values"]
    data_sum = 0
    for data_point in data_points_list:
        data_sum += float(data_point[1])
    data_avg = float(data_sum / len(data_points_list))
    if resource_type == "cpu":
        return to_mili_cpus(data_avg)
    else:
        return to_mega_bytes(data_avg)


def get_average_within_query_time_range(data, resource_type):
    val_by_pod_name = {"fortioclient": 0, "fortioserver": 0, "istio-ingressgateway": 0}
    if data["data"]["result"]:
        for item in data["data"]["result"]:
            pod_name = item["metric"]["pod"]
            if "fortioclient" in pod_name:
                val_by_pod_name["fortioclient"] = calculate_average(item, resource_type)
            if "fortioserver" in pod_name:
                val_by_pod_name["fortioserver"] = calculate_average(item, resource_type)
            if "istio-ingressgateway" in pod_name:
                val_by_pod_name["istio-ingressgateway"] = calculate_average(item, resource_type)
    return val_by_pod_name


class Prom:
    # url: base url for prometheus
    def __init__(
            self,
            url,
            nseconds,
            end=None,
            host=None,
            start=None,
            aggregate=True):
        self.url = url
        self.nseconds = nseconds
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
        self.aggregate = aggregate

    def fetch_by_query(self, query):
        resp = requests.get(self.url + "/api/v1/query_range", {
            "query": query,
            "start": self.start,
            "end": self.end,
            "step": 15
        }, headers=self.headers)

        if not resp.ok:
            raise Exception(str(resp))

        return resp.json()

    # We query from start_time to end_time and make 15s as the time interval to note down a data point
    # The function is to calculate the average of all the data points we get

    def fetch(self, query, groupby=None, xform=None):
        data = self.fetch_by_query(query)
        return compute_min_max_avg(
            data,
            groupby=groupby,
            xform=xform,
            aggregate=self.aggregate)

    def fetch_istio_proxy_cpu_usage_by_pod_name(self):
        cpu_query = 'sum(rate(container_cpu_usage_seconds_total{job="kubernetes-cadvisor",container="istio-proxy"}[1m])) by (pod)'
        data = self.fetch_by_query(cpu_query)
        avg_cpu_dict = get_average_within_query_time_range(data, "cpu")
        return avg_cpu_dict

    def fetch_istio_proxy_memory_usage_by_pod_name(self):
        mem_query = 'container_memory_usage_bytes{job = "kubernetes-cadvisor", container="istio-proxy"}'
        data = self.fetch_by_query(mem_query)
        avg_mem_dict = get_average_within_query_time_range(data, "mem")
        return avg_mem_dict

    def fetch_istio_proxy_cpu_and_mem(self):
        out = {}

        avg_cpu_dict = self.fetch_istio_proxy_cpu_usage_by_pod_name()
        out["cpu_mili_avg_istio_proxy_fortioclient"] = avg_cpu_dict["fortioclient"]
        out["cpu_mili_avg_istio_proxy_fortioserver"] = avg_cpu_dict["fortioserver"]
        out["cpu_mili_avg_istio_proxy_istio-ingressgateway"] = avg_cpu_dict["istio-ingressgateway"]

        avg_mem_dict = self.fetch_istio_proxy_memory_usage_by_pod_name()
        out["mem_Mi_avg_istio_proxy_fortioclient"] = avg_mem_dict["fortioclient"]
        out["mem_Mi_avg_istio_proxy_fortioserver"] = avg_mem_dict["fortioserver"]
        out["mem_Mi_avg_istio_proxy_istio-ingressgateway"] = avg_mem_dict["istio-ingressgateway"]

        return out

    def fetch_cpu_by_container(self):
        return self.fetch(
            'irate(container_cpu_usage_seconds_total{job="kubernetes-cadvisor",container=~"discovery|istio-proxy|captured|uncaptured"}[1m])',
            metric_by_deployment_by_container,
            to_mili_cpus)

    def fetch_memory_by_container(self):
        return self.fetch(
            'container_memory_usage_bytes{job="kubernetes-cadvisor",container=~"discovery|istio-proxy|captured|uncaptured"}',
            metric_by_deployment_by_container,
            to_mega_bytes)

    def fetch_cpu_and_mem(self):
        out = flatten(self.fetch_cpu_by_container(),
                      "cpu_mili", aggregate=self.aggregate)
        out.update(flatten(self.fetch_memory_by_container(),
                           "mem_Mi", aggregate=self.aggregate))
        return out

    def fetch_num_requests_by_response_code(self, code):
        data = self.fetch_by_query(
            'sum(rate(istio_requests_total{reporter="destination", response_code="' +
            str(code) +
            '"}[' +
            str(
                self.nseconds) +
            's]))')
        if data["data"]["result"]:
            return data["data"]["result"][0]["values"]
        return []

    def fetch_500s_and_400s(self):
        res = {}
        data_404 = self.fetch_num_requests_by_response_code(404)
        data_503 = self.fetch_num_requests_by_response_code(503)
        data_504 = self.fetch_num_requests_by_response_code(504)
        if data_404:
            res["istio_requests_total_404"] = data_404[-1][1]
        else:
            res["istio_requests_total_404"] = "0"
        if data_503:
            res["istio_requests_total_503"] = data_503[-1][1]
        else:
            res["istio_requests_total_503"] = "0"
        if data_504:
            res["istio_requests_total_504"] = data_504[-1][1]
        else:
            res["istio_requests_total_504"] = "0"
        return res

    def fetch_sum_by_metric_name(self, metric, groupby=None):
        query = 'sum(rate(' + metric + '[' + str(self.nseconds) + 's]))'
        if groupby is not None:
            query = query + ' by (' + groupby + ')'

        data = self.fetch_by_query(query)
        res = {}
        if data["data"]["result"]:
            if groupby is not None:
                for i in range(len(data["data"]["result"])):
                    key = data["data"]["result"][i]["metric"][groupby]
                    values = data["data"]["result"][i]["values"]
                    if values:
                        res[metric + "_" + key] = values[-1][1]
                    else:
                        res[metric + "_" + key] = "0"
            else:
                values = data["data"]["result"][0]["values"]
                res[metric] = values[-1][1]
        else:
            res[metric] = "0"
        return res

    def fetch_histogram_by_metric_name(self, metric, percent, groupby):
        query = 'histogram_quantile(' + percent + ', sum(rate(' + metric + \
            '{}[' + str(self.nseconds) + 's])) by (' + \
            groupby + ', le)) * 1000'

        data = self.fetch_by_query(query)
        res = {}
        if data["data"]["result"]:
            for i in range(len(data["data"]["result"])):
                key = data["data"]["result"][i]["metric"][groupby]
                values = data["data"]["result"][i]["values"]
                if values:
                    res[metric + "_" + percent + "_" +
                        key] = values[-1][1]
                else:
                    res[metric + "_" + percent + "_" + key] = "0"
        return res

    def fetch_server_error_rate(self):
        query = 'sum(rate(grpc_server_handled_total{grpc_code=~"Unknown|Unimplemented|Internal|DataLoss"}[' + str(
            self.nseconds) + 's])) by (grpc_method)'
        data = self.fetch_by_query(query)
        res = {}
        if data["data"]["result"]:
            for i in range(len(data["data"]["result"])):
                key = data["data"]["result"][i]["metric"]["grpc_method"]
                values = data["data"]["result"][i]["values"]
                if values:
                    res["grpc_server_handled_total_5xx_" +
                        key] = values[-1][1]
                else:
                    res["grpc_server_handled_total_5xx_" + key] = "0"
        else:
            res["grpc_server_handled_total_5xx"] = "0"
        return res


def flatten(data, metric, aggregate):
    res = {}
    for group, summary in data.items():
        # remove - and istio- from group
        grp = group.replace("istio-", "")
        grp = grp.replace("-", "_")
        grp = grp.replace("/", "_")
        if aggregate:
            res[metric + "_min_" + grp] = summary[0]
            res[metric + "_avg_" + grp] = summary[1]
            res[metric + "_max_" + grp] = summary[2]
        else:
            res[metric + '_' + grp] = summary
    return res


# convert float bytes to in megabytes
def to_mega_bytes(mem):
    return float(mem / (1024 * 1024))


# convert float cpus to int mili cpus
def to_mili_cpus(cpu):
    return float(cpu * 1000.0)


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
    return mapped_name + "/" + metric['container']


# These deployments have columns in the table, so only these are watched.
Watched_Deployments = set(["istio-pilot",
                           "fortioserver",
                           "fortioclient",
                           "istio-ingressgateway"])


# returns deployment_name
def metric_by_deployment(metric):
    depl = metric['pod'].rsplit('-', 2)[0]
    if depl not in Watched_Deployments:
        return None
    return depl


def compute_min_max_avg(d, groupby=None, xform=None, aggregate=True):
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
                try:
                    values[idx] += float(v[idx][1])
                except IndexError:
                    # Data about that time is not yet populated.
                    break
        if aggregate:
            s = (min(values), sum(values) /
                 len(values), max(values), len(values))
            if xform is not None:
                s = (xform(s[0]), xform(s[1]), xform(s[2]), s[3])
        else:
            s = [xform(i) for i in values]
        summary[group] = s
    return summary


def main(argv):
    args = get_parser().parse_args(argv)
    p = Prom(args.url, args.nseconds, end=args.end,
             host=args.host, aggregate=args.aggregate)
    out = p.fetch_cpu_and_mem()
    resp_out = p.fetch_500s_and_400s()
    out.update(resp_out)
    indent = None
    if args.indent is not None:
        indent = int(args.indent)

    print(json.dumps(out, indent=indent))


def get_parser():
    parser = argparse.ArgumentParser(
        "Fetch cpu and memory stats from prometheus")
    parser.add_argument("url", help="prometheus base url")
    parser.add_argument(
        "nseconds", help="duration in seconds of the extract", type=int)
    parser.add_argument(
        "--end",
        help="relative time in seconds from now to end collection",
        type=int,
        default=0)
    parser.add_argument(
        "--host",
        help="host header when collection is thru ingress",
        default=None)
    parser.add_argument(
        "--indent", help="pretty print json with indent", default=None)
    parser.add_argument('--aggregate', dest='aggregate', action='store_true')
    parser.add_argument('--no-aggregate', dest='aggregate',
                        action='store_false')
    parser.set_defaults(aggregate=True)

    return parser


if __name__ == "__main__":
    import sys
    sys.exit(main(sys.argv[1:]))
