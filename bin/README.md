## export_grafana_snapshot

Description:

The script run at specified interval to export the selected snapshots url to google cloud datastore with extra attributes such as time, tags.
User can then search snapshots by different attributes on datastore console. The dashboard name and snapshot interval can be set through command line.

Due to limitation of grafana snapshot API, the implementation is done with selenium, so there is some uncertain latency.

Dependency:
1. pip3 install -r ./requirements.txt

Setup:

1. Verify grafana working, reference: https://istio.io/docs/tasks/telemetry/metrics/using-istio-dashboard/
1. Set grafana_api_tolen env variable. Navigate to local running grafana url to get the API key, if no one exists, generate a new one with admin permission
2. Set GOOGLE_APPLICATION_CREDENTIALS env variable. Follow the instruction here: https://cloud.google.com/docs/authentication/getting-started
3. Setup google cloud datastore, go to google cloud console and enable datastore in datastore mode.
4. If you do not have ChromeDriver in your PATH, download corresponding release to PATH from here: http://chromedriver.chromium.org/downloads

Run:

command line arguments:
1. --period: interval to export the snapshot url, default=1
2. --dashboard_name: name of dashboard to export, default="istio performance"

TODO:
1. Add multiple dashboards support

