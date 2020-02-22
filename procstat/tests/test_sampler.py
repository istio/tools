#!/usr/bin/env python3

import unittest
from modules.sampler import Sampler


class TestSampler(unittest.TestCase):
    def test_get_snapshot(self):
        sampler = Sampler()
        pd = sampler.get_snapshot()
        expected_keys = ["cpu_percent", "cpu_times"]
        for key in expected_keys:
            self.assertIn(key, pd)

if __name__ == '__main__':
    unittest.main()
