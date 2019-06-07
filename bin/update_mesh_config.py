import yaml   # requires pyyaml
import argparse
import sys


def str_presenter(dumper, data):
    if len(data.splitlines()) > 1:  # check for multiline string
        return dumper.represent_scalar('tag:yaml.org,2002:str', data, style='|')
    return dumper.represent_scalar('tag:yaml.org,2002:str', data)


yaml.add_representer(str, str_presenter)


def remove_keys(mp, keys):
    for key in keys:
        mp.pop(key, None)


def sanitize(res):
    meta = res["metadata"]
    remove_keys(meta, ['resourceVersion', 'selfLink',
                       'uid', 'creationTimestamp'])

    if "annotations" in meta:
        anno = meta["annotations"]
        remove_keys(anno, ['kubectl.kubernetes.io/last-applied-configuration'])


def update_mesh(infile, update_func):
    jj = yaml.load(open(infile), Loader=yaml.FullLoader)

    sanitize(jj)
    mc = yaml.load(jj['data']["mesh"], Loader=yaml.FullLoader)

    update_func(mc)

    jj['data']["mesh"] = yaml.dump(mc)
    print(yaml.dump(jj))


def enable_mixer(infile, mixer_server):
    def update_func(mc):
        mc["mixerCheckServer"] = mixer_server
        mc["mixerReportServer"] = mixer_server

    update_mesh(infile, update_func)


def disable_mixer(infile):
    def update_func(mc):
        remove_keys(mc, ['mixerCheckServer', 'mixerReportServer'])

    update_mesh(infile, update_func)


def getParser():
    parser = argparse.ArgumentParser("enable or disable mixer in meshconfig")
    parser.add_argument(
        "func", help="enable_mixer|disable_mixer")
    parser.add_argument(
        "filename", help="filename")

    parser.add_argument("--mixer-server", help="mixer address",
                        default="istio-telemetry.istio-system.svc.cluster.local:9091")

    return parser


def main(argv):
    args = getParser().parse_args(argv)

    if args.func == "enable_mixer":
        return enable_mixer(args.filename, args.mixer_server)

    if args.func == "disable_mixer":
        return disable_mixer(args.filename)


if __name__ == '__main__':
    import sys
    sys.exit(main(sys.argv[1:]))
