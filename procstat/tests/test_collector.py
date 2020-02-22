#!/usr/bin/env python3

import unittest
from time import sleep
from modules.collector import Collector


class TestCollector(unittest.TestCase):
    def test_collector_dump(self):
        RUN_SECONDS = 2
        SAMPLE_FREQUENCY = 10
        FILENAME="/tmp/foobar"

        with open(FILENAME, "wb") as file:
            expected_snapshot_count = SAMPLE_FREQUENCY * RUN_SECONDS
            collector = Collector(file, sample_interval=1/SAMPLE_FREQUENCY)
            collector.start()
            sleep(RUN_SECONDS)
            collector.stop()

        with open(FILENAME, "rb") as file:
            snapshots = list(collector.read_dump(file))
        self.assertEqual(expected_snapshot_count, len(snapshots))
        last_snapshot = snapshots[expected_snapshot_count - 1]
        self.assertIn("cpu_percent", last_snapshot)
        self.assertIn("cpu_times", last_snapshot)

if __name__ == '__main__':
    unittest.main()
