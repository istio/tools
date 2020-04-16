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
from threading import Thread
from time import sleep
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


# kubeclt related helper funcs
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


class Fortio:
    ports = {
        "http": {"direct_port": 8077, "port": 8080},
        "grpc": {"direct_port": 8076, "port": 8079},
        "direct_envoy": {"direct_port": 8076, "port": 8079},
    }

    def __init__(
            self,
            headers=None,
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
            cacert=None,
            load_gen_type="fortio",
            custom_profiling_command=None,
            custom_profiling_name="default-profile"):
        self.run_id = str(uuid.uuid4()).partition('-')[0]
        self.headers = headers
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
        self.custom_profiling_command = custom_profiling_command
        self.custom_profiling_name = custom_profiling_name
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
        self.load_gen_type = load_gen_type

        if mesh == "linkerd":
            self.mesh = "linkerd"
        elif mesh == "istio":
            self.mesh = "istio"
        else:
            sys.exit("invalid mesh %s, must be istio or linkerd" % mesh)

    def compute_uri(self, svc, port_type):
        basestr = "http://{svc}:{port}/echo?size={size}"
        if self.mode == "grpc":
            basestr = "-payload-size {size} {svc}:{port}"
        return basestr.format(svc=svc, port=self.ports[self.mode][port_type], size=self.size)

    def nosidecar(self, fortio_cmd):
        return fortio_cmd + "_base " + self.compute_uri(self.server.ip, "direct_port")

    def serversidecar(self, fortio_cmd):
        return fortio_cmd + "_serveronly " + self.compute_uri(self.server.ip, "port")

    def clientsidecar(self, fortio_cmd):
        return fortio_cmd + "_clientonly " + self.compute_uri(self.server.labels["app"], "direct_port")

    def bothsidecar(self, fortio_cmd):
        return fortio_cmd + "_both " + self.compute_uri(self.server.labels["app"], "port")

    def ingress(self, fortio_cmd):
        url = urlparse(self.run_ingress)
        # If scheme is not defined fallback to http
        if url.scheme == "":
            url = urlparse("http://{svc}".format(svc=self.run_ingress))

        return fortio_cmd + "_ingress {url}/echo?size={size}".format(
            url=url.geturl(), size=self.size)

    def execute_sidecar_mode(self, sidecar_mode, load_gen_type, load_gen_cmd, sidecar_mode_func, labels, perf_label_suffix):
        print('-------------- Running in {sidecar_mode} mode --------------'.format(sidecar_mode=sidecar_mode))
        if load_gen_type == "fortio":
            kubectl_exec(self.client.name, sidecar_mode_func(load_gen_cmd))

        if self.perf_record and len(perf_label_suffix) > 0:
            run_perf(
                self.mesh,
                self.server.name,
                labels + perf_label_suffix,
                duration=40)

    def generate_test_labels(self, conn, qps, size):
        size = size or self.size
        labels = self.run_id
        labels += "_qps_" + str(qps)
        labels += "_c_" + str(conn)
        labels += "_" + str(size)

        if self.mesh == "istio":
            labels += "_"
            labels += self.telemetry_mode
        elif self.mesh == "linkerd":
            labels += "_"
            labels += "linkerd"

        if self.extra_labels is not None:
            labels += "_" + self.extra_labels

        return labels

    def generate_headers_cmd(self, headers):
        headers_cmd = ""
        if headers is not None:
            for header_val in headers.split(","):
                headers_cmd += "-H=" + header_val + " "

        return headers_cmd

    def generate_fortio_cmd(self, headers_cmd, conn, qps, duration, grpc, cacert_arg, labels):
        if duration is None:
            duration = self.duration

        fortio_cmd = (
            "fortio load {headers} -c {conn} -qps {qps} -t {duration}s -a -r {r} {cacert_arg} {grpc} "
            "-httpbufferkb=128 -labels {labels}").format(
            headers=headers_cmd,
            conn=conn,
            qps=qps,
            duration=duration,
            r=self.r,
            grpc=grpc,
            cacert_arg=cacert_arg,
            labels=labels)

        return fortio_cmd

    def run(self, headers, conn, qps, size, duration):
        labels = self.generate_test_labels(conn, qps, size)

        grpc = ""
        if self.mode == "grpc":
            grpc = "-grpc -ping"

        cacert_arg = ""
        if self.cacert is not None:
            cacert_arg = "-cacert {cacert_path}".format(cacert_path=self.cacert)

        headers_cmd = self.generate_headers_cmd(headers)
        fortio_cmd = self.generate_fortio_cmd(headers_cmd, conn, qps, duration, grpc, cacert_arg, labels)

        def run_profiling_in_background(exec_cmd, podname, filename_prefix, profiling_command, labels):
            filename = "{filename_prefix}-{labels}-{podname}".format(
                filename_prefix=filename_prefix, labels=labels, podname=podname)
            profiler_cmd = "{exec_cmd} \"{profiling_command} > {filename}.profile\"".format(
                profiling_command=profiling_command,
                exec_cmd=exec_cmd,
                filename=filename
            )
            # Run the profile collection tool, and wait for it to finish.
            process = subprocess.Popen(shlex.split(profiler_cmd))
            process.wait()
            # Next we feed the profiling data to the flamegraphing script.
            flamegraph_cmd = "{exec_cmd} \"./FlameGraph/flamegraph.pl --title='{profiling_command} Flame Graph'  < {filename}.profile > {filename}.svg\"".format(
                exec_cmd=exec_cmd,
                profiling_command=profiling_command,
                filename=filename
            )
            process = subprocess.Popen(shlex.split(flamegraph_cmd))
            process.wait()
            # Lastly copy the resulting flamegraph out of the container
            kubectl_cp(podname + ":{filename}.svg".format(filename=filename),
                       "flame/flameoutput/{filename}.svg".format(filename=filename), "perf")

        threads = []

        if self.custom_profiling_command:
            # We run any custom profiling command on both pods, as one runs on each node we're interested in.
            for pod in [self.client.name, self.server.name]:
                exec_cmd_on_pod = "kubectl exec -n {namespace} {podname} -c perf -it -- bash -c ".format(
                    namespace=os.environ.get("NAMESPACE", "twopods"),
                    podname=pod
                )
                
                # Wait for node_exporter to run, which indicates the profiling initialization container has finished initializing.
                # once the init probe is supported, move this to a http probe instead in fortio.yaml
                ne_pid = ""
                attempts = 0
                while ne_pid == "" and attempts < 60:
                    ne_pid = getoutput("{exec_cmd} \"pgrep 'node_exporter'\"".format(exec_cmd=exec_cmd_on_pod)).strip()
                    attempts = attempts + 1
                    print(".")
                    sleep(1)

                # Find side car process id's in case the profiling command needs it.
                sidecar_ppid = getoutput("{exec_cmd} \"pgrep -f 'pilot-agent proxy sidecar'\"".format(exec_cmd=exec_cmd_on_pod)).strip()
                sidecar_pid = getoutput("{exec_cmd} \"pgrep -P {sidecar_ppid}\"".format(exec_cmd=exec_cmd_on_pod, sidecar_ppid=sidecar_ppid)).strip()
                profiling_command = self.custom_profiling_command.format(
                    duration=self.duration, sidecar_pid=sidecar_pid)
                threads.append(Thread(target=run_profiling_in_background, args=[
                    exec_cmd_on_pod, pod, self.custom_profiling_name, profiling_command, labels]))

        for thread in threads:
            thread.start()

        if self.run_baseline:
            self.execute_sidecar_mode("baseline", self.load_gen_type, fortio_cmd, self.nosidecar, labels, "")

        if self.run_serversidecar:
            self.execute_sidecar_mode("server sidecar", self.load_gen_type, fortio_cmd, self.serversidecar, labels, "_srv_serveronly")

        if self.run_clientsidecar:
            self.execute_sidecar_mode("client sidecar", self.load_gen_type, fortio_cmd, self.clientsidecar, labels, "_srv_clientonly")

        if self.run_bothsidecar:
            self.execute_sidecar_mode("both sidecar", self.load_gen_type, fortio_cmd, self.bothsidecar, labels, "_srv_bothsidecars")

        if self.run_ingress:
            print('-------------- Running in ingress mode --------------')
            kubectl_exec(self.client.name, self.ingress(fortio_cmd))
            if self.perf_record:
                run_perf(
                    self.mesh,
                    self.server.name,
                    labels + "_srv_ingress",
                    duration=40)

        if len(threads) > 0:
            if self.custom_profiling_command:
                for thread in threads:
                    thread.join()
            print("background profiler thread finished - flamegraphs are available in flame/flameoutput")

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


def validate_job_config(job_config):
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
        if not validate_job_config(job_config):
            exit(1)
        # TODO: hard to parse yaml into object directly because of existing constructor from CLI
        fortio = Fortio()
        fortio.headers = job_config.get('headers', None)
        fortio.conn = job_config.get('conn', 16)
        fortio.qps = job_config.get('qps', 1000)
        fortio.duration = job_config.get('duration', 240)
        fortio.load_gen_type = job_config.get("load_gen_type", "fortio")
        fortio.telemetry_mode = job_config.get('telemetry_mode', 'mixer')
        fortio.metrics = job_config.get('metrics', 'p90')
        fortio.size = job_config.get('size', 1024)
        fortio.perf_record = False
        fortio.run_serversidecar = job_config.get('run_serversidecar', False)
        fortio.run_clientsidecar = job_config.get('run_clientsidecar', False)
        fortio.run_bothsidecar = job_config.get('run_bothsidecar', True)
        fortio.run_baseline = job_config.get('run_baseline', False)
        fortio.run_ingress = job_config.get('run_ingress', False)
        fortio.mesh = job_config.get('mesh', 'istio')
        fortio.mode = job_config.get('mode', 'http')
        fortio.extra_labels = job_config.get('extra_labels')

        return fortio


def run_perf_test(args):
    min_duration = METRICS_START_SKIP_DURATION + METRICS_END_SKIP_DURATION

    # run with config files
    if args.config_file is not None:
        fortio = fortio_from_config_file(args)
    else:
        fortio = Fortio(
            headers=args.headers,
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
            cacert=args.cacert,
            load_gen_type=args.load_gen_type,
            custom_profiling_command=args.custom_profiling_command,
            custom_profiling_name=args.custom_profiling_name)

    if fortio.duration <= min_duration:
        print("Duration must be greater than {min_duration}".format(
            min_duration=min_duration))
        exit(1)

    for conn in fortio.conn:
        for qps in fortio.qps:
            fortio.run(headers=fortio.headers, conn=conn, qps=qps,
                       duration=fortio.duration, size=fortio.size)


def csv_to_int(s):
    return [int(i) for i in s.split(",")]


def get_parser():
    parser = argparse.ArgumentParser("Run performance test")
    parser.add_argument(
        "--headers",
        help="a list of `header:value` should be separated by comma",
        default=None)
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
        "--custom_profiling_command",
        help="Run custom profiling commands on the nodes for the client and server, and produce a flamegraph based on their outputs. E.g. --custom_profiling_command=\"/usr/share/bcc/tools/profile -df 40\"",
        default=False)
    parser.add_argument(
        "--custom_profiling_name",
        help="Name to be added to the flamegraph resulting from --custom_profiling_command",
        default="default-profile")
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
    parser.add_argument(
        "--load_gen_type",
        help="fortio or nighthawk",
        default="fortio",
    )

    define_bool(parser, "baseline", "run baseline for all", False)
    define_bool(parser, "serversidecar",
                "run serversidecar-only for all", False)
    define_bool(parser, "clientsidecar",
                "run clientsidecar-only for all", False)
    define_bool(parser, "bothsidecar",
                "run both clientsiecar and serversidecar", True)

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
    return run_perf_test(args)


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
