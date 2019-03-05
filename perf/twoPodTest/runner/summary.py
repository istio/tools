import argparse
import json
import csv

def run(args):
    return read_file(args.file)

def json_readr(file):
    for line in open(file, mode="r"):
        yield json.loads(line)


def write_csv(header, data):
    with open('/dev/stdout', 'w') as out:
        w = csv.writer(out)
        w.writerow(header)
        w.writerows(data)

def read_file(file):
    colnames = ['ActualQPS', 'ActualDuration', 'p99']
    res = []
    for stats in json_readr(file):
        data = [stats[c] for c in colnames]
        res.append(data)
    res.sort()
    header = ['QPS', 'Connections', 'p99']
    write_csv(header, res)
    return 0


def get_parser():
    parser = argparse.ArgumentParser("Run performance test")
    parser.add_argument(
        "file", help="fortio_output_file", type=str)
    return parser


def main(argv):
    args = get_parser().parse_args(argv)
    return run(args)


if __name__ == "__main__":
    import sys

    sys.exit(main(sys.argv[1:]))
