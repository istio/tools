#!/usr/bin/env python3

import psutil
import enum
import time

class Sampler:
    def __init__(self, process_names_of_interest=[], global_cpu_percent=True, global_cpu_times=True, per_cpu_percent=False, per_cpu_times=False):
        self.process_attrs_of_interest = [
            "pid", "name", "cpu_times", "cpu_percent"]
        self.global_cpu_percent = global_cpu_percent
        self.global_cpu_times = global_cpu_times
        self.per_cpu_percent = per_cpu_percent
        self.per_cpu_times = per_cpu_times
        self.processes_of_interest = self.get_processes_of_interest(
            process_names_of_interest)

    def get_processes_of_interest(self, process_names_of_interest):
        processes_of_interest = []
        for p in psutil.process_iter(attrs=self.process_attrs_of_interest):
            if p.info["name"] in process_names_of_interest:
                processes_of_interest.append(p)
        return processes_of_interest

    def get_snapshot(self):
        # This should be fast, as the measurement should least interfere with
        # what we're trying to measure. Consider adding a benchmark test for
        # this call to estimate the overhead we're adding.
        o = {}
        o["timestamp"] = time.time()
        if self.global_cpu_percent:
            o["cpu_percent"] = psutil.cpu_percent(
                interval=0, percpu=False)
        if self.global_cpu_times:
            o["cpu_times"] = psutil.cpu_times(percpu=False)
        if self.per_cpu_percent:
            o["per_cpu_percent"] = psutil.cpu_percent(
                interval=0, percpu=True)
        if self.per_cpu_times:
            o["per_cpu_times"] = psutil.cpu_times(percpu=True)
        o["cpu_stats"] = psutil.cpu_stats()
        o["processes"] = []
        for process in self.processes_of_interest:
            attrs = {}
            o["processes"].append(attrs)
            for attr in self.process_attrs_of_interest:
                attrs[attr] = process.info[attr]

        return o
