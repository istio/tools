import argparse
import datetime
import json
import os
import time

import requests
import requests.auth
import schedule

"""Define dashboard and snapshot related api wrapper for grafana
   Use schedule to run export snapshot jobs periodically
"""


class GrafanaAPI(object):

    def __init__(self):
        self.grafana_api_token = os.environ.get("grafana_api_token")
        if self.grafana_api_token is None:
            print("please set grafana_api_token env variable")
            exit(1)

        # init value of the host
        self.grafana_host = {
            "host": os.environ.get("grafana_host", "localhost"),
            "port": os.environ.get("grafana_host", 3000)
        }

        # init request session
        self.session = requests.Session()
        self.session.headers.update({
            "Authorization": "Bearer {}".format(self.grafana_api_token)
        })

    def get_dashboard_by_uid(self, uid):
        """
        :param uid: id of a dashboard
        :return: dashboard data
        """
        path = "api/dashboards/uid/{0}".format(uid)
        url = self.construct_url(path)

        print("Get dashboard by uid {0}".format(uid))

        resp = self.session.get(url)
        if not resp.ok:
            raise Exception(resp.content)

        return resp.json()

    def search_dashboard(self, dashboard_name):
        """
        :param dashboard_name: can be any keyword, use name to search here
        :return: dashboard uid found
        """
        path = "api/search/?query={0}".format(dashboard_name)
        url = self.construct_url(path)

        print("Search dashboard with name {0}".format(dashboard_name))

        resp = self.session.get(url)

        if not resp.ok:
            raise Exception(resp.content)
        if len(resp.json()) == 0:
            raise Exception("cannot find dashboard with name")
        return resp.json()[0].get("uid")

    def post_snapshot(self, dashboard, name="Istio Dashboard",
                      expires=None, key=None, deletekey=None, external=False):
        """
        Export dashboard to snapshot
        Snapshot name is formatted with dashboard name and timestamp
        """
        path = "api/snapshots"
        if external is True:
            if deletekey is None or key is None:
                print("Please specify key and delete key for external snapshot")
                exit(1)

        data = dashboard
        data['external'] = external
        current_dt = datetime.datetime.now().strftime('%Y-%m-%d %H:%M')
        data['name'] = name + "_" + str(current_dt)
        if expires is not None:
            data['expires'] = expires
        if key is not None:
            data['key'] = key
        if deletekey is not None:
            data['deletekey'] = deletekey

        print("Export dashboard name: {0} as snapshot, key:{1}"
              .format(data['name'], key))

        resp = self.session.post(self.construct_url(path), json=data)

        if not resp.ok:
            raise Exception(resp.content)
        return resp.json()

    def get_snapshots(self):
        """
        :return: all the snapshots exported
        """
        path = "api/dashboard/snapshots"
        url = self.construct_url(path)
        print(url)

        resp = self.session.get(url)
        if not resp.ok:
            raise Exception(resp.content)
        print(json.dumps(resp.json(), indent=4, sort_keys=True))

    def construct_url(self, path, host_url=None, port=None, protocol="http"):
        """
        To construct url for request
        """
        if host_url is None:
            host_url = self.grafana_host['host']
        if port is None:
            port = self.grafana_host['port']
        params = {
            'protocol': protocol,
            'host': host_url,
            'path': path,
        }
        if port is None:
            url_pattern = '{protocol}://{host}/{path}'
        else:
            params['port'] = port
            url_pattern = '{protocol}://{host}:{port}/{path}'

        return url_pattern.format(**params)

    def export_snapshot(self, dashboard_name, external):
        """
        :param dashboard_name: dashboard to export snapshot
        :param external: export locally or to raintank
        """
        dashboard_id = self.search_dashboard(dashboard_name)
        dashboard =\
            self.get_dashboard_by_uid(dashboard_id)
        self.post_snapshot(dashboard,
                           name=dashboard_name,
                           key=dashboard_id,
                           deletekey=dashboard_id,
                           external=external)


def getParser():
    parser = argparse.ArgumentParser("Define dashboard name"
                                     " and frequency to export snapshot")
    parser.add_argument(
        "--dashboard_name", help="Name of the dashboard to take snapshot",
        default="istio workload")

    parser.add_argument("--period", help="Period in minutes"
                        " to automatically export snapshots", default=5)

    parser.add_argument("--external", help="False to export snapshots locally"
                        "True to raintank", default=False)

    return parser


def main(argv):
    args = getParser().parse_args(argv)
    grafana_api = GrafanaAPI()

    # Run at schedule
    schedule.every(args.period).minutes.do(grafana_api.export_snapshot,
                                           args.dashboard_name,
                                           args.external)
    while True:
        schedule.run_pending()
        time.sleep(1)


if __name__ == '__main__':
    import sys
    sys.exit(main(sys.argv[1:]))
