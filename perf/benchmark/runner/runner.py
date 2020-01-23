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

import collections
import os
import json
import argparse
import subprocess
import shlex
import uuid
import sys
from subprocess import getoutput
from urllib.parse import urlparse
import yaml
from fortio import METRICS_START_SKIP_DURATION, METRICS_END_SKIP_DURATION


POD = collections.namedtuple('Pod', ['name', 'namespace', 'ip', 'labels'])


def pod_info(filterstr="", namespace=os.environ.get("NAMESPACE", "twopods"), multi_ok=True):
    cmd = "kubectl -n {namespace} get pod {filterstr}  -o json".format(
        namespace=namespace, filterstr=filterstr)
    op = getoutput(cmd)
    o = json.loads(op)
    items = o['items']

    if not multi_ok and len(items) > 1:
        raise Exception("more than one found " + op)

    if not items:
        raise Exception("no pods found with command [" + cmd + "]")

    i = items[0]
    return POD(i['metadata']['name'], i['metadata']['namespace'],
               i['status']['podIP'], i['metadata']['labels'])


def run_command(command):
    process = subprocess.Popen(shlex.split(command))
    process.wait()


def run_command_sync(command):
    op = getoutput(command)
    return op.strip()


class Fortio:
    ports = {
        "http": {"direct_port": 8077, "port": 8080},
        "grpc": {"direct_port": 8076, "port": 8079},
        "direct_envoy": {"direct_port": 8076, "port": 8079},
    }

    def __init__(
            self,
            conn=None,
            qps=None,
            duration=None,
            size=None,
            mode="http",
            telemetry_mode="mixer",
            perf_record=False,
            server="fortioserver",
            client="fortioclient",
            additional_args=None,
            filter_fn=None,
            extra_labels=None,
            baseline=False,
            serversidecar=False,
            clientsidecar=False,
            bothsidecar=True,
            ingress=None,
            mesh="istio",
            cacert=None):
        self.run_id = str(uuid.uuid4()).partition('-')[0]
        self.conn = conn
        self.qps = qps
        self.size = size
        self.duration = duration
        self.mode = mode
        self.ns = os.environ.get("NAMESPACE", "twopods")
        # bucket resolution in seconds
        self.r = "0.00005"
        self.telemetry_mode = telemetry_mode
        self.perf_record = perf_record
        self.server = pod_info("-lapp=" + server, namespace=self.ns)
        self.client = pod_info("-lapp=" + client, namespace=self.ns)
        self.additional_args = additional_args
        self.filter_fn = filter_fn
        self.extra_labels = extra_labels
        self.run_baseline = baseline
        self.run_serversidecar = serversidecar
        self.run_clientsidecar = clientsidecar
        self.run_bothsidecar = bothsidecar
        self.run_ingress = ingress
        self.cacert = cacert

        if mesh == "linkerd":
            self.mesh = "linkerd"
        elif mesh == "istio":
            self.mesh = "istio"
        else:
            sys.exit("invalid mesh %s, must be istio or linkerd" % mesh)

    def nosidecar(self, fortio_cmd):
        basestr = "http://{svc}:{port}/echo?size={size}"
        if self.mode == "grpc":
            basestr = "-payload-size {size} {svc}:{port}"
        return fortio_cmd + "_base " + basestr.format(
            svc=self.server.ip, port=self.ports[self.mode]["direct_port"], size=self.size)

    def serversidecar(self, fortio_cmd):
        basestr = "http://{svc}:{port}/echo?size={size}"
        if self.mode == "grpc":
            basestr = "-payload-size {size} {svc}:{port}"
        return fortio_cmd + "_serveronly " + basestr.format(
            svc=self.server.ip, port=self.ports[self.mode]["port"], size=self.size)

    def clientsidecar(self, fortio_cmd):
        basestr = "http://{svc}:{port}/echo?size={size}"
        if self.mode == "grpc":
            basestr = "-payload-size {size} {svc}:{port}"
        return fortio_cmd + "_clientonly " + basestr.format(
            svc=self.server.labels["app"], port=self.ports[self.mode]["direct_port"], size=self.size)

    def bothsidecar(self, fortio_cmd):
        basestr = "http://{svc}:{port}/echo?size={size}"
        if self.mode == "grpc":
            basestr = "-payload-size {size} {svc}:{port}"
        return fortio_cmd + "_both " + basestr.format(
            svc=self.server.labels["app"], port=self.ports[self.mode]["port"], size=self.size)

    def ingress(self, fortio_cmd):
        url = urlparse(self.run_ingress)
        # If scheme is not defined fallback to http
        if url.scheme == "":
            url = urlparse("http://{svc}".format(svc=self.run_ingress))

        return fortio_cmd + "_ingress {url}/echo?size={size}".format(
            url=url.geturl(), size=self.size)

    def run(self, conn, qps, size, duration):
        size = size or self.size
        if duration is None:
            duration = self.duration

        labels = self.run_id
        labels += "_qps_" + str(qps)
        labels += "_c_" + str(conn)
        labels += "_" + str(size)
        # Mixer label
        if self.mesh == "istio":
            labels += "_"
            labels += self.telemetry_mode
        elif self.mesh == "linkerd":
            labels += "_"
            labels += "linkerd"

        if self.extra_labels is not None:
            labels += "_" + self.extra_labels

        grpc = ""
        if self.mode == "grpc":
            grpc = "-grpc -ping"

        cacert_arg = ""
        if self.cacert is not None:
            cacert_arg = "-cacert {cacert_path}".format(cacert_path=self.cacert)

        fortio_cmd = (
            "fortio load -c {conn} -qps {qps} -t {duration}s -a -r {r} {cacert_arg} {grpc} -httpbufferkb=128 " +
            "-labels {labels}").format(
                conn=conn,
                qps=qps,
                duration=duration,
                r=self.r,
                grpc=grpc,
                cacert_arg=cacert_arg,
                labels=labels)

        if self.run_ingress:
            print('-------------- Running in ingress mode --------------')
            kubectl_exec(self.client.name, self.ingress(fortio_cmd))
            if self.perf_record:
                run_perf(
                    self.mesh,
                    self.server.name,
                    labels + "_srv_ingress",
                    duration=40)

        if self.run_serversidecar:
            print('-------------- Running in server sidecar mode --------------')
            kubectl_exec(self.client.name, self.serversidecar(fortio_cmd))
            if self.perf_record:
                run_perf(
                    self.mesh,
                    self.server.name,
                    labels + "_srv_serveronly",
                    duration=40)

        if self.run_clientsidecar:
            print('-------------- Running in client sidecar mode --------------')
            kubectl_exec(self.client.name, self.clientsidecar(fortio_cmd))
            if self.perf_record:
                run_perf(
                    self.mesh,
                    self.client.name,
                    labels + "_srv_clientonly",
                    duration=40)

        if self.run_bothsidecar:
            print('-------------- Running in both sidecar mode --------------')
            kubectl_exec(self.client.name, self.bothsidecar(fortio_cmd))
            if self.perf_record:
                run_perf(
                    self.mesh,
                    self.server.name,
                    labels + "_srv_bothsidecars",
                    duration=40)

        if self.run_baseline:
            print('-------------- Running in baseline mode --------------')
            kubectl_exec(self.client.name, self.nosidecar(fortio_cmd))


PERFCMD = "/usr/lib/linux-tools/4.4.0-131-generic/perf"
FLAMESH = "flame.sh"
PERFSH = "get_perfdata.sh"
PERFWD = "/etc/istio/proxy/"

WD = os.getcwd()
LOCAL_FLAMEDIR = os.path.join(WD, "../flame/")
LOCAL_FLAMEPATH = LOCAL_FLAMEDIR + FLAMESH
LOCAL_PERFPATH = LOCAL_FLAMEDIR + PERFSH
LOCAL_FLAMEOUTPUT = LOCAL_FLAMEDIR + "flameoutput/"


def run_perf(mesh, pod, labels, duration=20):
    filename = labels + "_perf.data"
    filepath = PERFWD + filename
    perfpath = PERFWD + PERFSH

    # copy executable over
    kubectl_cp(LOCAL_PERFPATH, pod + ":" + perfpath, mesh + "-proxy")

    kubectl_exec(
        pod,
        "{perf_cmd} {filename} {duration}".format(
            perf_cmd=perfpath,
            filename=filename,
            duration=duration),
        container=mesh + "-proxy")

    kubectl_cp(pod + ":" + filepath + ".perf", LOCAL_FLAMEOUTPUT + filename + ".perf", mesh + "-proxy")
    run_command_sync(LOCAL_FLAMEPATH + " " + filename + ".perf")


def kubectl_cp(from_file, to_file, container):
    namespace = os.environ.get("NAMESPACE", "twopods")
    cmd = "kubectl --namespace {namespace} cp {from_file} {to_file} -c {container}".format(
        namespace=namespace,
        from_file=from_file,
        to_file=to_file,
        container=container)
    print(cmd, flush=True)
    run_command_sync(cmd)


def kubectl_exec(pod, remote_cmd, runfn=run_command, container=None):
    namespace = os.environ.get("NAMESPACE", "twopods")
    c = ""
    if container is not None:
        c = "-c " + container
    cmd = "kubectl --namespace {namespace} exec {pod} {c} -- {remote_cmd}".format(
        pod=pod,
        remote_cmd=remote_cmd,
        c=c,
        namespace=namespace)
    print(cmd, flush=True)
    runfn(cmd)


def rc(command):
    process = subprocess.Popen(command.split(), stdout=subprocess.PIPE)
    while True:
        output = process.stdout.readline()
        if output == '' and process.poll() is not None:
            break
        if output:
            print(output.strip() + "\n")
    return process.poll()


def validate(job_config):
    required_fields = {"conn": list, "qps": list, "duration": int}
    for k in required_fields:
        if k not in job_config:
            print("missing required parameter {}".format(k))
            return False
        exp_type = required_fields[k]
        if not isinstance(job_config[k], exp_type):
            print("expecting type of parameter {} to be {}, got {}".format(k, exp_type, type(job_config[k])))
            return False
    return True


def fortio_from_config_file(args):
    with open(args.config_file) as f:
        job_config = yaml.safe_load(f)
        if not validate(job_config):
            exit(1)
        # TODO: hard to parse yaml into object directly because of existing constructor from CLI
        fortio = Fortio()
        fortio.conn = job_config.get('conn', 16)
        fortio.qps = job_config.get('qps', 1000)
        fortio.duration = job_config.get('duration', 240)
        fortio.telemetry_mode = job_config.get('telemetry_mode', 'mixer')
        fortio.metrics = job_config.get('metrics', 'p90')
        fortio.size = job_config.get('size', 1024)
        fortio.perf_record = job_config.get('perf_record', False)
        fortio.run_serversidecar = job_config.get('run_serversidecar', False)
        fortio.run_clientsidecar = job_config.get('run_clientsidecar', False)
        fortio.run_bothsidecar = job_config.get('run_bothsidecar', False)
        fortio.run_baseline = job_config.get('run_baseline', True)
        fortio.mesh = job_config.get('mesh', 'istio')
        fortio.mode = job_config.get('mode', 'http')
        fortio.extra_labels = job_config.get('extra_labels')

        return fortio


def run(args):
    min_duration = METRICS_START_SKIP_DURATION + METRICS_END_SKIP_DURATION

    # run with config files
    if args.config_file is not None:
        fortio = fortio_from_config_file(args)
    else:
        fortio = Fortio(
            conn=args.conn,
            qps=args.qps,
            duration=args.duration,
            size=args.size,
            perf_record=args.perf,
            extra_labels=args.extra_labels,
            baseline=args.baseline,
            serversidecar=args.serversidecar,
            clientsidecar=args.clientsidecar,
            bothsidecar=args.bothsidecar,
            ingress=args.ingress,
            mode=args.mode,
            mesh=args.mesh,
            telemetry_mode=args.telemetry_mode,
            cacert=args.cacert)

    if fortio.duration <= min_duration:
        print("Duration must be greater than {min_duration}".format(
            min_duration=min_duration))
        exit(1)

    for conn in fortio.conn:
        for qps in fortio.qps:
            fortio.run(conn=conn, qps=qps,
                       duration=fortio.duration, size=fortio.size)


def csv_to_int(s):
    return [int(i) for i in s.split(",")]


def get_parser():
    parser = argparse.ArgumentParser("Run performance test")
    parser.add_argument(
        "--conn",
        help="number of connections, comma separated list",
        type=csv_to_int,)
    parser.add_argument(
        "--qps",
        help="qps, comma separated list",
        type=csv_to_int,)
    parser.add_argument(
        "--duration",
        help="duration in seconds of the extract",
        type=int)
    parser.add_argument(
        "--size",
        help="size of the payload",
        type=int,
        default=1024)
    parser.add_argument(
        "--mesh",
        help="istio or linkerd",
        default="istio")
    parser.add_argument(
        "--telemetry_mode",
        help="run with different mixer configurations: mixer, none, telemetryv2",
        default="mixer")
    parser.add_argument(
        "--client",
        help="where to run the test from",
        default=None)
    parser.add_argument(
        "--server",
        help="pod ip of the server",
        default=None)
    parser.add_argument(
        "--perf",
        help="also run perf and produce flame graph",
        default=False)
    parser.add_argument(
        "--ingress",
        help="run traffic through ingress, should be a valid URL",
        default=None)
    parser.add_argument(
        "--extra_labels",
        help="extra labels",
        default=None)
    parser.add_argument(
        "--mode",
        help="http or grpc",
        default="http")
    parser.add_argument(
        "--config_file",
        help="config yaml file",
        default=None)
    parser.add_argument(
        "--cacert",
        help="path to the cacert for the fortio client inside the container",
        default=None)

    define_bool(parser, "baseline", "run baseline for all", False)
    define_bool(parser, "serversidecar",
                "run serversidecar-only for all", False)
    define_bool(parser, "clientsidecar",
                "run clientsidecar-only for all", False)
    define_bool(parser, "bothsidecar",
                "run clientsiecar and serversidecar for all", True)

    return parser


def define_bool(parser, opt, help_arg, default_val):
    parser.add_argument(
        "--" + opt, help=help_arg, dest=opt, action='store_true')
    parser.add_argument(
        "--no_" + opt, help="do not " + help_arg, dest=opt, action='store_false')
    val = {opt: default_val}
    parser.set_defaults(**val)


def main(argv):
    args = get_parser().parse_args(argv)
    print(args)
    return run(args)


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
