import argparse
import datetime
import json
import os
import time

import requests
import requests.auth
import schedule
from datetime import datetime
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from google.cloud import datastore


"""Define dashboard related api wrapper for grafana.
   Use schedule to run export snapshot jobs periodically.
   Due to limitation of post snapshot api, use selenium to mock action
   of clicking share.
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

        return resp.json()['dashboard']

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
        return resp.json()[0]

    def get_snapshots(self):
        """
        :return: all the snapshots exported
        """
        path = "api/dashboard/snapshots"
        url = self.construct_url(path)

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


def persist(datastore_client, dashboard, snapshot_url, kind='Snapshot'):

    # The Cloud Datastore key
    snapshot_key = datastore_client.key(kind, dashboard['uid'])

    # Prepares the new entity
    # Can add more attributes
    snapshot = datastore.Entity(key=snapshot_key)
    snapshot['url'] = snapshot_url
    snapshot['dashboard_uid'] = dashboard['uid']
    snapshot['dashboard_name'] = dashboard['title']

    snapshot['tags'] = " ".join(filter(None, dashboard['tags']))

    current_dt = datetime.now().strftime('%Y-%m-%d %H:%M')
    snapshot['timestamp'] = str(current_dt)

    # Saves the entity
    datastore_client.put(snapshot)

    print('Saved snapshot url {} for dashboard {}'.
          format(snapshot['url'], dashboard['title']))


def export_snapshot(grafana_api, datastore_client, dashboard_name):
    """
    :param grafana_api: grafana api client
    :param datastore_client: datastore api client
    :param dashboard_name: dashboard to export snapshot
    """
    dashboard = grafana_api.search_dashboard(dashboard_name)
    # remove / at beginning
    path = dashboard['url'][1:]
    dashboard_url = grafana_api.construct_url(path)

    snapshot_url = click_share_link(dashboard_url)

    persist(datastore_client, dashboard, snapshot_url)


def click_share_link(dashboard_url):
    chrome_options = webdriver.ChromeOptions()
    chrome_options.add_argument("--headless")
    driver = webdriver.Chrome(chrome_options=chrome_options)
    print("dashboard url:{0}".format(dashboard_url))
    driver.get(dashboard_url)

    element = driver.find_element_by_class_name("navbar-button--share")
    element.click()

    snapshot_tap = WebDriverWait(driver, 1).until(
        EC.presence_of_element_located((By.LINK_TEXT, "Snapshot")))
    snapshot_tap.click()

    snapshot_share = WebDriverWait(driver, 1).until(
        EC.presence_of_element_located((By.CLASS_NAME, "btn-secondary")))
    snapshot_share.click()
    snapshot_link = WebDriverWait(driver, 10).until(
        EC.presence_of_element_located((By.CLASS_NAME, "share-modal-link")))
    print(snapshot_link.text)
    return snapshot_link.text


def getParser():
    parser = argparse.ArgumentParser("Define dashboard name"
                                     " and frequency to export snapshot")
    parser.add_argument(
        "--dashboard_name", help="Name of the dashboard to take snapshot",
        default="istio performance")

    parser.add_argument("--period", help="Period in minutes"
                        " to automatically export snapshots", default=1)

    return parser


def main(argv):
    args = getParser().parse_args(argv)
    grafana_api = GrafanaAPI()

    # init datastore client
    datastore_client = datastore.Client()

    # Run at schedule
    schedule.every(args.period).minutes.do(export_snapshot,
                                           grafana_api,
                                           datastore_client,
                                           args.dashboard_name)
    while True:
        schedule.run_pending()
        time.sleep(1)


if __name__ == '__main__':
    import sys
    sys.exit(main(sys.argv[1:]))
