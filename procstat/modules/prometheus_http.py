#!/usr/bin/env python3

from prometheus_client import start_http_server, Histogram, Summary, Gauge
import random
import time
from os import pipe, fdopen
from signal import signal, SIGINT, SIGTERM
from argparse import ArgumentParser
from modules.sampler import Sampler
from modules.collector import Collector

global COLLECTOR


def signal_handler(a, b):
    '''
    We will gracefully quit upon observing SIGTERM/SIGINT.
    We do so by calling stop on the collector, which in turn will end up
    closing the write side of the pipe that it is writing to.
    This will be noticed by the code below, which processes the read side.
    '''
    print("stopping... ")
    global COLLECTOR
    COLLECTOR.stop()


class Prom:
    def __init__(self):
        pass

    def run(self, arguments):
        parser = ArgumentParser(description='Proc stat sampler CLI')
        parser.add_argument("--track-proc-name", type=str, nargs="*",
                            help='Optional process name(s) to track.', default=[])
        parser.add_argument("--sample-frequency", type=int, default=1,
                            help='Number of samples to obtain per second.')
        parser.add_argument("--http-port", type=int, default=8000,
                            help='Http port for exposing prometheus metrics.')

        args = parser.parse_args(arguments)

        signal(SIGINT, signal_handler)
        signal(SIGTERM, signal_handler)

        global COLLECTOR
        # We hand the write side of the pipe to our proc stat collector.
        pipe_read_fd, pipe_write_fd = pipe()
        COLLECTOR = Collector(fdopen(pipe_write_fd, "wb", 1024), sampler=Sampler(
            process_names_of_interest=args.track_proc_name), sample_interval=1.0/args.sample_frequency)
        # Start serving prometheus stats over http
        start_http_server(args.http_port)

        # Start sampling proc stat.
        COLLECTOR.start()
        
        cpu_times_guest = Gauge('cpu_times_guest', '')
        cpu_times_guest_nice = Gauge('cpu_times_guest_nice', '')
        cpu_times_idle = Gauge('cpu_times_idle', '')
        cpu_times_iowait = Gauge('cpu_times_iowait', '')
        cpu_times_irq = Gauge('cpu_times_irq', '')
        cpu_times_nice = Gauge('cpu_times_nice', '')
        cpu_times_softirq = Gauge('cpu_times_softirq', '')
        cpu_times_steal = Gauge('cpu_times_steal', '')
        cpu_times_system = Gauge('cpu_times_system', '')
        cpu_times_user = Gauge('cpu_times_user', '')

        cpu_stats_ctx_switches = Gauge('cpu_stats_ctx_switches', '')
        cpu_stats_interrupts = Gauge('cpu_stats_interrupts', '')
        cpu_stats_soft_interrupts = Gauge('cpu_stats_soft_interrupts', '')
        cpu_stats_syscalls = Gauge('cpu_stats_syscalls', '')

        # The collector will write proc stat samples to the file descriptor we handed it above.
        # We will read those here, and update the prometheus stats according to these samples.
        with fdopen(pipe_read_fd, "rb", 1024) as f:
            it = COLLECTOR.read_dump(f)
            # TODO(oschaaf): Add an option here to also stream the raw data to another fd,
            # as we loose information in the summary we serve over http. This could be helpfull
            # when in-depth analysis is desired of an observed problem.
            for entry in it:
                cpu_times_guest.set(entry["cpu_times"].guest)
                cpu_times_guest_nice.set(entry["cpu_times"].guest_nice)
                cpu_times_idle.set(entry["cpu_times"].idle)
                cpu_times_iowait.set(entry["cpu_times"].iowait)
                cpu_times_irq.set(entry["cpu_times"].irq)
                cpu_times_nice.set(entry["cpu_times"].nice)
                cpu_times_softirq.set(entry["cpu_times"].softirq)
                cpu_times_steal.set(entry["cpu_times"].steal)
                cpu_times_system.set(entry["cpu_times"].system)
                cpu_times_user.set(entry["cpu_times"].user)

                cpu_stats_ctx_switches.set(entry["cpu_stats"].ctx_switches)
                cpu_stats_interrupts.set(entry["cpu_stats"].interrupts)
                cpu_stats_soft_interrupts.set(entry["cpu_stats"].soft_interrupts)
                cpu_stats_syscalls.set(entry["cpu_stats"].syscalls)
        print("stopped")
