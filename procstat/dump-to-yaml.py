#!/usr/bin/env python3

import argparse
from modules.collector import Collector
from modules.yaml_formatter import to_yaml
import sys
import time
import os

def main():
    parser = argparse.ArgumentParser(description='Transforms dumps from the sampler to yaml')
    parser.add_argument("--dump-path", type=str, help='Path where the target dump resides.')
    args = parser.parse_args()
    with open(args.dump_path, "rb") as file:
        collector = Collector(file=None)
        yaml = to_yaml(list(collector.read_dump(file)))
        print(yaml)


if __name__ == "__main__":
    main()
