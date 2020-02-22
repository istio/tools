#!/usr/bin/env python3

import sys
from modules.prometheus_http import Prom

if __name__ == '__main__':
    p = Prom()
    p.run(sys.argv[1:])