#!/usr/bin/env python3

import argparse
from modules.collector import Collector
from modules.sampler import Sampler
import sys
import time
import os
import signal
import tempfile

global COLLECTOR
global STOPPED


def signal_handler(a, b):
    print("stopping... ")
    global COLLECTOR, STOPPED
    COLLECTOR.stop()
    STOPPED = True


def main():
    parser = argparse.ArgumentParser(description='Proc stat sampler CLI')
    parser.add_argument("--sample-frequency", type=int, default=1,
                        help='Number of samples to obtain per second. Defaults to 1 per second.')
    parser.add_argument("--track-proc-name", type=str, nargs="*",
                        help='Optional process name(s) to track.', default=[])
    parser.add_argument("--dump-path", type=str, default=tempfile.mktemp(prefix="psd-", suffix=".pickle-dump"),
                        help='Path where the result will be written.')
    args = parser.parse_args()
    print("proc stat sampler")
    print("--sample-frequency: ", args.sample_frequency)
    print("--track-proc-name: ", args.track_proc_name)
    print("--dump-path", args.dump_path)
    print("SIGINT/SIGTERM (ctrl+c) to stop")

    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    sampler = Sampler(process_names_of_interest=args.track_proc_name)
    global STOPPED, COLLECTOR
    STOPPED = False
    COLLECTOR = Collector(open(args.dump_path, "wb"), sampler=sampler)
    COLLECTOR.start()
    while not STOPPED:
        time.sleep(1)
    print("Stopped. Dump written to path '%s'" % args.dump_path)


if __name__ == "__main__":
    main()
