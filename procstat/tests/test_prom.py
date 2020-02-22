#!/usr/bin/env python3

import unittest
import os

class TestProm(unittest.TestCase):
    def test_prom_101(self):
        assert os.system("tests/test_prom.sh") == 0

if __name__ == '__main__':
    unittest.main()
