from __future__ import print_function
import datetime
import calendar
import requests
import collections
import os
import json
import argparse
import subprocess
import shlex
import uuid
import time

POD = collections.namedtuple('Pod', ['name', 'namespace', 'ip', 'labels'])


def pod_info(filterstr="", namespace="service-graph", multi_ok=True):
    cmd =  "kubectl -n {namespace} get pod {filterstr}  -o json".format(namespace=namespace, filterstr=filterstr)
    op = subprocess.check_output(shlex.split(cmd))
    o = json.loads(op)
    items = o['items']

    if not multi_ok and len(items) > 1:
        raise Exception("more than one found " + op)

    if len(items) < 1:
        raise Exception("no pods found with command [" + cmd + "]")

    i = items[0]
    return POD(i['metadata']['name'], i['metadata']['namespace'],
               i['status']['podIP'], i['metadata']['labels'])


def run_command(command):
    process = subprocess.Popen(shlex.split(command))
    return process


def run_command_sync(command):
    op = subprocess.check_output(command, shell=True)
    return op.strip()


class Fortio(object):
    ports = {
        "http": {"direct_port": 8077, "port": 8080},
        "grpc": {"direct_port": 8076, "port": 8079},
        "direct_envoy": {"direct_port": 8076, "port": 8079}
    }

    def __init__(self, conn=None, qps=None, size=None, mode="http", duration=240, mixer=True, perf_record=False,
                 mixer_cache=True, server="fortioserver", client="fortioclient", additional_args=None, filterFn=None, labels=None,
                 baseline=False, serversidecar=True, clientsidecar=False):
        self.runid = str(uuid.uuid4()).partition('-')[0]
        self.conn = conn
        self.qps = qps
        self.duration = duration
        self.mode = mode
        self.size = size
        self.ns = os.environ.get("NAMESPACE", "service-graph")
        # bucket resolution in seconds
        self.r = "0.00005"
        self.mixer = mixer
        self.mixer_cache = mixer_cache
        self.additional_args = additional_args
        self.filterFn = filterFn
        self.perf_record = perf_record
        self.server = pod_info("-lapp=" + server, namespace=self.ns)
        self.client = pod_info("-lapp=" + client, namespace=self.ns)
        self.labels = labels
        self.run_serversidecar = serversidecar
        self.run_clientsidecar = clientsidecar
        self.run_baseline = baseline

    def nosidecar(self, fortio_cmd):
        return fortio_cmd + "_base http://{svc}:{port}/echo?size={size}".format(
            svc=self.server.ip, port=self.ports[self.mode]["direct_port"], size=self.size)

    def serversidecar(self, fortio_cmd):
        return fortio_cmd + "_serveronly http://{svc}:{port}/echo?size={size}".format(
            svc=self.server.ip, port=self.ports[self.mode]["port"], size=self.size)

    def bothsidecar(self, fortio_cmd):
        return fortio_cmd + "_both http://{svc}:{port}/echo?size={size}".format(
            svc=self.server.labels["app"], port=self.ports[self.mode]["port"], size=self.size)

    def run(self, conn=None, qps=None, size=None, duration=None):
        conn = conn or self.conn
        if qps is None:
            qps = self.qps
        size = size or self.size
        if duration is None:
            duration = self.duration

        labels = self.runid
        labels += "_qps_" + str(qps)
        labels += "_"
        labels += "c_" + str(conn)
        #labels += "_"
        #labels += "mixer" if self.mixer else "nomixer"
        #labels += "_"
        #labels += "mixercache" if self.mixer_cache else "nomixercache"
        labels += "_"
        labels += str(self.size)

        if self.labels is not None:
            labels += "_" + self.labels

        fortio_cmd = ("fortio load -c {conn} -qps {qps} -t {duration}s -a -r {r} -httpbufferkb=128 " +
                      "-labels {labels}").format(conn=conn, qps=qps, duration=duration, r=self.r, labels=labels)

        if self.run_serversidecar:
            p = kubectl(self.client.name, self.serversidecar(fortio_cmd))
            if self.perf_record:
                perf(self.server.name, labels + "_srv_serveronly", duration=40)
            p.wait()

        if self.run_clientsidecar:
            p = kubectl(self.client.name, self.bothsidecar(fortio_cmd))
            if self.perf_record:
                perf(self.server.name, labels +
                     "_srv_bothsidecars", duration=40)
            p.wait()

        if self.run_baseline:
            p = kubectl(self.client.name, self.nosidecar(fortio_cmd))
            p.wait()


PERFCMD = "/usr/lib/linux-tools/4.4.0-131-generic/perf"
PERFSH = "get_perfdata.sh"
PERFWD = "/etc/istio/proxy/"


def perf(pod, labels, duration=20, runfn=run_command_sync):
    filename = labels + "_perf.data"
    filepath = PERFWD + filename
    perfpath = PERFWD + PERFSH

    # copy executable over
    kubecp(PERFSH, pod + ":" + perfpath)

    perf = kubectl(pod,
                   "{perf_cmd} {filename} {duration}".format(perf_cmd=perfpath,
                                                             filename=filename, duration=duration),
                   runfn=run_command_sync, container="istio-proxy")

    print(perf)

    print(kubecp(pod + ":" + filepath + ".perf", filename + ".perf"))

    run_command_sync("./flame.sh " + filename + ".perf")
    return perf


def kubecp(from_file, to_file):
    namespace = os.environ.get("NAMESPACE", "service-graph")
    cmd = "kubectl --namespace {namespace} cp {from_file} {to_file} -c istio-proxy".format(
        from_file=from_file, to_file=to_file, namespace=namespace)
    print(cmd)
    return run_command_sync(cmd)


def kubectl(pod, remote_cmd, runfn=run_command, container=None):
    namespace = os.environ.get("NAMESPACE", "service-graph")
    c = ""
    if container is not None:
        c = "-c " + container
    cmd = "kubectl --namespace {namespace} exec -i -t {pod} {c} -- {remote_cmd}".format(
        pod=pod, remote_cmd=remote_cmd, c=c, namespace=namespace)
    print(cmd)
    return runfn(cmd)


def rc(command):
    process = subprocess.Popen(command.split(), stdout=subprocess.PIPE)
    while True:
        output = process.stdout.readline()
        if output == '' and process.poll() is not None:
            break
        if output:
            print(output.strip() + "\n")
    return process.poll()


def run(args):
    fortio = Fortio(size=args.size, duration=args.duration, perf_record=args.perf, labels=args.labels,
                    baseline=args.baseline, serversidecar=args.serversidecar, clientsidecar=args.clientsidecar)

    for conn in args.conn:
        for qps in args.qps:
            fortio.run(conn=conn, qps=qps)


def csv_to_int(s):
    return [int(i) for i in s.split(",")]


def getParser():
    parser = argparse.ArgumentParser("Run performance test")
    parser.add_argument(
        "conn", help="number of connections, comma separated list", type=csv_to_int)
    parser.add_argument(
        "qps", help="qps, comma separated list", type=csv_to_int)
    parser.add_argument(
        "duration", help="duration in seconds of the extract", type=int)
    parser.add_argument("--size", help="size of the payload",
                        type=int, default=1024)
    parser.add_argument(
        "--client", help="where to run the test from", default=None)
    parser.add_argument("--server", help="pod ip of the server", default=None)
    parser.add_argument("--perf", help="also run perf and produce flamegraph",
                        default=False, action='store_true')
    parser.add_argument(
        "--baseline", help="run baseline for all", type=bool, default=True)
    parser.add_argument(
        "--serversidecar", help="run serversidecar for all", type=bool, default=False)
    parser.add_argument(
        "--clientsidecar", help="run clientsidecar and serversidecar for all", type=bool, default=True)
    parser.add_argument("--labels", help="extra labels", default=None)
    return parser


def main(argv):
    args = getParser().parse_args(argv)
    print(args)
    return run(args)

if __name__ == "__main__":
    import sys
    sys.exit(main(sys.argv[1:]))
