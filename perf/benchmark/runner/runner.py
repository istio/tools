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

POD = collections.namedtuple('Pod', ['name', 'namespace', 'ip', 'labels'])


def pod_info(filterstr="", namespace="twopods", multi_ok=True):
    cmd = "kubectl -n {namespace} get pod {filterstr}  -o json".format(
        namespace=namespace, filterstr=filterstr)
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
        "http": {"direct_port": 8077, "port": 8080, "ingress": 80},
        "grpc": {"direct_port": 8076, "port": 8079, "ingress": 80},
        "direct_envoy": {"direct_port": 8076, "port": 8079},
    }

    def __init__(
            self,
            conn=None,
            qps=None,
            duration=None,
            size=None,
            mode="http",
            mixer=True,
            mixer_cache=True,
            perf_record=False,
            server="fortioserver",
            client="fortioclient",
            additional_args=None,
            filterFn=None,
            labels=None,
            baseline=False,
            serversidecar=False,
            clientsidecar=True,
            ingress=None,
            mesh="istio"):
        self.run_id = str(uuid.uuid4()).partition('-')[0]
        self.conn = conn
        self.qps = qps
        self.size = size
        self.duration = duration
        self.mode = mode
        self.ns = os.environ.get("NAMESPACE", "twopods")
        # bucket resolution in seconds
        self.r = "0.00005"
        self.mixer = mixer
        self.mixer_cache = mixer_cache
        self.perf_record = perf_record
        self.server = pod_info("-lapp=" + server, namespace=self.ns)
        self.client = pod_info("-lapp=" + client, namespace=self.ns)
        self.additional_args = additional_args
        self.filterFn = filterFn
        self.labels = labels
        self.run_baseline = baseline
        self.run_serversidecar = serversidecar
        self.run_clientsidecar = clientsidecar
        self.run_ingress = ingress

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

    def bothsidecar(self, fortio_cmd):
        basestr = "http://{svc}:{port}/echo?size={size}"
        if self.mode == "grpc":
            basestr = "-payload-size {size} {svc}:{port}"
        return fortio_cmd + "_both " + basestr.format(
            svc=self.server.labels["app"], port=self.ports[self.mode]["port"], size=self.size)

    def ingress(self, fortio_cmd):
        svc = self.run_ingress
        if ':' not in svc:
            svc += ":{port}".format(port=self.ports[self.mode]["ingress"])

        return fortio_cmd + "_ingress http://{svc}/echo?size={size}".format(
            svc=svc, size=self.size)

    def run(self, conn, qps, size, duration):
        size = size or self.size
        if duration is None:
            duration = self.duration

        labels = self.run_id
        labels += "_qps_" + str(qps)
        labels += "_c_" + str(conn)
        # TODO add mixer labels back
        # labels += "_"
        # labels += "mixer" if self.mixer else "nomixer"
        # labels += "_"
        # labels += "mixercache" if self.mixer_cache else "nomixercache"
        labels += "_" + str(self.size)

        if self.labels is not None:
            labels += "_" + self.labels

        grpc = ""
        if self.mode == "grpc":
            grpc = "-grpc -ping"

        fortio_cmd = (
            "fortio load -c {conn} -qps {qps} -t {duration}s -a -r {r} {grpc} -httpbufferkb=128 " +
            "-labels {labels}").format(
            conn=conn,
            qps=qps,
            duration=duration,
            r=self.r,
            grpc=grpc,
            labels=labels)

        if self.run_ingress:
            p = kubectl(self.client.name, self.ingress(fortio_cmd))
            if self.perf_record:
                perf(self.mesh,
                     self.server.name,
                     labels + "_srv_ingress",
                     duration=40)
            p.wait()

        if self.run_serversidecar:
            p = kubectl(self.client.name, self.serversidecar(fortio_cmd))
            if self.perf_record:
                perf(
                    self.mesh,
                    self.server.name,
                    labels + "_srv_serveronly",
                    duration=40)
            p.wait()

        if self.run_clientsidecar:
            p = kubectl(self.client.name, self.bothsidecar(fortio_cmd))
            if self.perf_record:
                perf(self.mesh,
                     self.server.name,
                     labels + "_srv_bothsidecars",
                     duration=40)
            p.wait()

        if self.run_baseline:
            p = kubectl(self.client.name, self.nosidecar(fortio_cmd))
            p.wait()


PERFCMD = "/usr/lib/linux-tools/4.4.0-131-generic/perf"
PERFSH = "get_perfdata.sh"
PERFWD = "/etc/istio/proxy/"


def perf(mesh, pod, labels, duration=20, runfn=run_command_sync):
    filename = labels + "_perf.data"
    filepath = PERFWD + filename
    perfpath = PERFWD + PERFSH

    # copy executable over
    kubecp(mesh, PERFSH, pod + ":" + perfpath)

    perf = kubectl(
        pod,
        "{perf_cmd} {filename} {duration}".format(
            perf_cmd=perfpath,
            filename=filename,
            duration=duration),
        runfn=runfn,
        container=mesh + "-proxy")

    print(perf)

    print(kubecp(mesh, pod + ":" + filepath + ".perf", filename + ".perf"))

    run_command_sync("./flame.sh " + filename + ".perf")
    return perf


def kubecp(mesh, from_file, to_file):
    namespace = os.environ.get("NAMESPACE", "twopods")
    cmd = "kubectl --namespace {namespace} cp {from_file} {to_file} -c" + mesh + \
        "-proxy".format(from_file=from_file,
                        to_file=to_file,
                        namespace=namespace)
    print(cmd)
    return run_command_sync(cmd)


def kubectl(pod, remote_cmd, runfn=run_command, container=None):
    namespace = os.environ.get("NAMESPACE", "twopods")
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
    fortio = Fortio(
        conn=args.conn,
        qps=args.qps,
        duration=args.duration,
        size=args.size,
        perf_record=args.perf,
        labels=args.labels,
        baseline=args.baseline,
        serversidecar=args.serversidecar,
        clientsidecar=args.clientsidecar,
        ingress=args.ingress,
        mode=args.mode,
        mesh=args.mesh)

    for conn in args.conn:
        for qps in args.qps:
            fortio.run(conn=conn, qps=qps, duration=args.duration, size=args.size)


def csv_to_int(s):
    return [int(i) for i in s.split(",")]


def getParser():
    parser = argparse.ArgumentParser("Run performance test")
    parser.add_argument(
        "conn",
        help="number of connections, comma separated list",
        type=csv_to_int,)
    parser.add_argument(
        "qps",
        help="qps, comma separated list",
        type=csv_to_int,)
    parser.add_argument(
        "duration",
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
        help="run traffic through ingress",
        default=None)
    parser.add_argument(
        "--labels",
        help="extra labels",
        default=None)
    parser.add_argument(
        "--mode",
        help="http or grpc",
        default="http")

    define_bool(parser, "baseline", "run baseline for all", False)
    define_bool(parser, "serversidecar", "run serversidecar-only for all", False)
    define_bool(parser, "clientsidecar", "run clientsidecar and serversidecar for all", True)

    return parser


def define_bool(parser, opt, help_arg, default_val):
    parser.add_argument(
        "--" + opt, help=help_arg, dest=opt, action='store_true')
    parser.add_argument(
        "--no-" + opt, help="do not " + help_arg, dest=opt, action='store_false')
    val = {opt: default_val}
    parser.set_defaults(**val)


def main(argv):
    args = getParser().parse_args(argv)
    print(args)
    return run(args)


if __name__ == "__main__":
    import sys
    sys.exit(main(sys.argv[1:]))
