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
import socket
import argparse
import subprocess
import shlex
import uuid
import sys
import tempfile
import time
import multiprocessing
from subprocess import getoutput
from urllib.parse import urlparse
import yaml
from fortio import METRICS_START_SKIP_DURATION, METRICS_END_SKIP_DURATION

NAMESPACE = os.environ.get("NAMESPACE", "twopods-istio")
NIGHTHAWK_GRPC_SERVICE_PORT_FORWARD = 9999
POD = collections.namedtuple('Pod', ['name', 'namespace', 'ip', 'labels'])
NIGHTHAWK_DOCKER_IMAGE = "envoyproxy/nighthawk-dev:59683b759eb8f8bd8cce282795c08f9e2b3313d4"
processes = []

def pod_info(filterstr="", namespace=NAMESPACE, multi_ok=True):
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
    cmd = "kubectl --namespace {namespace} cp {from_file} {to_file} -c {container}".format(
        namespace=NAMESPACE,
        from_file=from_file,
        to_file=to_file,
        container=container)
    print(cmd, flush=True)
    run_command_sync(cmd)


def kubectl_exec(pod, remote_cmd, runfn=run_command, container=None):
    c = ""
    if container is not None:
        c = "-c " + container
    cmd = "kubectl --namespace {namespace} exec {pod} {c} -- {remote_cmd}".format(
        pod=pod,
        remote_cmd=remote_cmd,
        c=c,
        namespace=NAMESPACE)
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
            frequency=None,
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
            load_gen_type="fortio"):
        self.run_id = str(uuid.uuid4()).partition('-')[0]
        self.headers = headers
        self.conn = conn
        self.qps = qps
        self.size = size
        self.duration = duration
        self.frequency = frequency
        self.mode = mode
        self.ns = NAMESPACE
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
        self.load_gen_type = load_gen_type

        if mesh == "linkerd":
            self.mesh = "linkerd"
        elif mesh == "istio":
            self.mesh = "istio"
        else:
            sys.exit("invalid mesh %s, must be istio or linkerd" % mesh)

    def get_protocol_uri_fragment(self):
        return "https" if self.mode == "grpc" else "http"

    def compute_uri(self, svc, port_type):
        if self.load_gen_type == "fortio":
            basestr = "http://{svc}:{port}/echo?size={size}"
            if self.mode == "grpc":
                basestr = "-payload-size {size} {svc}:{port}"
            return basestr.format(svc=svc, port=self.ports[self.mode][port_type], size=self.size)
        elif self.load_gen_type == "nighthawk":
            return "{protocol}://{svc}:{port}/".format(
                svc=svc, port=self.ports[self.mode][port_type], protocol=self.get_protocol_uri_fragment())
        else:
            sys.exit("invalid load generator %s, must be fortio or nighthawk", self.load_gen_type)

    # Baseline is no sidecar mode
    def baseline(self, load_gen_cmd, sidecar_mode):
        return load_gen_cmd + "_" + sidecar_mode + " " + self.compute_uri(self.server.ip, "direct_port")

    def serversidecar(self, load_gen_cmd, sidecar_mode):
        return load_gen_cmd + "_" + sidecar_mode + " " + self.compute_uri(self.server.ip, "port")

    def clientsidecar(self, load_gen_cmd, sidecar_mode):
        return load_gen_cmd + "_" + sidecar_mode + " " + self.compute_uri(self.server.labels["app"], "direct_port")

    def bothsidecar(self, load_gen_cmd, sidecar_mode):
        return load_gen_cmd + "_" + sidecar_mode + " " + self.compute_uri(self.server.labels["app"], "port")

    def ingress(self, load_gen_cmd):
        url = urlparse(self.run_ingress)
        # If scheme is not defined fallback to http
        if url.scheme == "":
            url = urlparse("http://{svc}".format(svc=self.run_ingress))

        return load_gen_cmd + "_ingress {url}/echo?size={size}".format(
            url=url.geturl(), size=self.size)

    def execute_sidecar_mode(self, sidecar_mode, load_gen_type, load_gen_cmd, sidecar_mode_func, labels, perf_label_suffix):
        print('-------------- Running in {sidecar_mode} mode --------------'.format(sidecar_mode=sidecar_mode))
        if load_gen_type == "fortio":
            print("p1 ++++++++++")
            p = multiprocessing.Process(target=kubectl_exec, args=[self.client.name, sidecar_mode_func(load_gen_cmd, sidecar_mode)])
            p.start()
            processes.append(p)
        elif load_gen_type == "nighthawk":
            run_nighthawk(self.client.name, sidecar_mode_func(load_gen_cmd, sidecar_mode), labels + "_" + sidecar_mode)

        if self.perf_record and len(perf_label_suffix) > 0:
            run_perf(
                self.server.name,
                labels + perf_label_suffix,
                duration=40,
                frequency=self.frequency)

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

    def generate_nighthawk_cmd(self, cpus, conn, qps, duration, labels):
        labels = "nighthawk_" + labels
        nighthawk_args = [
            "nighthawk_client",
            "--concurrency {cpus}",
            "--output-format json",
            "--prefetch-connections",
            "--open-loop",
            "--experimental-h1-connection-reuse-strategy lru",
            "--experimental-h2-use-multiple-connections",
            "--connections {conn}",
            "--burst-size {conn}",
            "--rps {qps}",
            "--duration {duration}",
            "--request-header \"x-nighthawk-test-server-config: {{response_body_size:{size}}}\""
        ]

        # Our "gRPC" mode actually means:
        #  - https (see get_protocol_uri_fragment())
        #  - h2
        #  - with long running connections
        #  - Also transfer request body sized according to "size".
        if self.mode == "grpc":
            nighthawk_args.append("--h2")
            if self.size:
                nighthawk_args.append(
                    "--request-header \"content-length: {size}\"")

        # Note: Labels is the last arg, and there's stuff depending on that.
        # watch out when moving it.
        nighthawk_args.append("--label {labels}")

        # As the worker count acts as a multiplier, we divide by qps/conn by the number of cpu's to spread load accross the workers so the sum of the workers will target the global qps/connection levels.
        nighthawk_cmd = " ".join(nighthawk_args).format(
            conn=round(conn / cpus),
            qps=round(qps / cpus),
            duration=duration,
            labels=labels,
            size=self.size,
            cpus=cpus,
            port_forward=NIGHTHAWK_GRPC_SERVICE_PORT_FORWARD)

        return nighthawk_cmd

    def run(self, headers, conn, qps, size, duration):
        labels = self.generate_test_labels(conn, qps, size)

        grpc = ""
        if self.mode == "grpc":
            grpc = "-grpc -ping"

        cacert_arg = ""
        if self.cacert is not None:
            cacert_arg = "-cacert {cacert_path}".format(cacert_path=self.cacert)

        headers_cmd = self.generate_headers_cmd(headers)

        load_gen_cmd = ""
        if self.load_gen_type == "fortio":
            load_gen_cmd = self.generate_fortio_cmd(headers_cmd, conn, qps, duration, grpc, cacert_arg, labels)
        elif self.load_gen_type == "nighthawk":
            # TODO(oschaaf): Figure out how to best determine the right concurrency for Nighthawk.
            # Results seem to get very noisy as the number of workers increases, are the clients
            # and running on separate sets of vCPU cores? nproc yields the same concurrency as goprocs
            # use with the Fortio version.
            # client_cpus = int(run_command_sync(
            #     "kubectl exec -n \"{ns}\" svc/fortioclient -c shell nproc".format(ns=NAMESPACE)))
            # print("Client pod has {client_cpus} cpus".format(client_cpus=client_cpus))

            # See the comment above, we restrict execution to a single nighthawk worker for
            # now to avoid noise.
            workers = 1
            load_gen_cmd = self.generate_nighthawk_cmd(workers, conn, qps, duration, labels)

        perf_label = ""
        sidecar_mode = ""
        sidecar_mode_func = None

        if self.run_baseline:
            sidecar_mode = "baseline"
            sidecar_mode_func = self.baseline
            self.execute_sidecar_mode(sidecar_mode, self.load_gen_type, load_gen_cmd, sidecar_mode_func, labels, perf_label)

        if self.run_serversidecar:
            perf_label = "_srv_serveronly"
            sidecar_mode = "serveronly"
            sidecar_mode_func = self.serversidecar
            self.execute_sidecar_mode(sidecar_mode, self.load_gen_type, load_gen_cmd, sidecar_mode_func, labels, perf_label)

        if self.run_clientsidecar:
            perf_label = "_srv_clientonly"
            sidecar_mode = "clientonly"
            sidecar_mode_func = self.clientsidecar
            self.execute_sidecar_mode(sidecar_mode, self.load_gen_type, load_gen_cmd, sidecar_mode_func, labels, perf_label)

        if self.run_bothsidecar:
            perf_label = "_srv_bothsidecars"
            sidecar_mode = "both"
            sidecar_mode_func = self.bothsidecar
            self.execute_sidecar_mode(sidecar_mode, self.load_gen_type, load_gen_cmd, sidecar_mode_func, labels, perf_label)

        if self.run_ingress:
            perf_label = "_srv_ingress"
            print('-------------- Running in ingress mode --------------')
            kubectl_exec(self.client.name, self.ingress(load_gen_cmd))
            if self.perf_record:
                run_perf(
                    self.mesh,
                    self.server.name,
                    labels + "_srv_ingress",
                    duration=40,
                    frequency=self.frequency)


WD = os.getcwd()
LOCAL_FLAMEDIR = os.path.join(WD, "../flame/")
PERF_PROXY_FILE = "get_proxy_perf.sh"
LOCAL_FLAME_PROXY_FILE_PATH = LOCAL_FLAMEDIR + PERF_PROXY_FILE
LOCAL_FLAMEOUTPUT = LOCAL_FLAMEDIR + "flameoutput/"


def run_perf(pod, labels, duration, frequency):
    if duration is None:
        duration = 40
    if frequency is None:
        frequency = 99
    os.environ["PERF_DATA_FILENAME"] = labels + "_perf.data"
    print(os.environ["PERF_DATA_FILENAME"])
    print("p2 ++++++++++")
    p = multiprocessing.Process(target=run_command_sync, args=[LOCAL_FLAME_PROXY_FILE_PATH + " -p {pod} -n {namespace} -d {duration} -f {frequency}".format(
        pod=pod, namespace=NAMESPACE, duration=duration, frequency=frequency)])
    p.start()
    processes.append(p)


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
        fortio.load_gen_type = os.environ.get("LOAD_GEN_TYPE", "fortio")
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


def can_connect_to_nighthawk_service():
    # TODO(oschaaf): re-instate going through the gRPC service.
    return True
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        return sock.connect_ex(('127.0.0.1', NIGHTHAWK_GRPC_SERVICE_PORT_FORWARD)) == 0


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
            load_gen_type=args.load_gen_type)

    if fortio.duration <= min_duration:
        print("Duration must be greater than {min_duration}".format(
            min_duration=min_duration))
        exit(1)

    port_forward_process = None

    if args.load_gen_type == "nighthawk":
        # Create a port_forward for accessing nighthawk_service.
        if not can_connect_to_nighthawk_service():
            popen_cmd = "kubectl -n \"{ns}\" port-forward svc/fortioclient {port}:9999".format(
                ns=NAMESPACE,
                port=NIGHTHAWK_GRPC_SERVICE_PORT_FORWARD)
            port_forward_process = subprocess.Popen(shlex.split(
                popen_cmd), stdout=subprocess.PIPE)
            max_tries = 10
            while max_tries > 0 and not can_connect_to_nighthawk_service():
                time.sleep(0.5)
                max_tries = max_tries - 1

        if not can_connect_to_nighthawk_service():
            print("Failure connecting to nighthawk_service")
            sys.exit(-1)
        else:
            print("Able to connect to nighthawk_service, proceeding")

    try:
        for conn in fortio.conn:
            for qps in fortio.qps:
                fortio.run(headers=fortio.headers, conn=conn, qps=qps,
                           duration=fortio.duration, size=fortio.size)
    finally:
        if port_forward_process is not None:
            port_forward_process.kill()


def run_nighthawk(pod, remote_cmd, labels):
    kube_cmd = "kubectl --namespace {namespace} exec {pod} -c captured -- {remote_cmd}".format(
        pod=pod,
        remote_cmd=remote_cmd,
        namespace=NAMESPACE)
    print("nighthawk commandline: " + kube_cmd)
    process = subprocess.Popen(shlex.split(kube_cmd), stdout=subprocess.PIPE)
    (output, err) = process.communicate()
    exit_code = process.wait()

    if exit_code == 0:
        with tempfile.NamedTemporaryFile(dir='/tmp', delete=True) as tmpfile:
            dest = tmpfile.name
            with open("%s.json" % dest, 'wb') as f:
                f.write(output)
            print("Dumped Nighthawk's json to {dest}".format(dest=dest))

            # Send human readable output to the command line.
            os.system(
                "cat {dest}.json | docker run -i --rm {docker_image} nighthawk_output_transform --output-format human".format(docker_image=NIGHTHAWK_DOCKER_IMAGE, dest=dest))
            # Transform to Fortio's reporting server json format
            os.system("cat {dest}.json | docker run -i --rm {docker_image} nighthawk_output_transform --output-format fortio > {dest}.fortio.json".format(
                dest=dest, docker_image=NIGHTHAWK_DOCKER_IMAGE))
            # Copy to the Fortio report server data directory.
            # TODO(oschaaf): We output the global aggregated statistics here of request_to_response, which excludes connection set up time.
            # It would be nice to dump a series instead, as we have more details available in the Nighthawk json:
            # - queue/connect time
            # - time spend blocking in closed loop mode
            # - initiation time to completion (spanning the complete lifetime of a request/reply, including queue/connect time)
            # - per worker output may sometimes help interpret plots that don't have a nice knee-shaped shape.
            kubectl_cp("{dest}.fortio.json".format(
                dest=dest), "{pod}:/var/lib/fortio/{datetime}_nighthawk_{labels}.json".format(pod=pod, labels=labels, datetime=time.strftime("%Y-%m-%d-%H%M%S")), "shell")
    else:
        print("nighthawk remote execution error: %s" % exit_code)
        if output:
            print("--> stdout: %s" % output.decode("utf-8"))
        if err:
            print("--> stderr: %s" % err.decode("utf-8"))


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
        "--frequency",
        help="sampling frequency of generating flame graph",
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
    return run_perf_test(args)


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
