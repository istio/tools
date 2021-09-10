# Istio documentation evaluation

The tooling in this directory is used to evaluate documentation on the Istio.io
website, looking for documentation without tests and evaluating it based on
Google analytics results. Pages are divided into different priorities based on
thresholds for P0, P1, and P2 and a CSV file is generated with the results. This
can then be imported into a Google docs spreadsheet in order to create the
testing day spreadsheet.

## Usage

First, export a CSV containing the Istio.io analytics. In this case, it's called
istio.csv.

```bash
go build .
./evaluate-docs --analyticspath istio.csv --docspath ~/code/istio/istio.io/istio.io/content/ --outpath out.csv
```

* analyticspath represents the path to the analytics CSV
* docspath represents the path to a cloned, up to date copy of the Istio.io docs repo
* outpath represents the file to store the evaluation results in


