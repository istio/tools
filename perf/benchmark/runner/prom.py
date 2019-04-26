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


class Prom(object):
    # url: base url for prometheus
    #

    def __init__(self, url, nseconds, end=None, host=None, start=None, aggregate=True):
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

    def fetch(self, query, groupby=None, xform=None):
        resp = requests.get(self.url + "/api/v1/query_range", {
            "query": query,
            "start": self.start,
            "end": self.end,
            "step": "15"
        }, headers=self.headers)

        if not resp.ok:
            raise Exception(str(resp.text))

        data = resp.json()
        return computeMinMaxAvg(data, groupby=groupby, xform=xform, aggregate=self.aggregate)

    def fetch_cpu_by_container(self):
        return self.fetch(
            'irate(container_cpu_usage_seconds_total{container_name=~"mixer|policy|discovery|istio-proxy|captured|uncaptured"}[1m])',
            metric_by_deployment_by_container,
            to_miliCpus)

    def fetch_memory_by_container(self):
        return self.fetch(
            'container_memory_usage_bytes{container_name=~"mixer|policy|discovery|istio-proxy|captured|uncaptured"}',
            metric_by_deployment_by_container,
            to_megaBytes)

    def fetch_cpu_and_mem(self):
        out = flatten(self.fetch_cpu_by_container(),
                      "cpu_mili", aggregate=self.aggregate)
        out.update(flatten(self.fetch_memory_by_container(),
                           "mem_MB", aggregate=self.aggregate))
        return out

    def fetch_by_query(self, query):
        resp = requests.get(self.url + "/api/v1/query_range", {
            "query": query,
            "start": self.start,
            "end": self.end,
            "step": str(self.nseconds)
        }, headers=self.headers)

        if not resp.ok:
            raise Exception(str(resp))

        return resp.json()

    def fetch_num_requests_by_response_code(self, code):
        data = self.fetch_by_query(
            'sum(rate(istio_requests_total{reporter="destination", response_code="' + str(code) + '"}[' + str(self.nseconds) + 's]))')
        if len(data["data"]["result"]) > 0:
            return data["data"]["result"][0]["values"]
        return []

    def fetch_500s_and_400s(self):
        res = {}
        data_404 = self.fetch_num_requests_by_response_code(404)
        data_503 = self.fetch_num_requests_by_response_code(503)
        data_504 = self.fetch_num_requests_by_response_code(504)
        if len(data_404) > 0:
            res["istio_requests_total_404"] = data_404[-1][1]
        else:
            res["istio_requests_total_404"] = "0"
        if len(data_503) > 0:
            res["istio_requests_total_503"] = data_503[-1][1]
        else:
            res["istio_requests_total_503"] = "0"
        if len(data_504) > 0:
            res["istio_requests_total_504"] = data_504[-1][1]
        else:
            res["istio_requests_total_504"] = "0"
        return res

    def fetch_mixer_rules_count_by_metric_name(self, metric_name):
        data = self.fetch_by_query(
            'scalar(topk(1, max(' + metric_name + ') by (configID)))')
        if len(data["data"]["result"]) > 0:
            return data["data"]["result"][0]["values"]
        return []

    def fetch_mixer_rules_count(self):
        res = {}
        config_count = self.fetch_mixer_rules_count_by_metric_name(
            "mixer_config_rule_config_count")
        config_error_count = self.fetch_mixer_rules_count_by_metric_name(
            "mixer_config_rule_config_error_count")
        config_match_error_count = self.fetch_mixer_rules_count_by_metric_name(
            "mixer_config_rule_config_match_error_count")
        unsatisfied_action_handler_count = self.fetch_mixer_rules_count_by_metric_name(
            "mixer_config_unsatisfied_action_handler_count")
        instance_count = self.fetch_mixer_rules_count_by_metric_name(
            "mixer_config_instance_config_count")
        handler_count = self.fetch_mixer_rules_count_by_metric_name(
            "mixer_config_handler_config_count")
        attribute_count = self.fetch_mixer_rules_count_by_metric_name(
            "mixer_config_attribute_count")

        if len(config_count) > 0:
            res["mixer_config_rule_config_count"] = config_count[-1][1]
        else:
            res["mixer_config_rule_config_count"] = "0"
        if len(config_error_count) > 0:
            res["mixer_config_rule_config_error_count"] = config_error_count[-1][1]
        else:
            res["mixer_config_rule_config_error_count"] = "0"
        if len(config_match_error_count) > 0:
            res["mixer_config_rule_config_match_error_count"] = config_match_error_count[-1][1]
        else:
            res["mixer_config_rule_config_match_error_count"] = "0"
        if len(unsatisfied_action_handler_count) > 0:
            res["mixer_config_unsatisfied_action_handler_count"] = unsatisfied_action_handler_count[-1][1]
        else:
            res["mixer_config_unsatisfied_action_handler_count"] = "0"
        if len(instance_count) > 0:
            res["mixer_config_instance_config_count"] = instance_count[-1][1]
        else:
            res["mixer_config_instance_config_count"] = "0"
        if len(handler_count) > 0:
            res["mixer_config_handler_config_count"] = handler_count[-1][1]
        else:
            res["mixer_config_handler_config_count"] = "0"
        if len(attribute_count) > 0:
            res["mixer_config_attribute_count"] = attribute_count[-1][1]
        else:
            res["mixer_config_attribute_count"] = "0"

        return res

    def fetch_sum_by_metric_name(self, metric, groupby=None):
        query = 'sum(rate(' + metric + '[' + str(self.nseconds) + 's]))'
        if groupby is not None:
            query = query + ' by (' + groupby + ')'

        data = self.fetch_by_query(query)
        res = {}
        if len(data["data"]["result"]) > 0:
            if groupby is not None:
                for i in range(len(data["data"]["result"])):
                    key = data["data"]["result"][i]["metric"][groupby]
                    values = data["data"]["result"][i]["values"]
                    if len(values) > 0:
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
        if len(data["data"]["result"]) > 0:
            for i in range(len(data["data"]["result"])):
                key = data["data"]["result"][i]["metric"][groupby]
                values = data["data"]["result"][i]["values"]
                if len(values) > 0:
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
        if len(data["data"]["result"]) > 0:
            for i in range(len(data["data"]["result"])):
                print(data)
                key = data["data"]["result"][i]["metric"]["grpc_method"]
                values = data["data"]["result"][i]["values"]
                if len(values) > 0:
                    res["grpc_server_handled_total_5xx_" +
                        key] = values[-1][1]
                else:
                    res["grpc_server_handled_total_5xx_" + key] = "0"
        else:
            res["grpc_server_handled_total_5xx"] = "0"
        return res

    def fetch_non_successes_rate(self):
        query = 'sum(irate(grpc_server_handled_total{grpc_code!="OK",grpc_service=~".*Mixer"}[' + str(
            self.nseconds) + 's])) by (grpc_method)'
        data = self.fetch_by_query(query)
        res = {}
        if len(data["data"]["result"]) > 0:
            for i in range(len(data["data"]["result"])):
                key = data["data"]["result"][i]["metric"]["groupby"]
                values = data["data"]["result"][i]["values"]
                if len(values) > 0:
                    res["grpc_server_handled_total_4xx_" +
                        key] = values[-1][1]
                else:
                    res["grpc_server_handled_total_4xx_" + key] = "0"
        else:
            res["grpc_server_handled_total_4xx"] = "0"
        return res

    def fetch_mixer_traffic_overview(self):
        res = {}
        total_mixer_call = self.fetch_sum_by_metric_name(
            "grpc_server_handled_total")
        res.update(total_mixer_call)

        total_mixer_call_by_method = self.fetch_sum_by_metric_name(
            "grpc_server_handled_total", "grpc_method")
        res.update(total_mixer_call_by_method)

        response_durations_5 = self.fetch_histogram_by_metric_name(
            "grpc_server_handling_seconds_bucket", "0.5", "grpc_method")
        res.update(response_durations_5)

        response_durations_9 = self.fetch_histogram_by_metric_name(
            "grpc_server_handling_seconds_bucket", "0.9", "grpc_method")
        res.update(response_durations_9)

        response_durations_99 = self.fetch_histogram_by_metric_name(
            "grpc_server_handling_seconds_bucket", "0.99", "grpc_method")
        res.update(response_durations_99)

        server_error_rate_5xx = self.fetch_server_error_rate()
        res.update(server_error_rate_5xx)

        non_successes_rate_4xx = self.fetch_non_successes_rate()
        res.update(non_successes_rate_4xx)
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
                           "istio-policy", "fortioserver", "fortioclient", "istio-ingressgateway"])

# returns deployment_name


def metric_by_deployment(metric):
    depl = metric['pod_name'].rsplit('-', 2)[0]
    if depl not in Watched_Deployments:
        return None

    return depl


def computeMinMaxAvg(d, groupby=None, xform=None, aggregate=True):
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
                except IndexError as err:
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
    args = getParser().parse_args(argv)
    p = Prom(args.url, args.nseconds, end=args.end,
             host=args.host, aggregate=args.aggregate)
    out = p.fetch_cpu_and_mem()
    resp_out = p.fetch_500s_and_400s()
    out.update(resp_out)
    mixer_rules = p.fetch_mixer_rules_count()
    out.update(mixer_rules)
    mixer_overview = p.fetch_mixer_traffic_overview()
    out.update(mixer_overview)
    indent = None
    if args.indent is not None:
        indent = int(args.indent)

    print(json.dumps(out, indent=indent))


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
