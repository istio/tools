#!/usr/bin/env python3

import unittest
from modules.yaml_formatter import to_yaml
from modules.collector import Collector
from time import sleep
from modules.sampler import Sampler
from multiprocessing import Process
import psutil


class TestSampler(unittest.TestCase):
    def foo_process_run(self):
        sleep(5)

    def test_yaml_formatting(self):
        FILENAME="/tmp/foobar"
        p = Process(target=self.foo_process_run)
        p.start()
        sampler = Sampler([psutil.Process(p.pid).name()])
        with open(FILENAME, "wb") as file:
            collector = Collector(file, sample_interval=0.5, sampler=sampler)
            collector.start()
            sleep(2)
            collector.stop()
            p.kill()
        with open(FILENAME, "rb") as file:
            yaml = to_yaml(list(collector.read_dump(file)))
        #print(yaml)
        self.assertIn("timestamp:", yaml)
        self.assertIn("cpu_times:", yaml)
        self.assertIn("cpu_percent:", yaml)
        self.assertIn("    pid: %s" % p.pid, yaml)

if __name__ == '__main__':
    unittest.main()
