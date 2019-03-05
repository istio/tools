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
    res = []
    for stats in json_readr(file):
        data = [
            stats['ActualQPS'],
            stats['NumThreads'],
            stats['Labels'].split('_')[-1],
            stats['p99']/1000,
            stats['p50']/1000,
        ]
        res.append(data)
    res.sort()
    header = ['QPS', 'Connections', 'Test', 'p99 (ms)', 'p50 (ms)']
    write_csv(header, res)


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
