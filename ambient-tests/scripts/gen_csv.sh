#! /bin/bash
# create csv from key-value apirs

set -eu
shopt -s extglob
source scripts/config.sh

# non csv files
for file in $RESULTS/{TCP_STREAM,TCP_CRR,TCP_RR}
do
    echo $file
    base=$(basename "$file")
    python ./scripts/results_to_csv.py \
        "$TEST_RUN_SEPARATOR"          \
        < "$file"                      \
        > "$RESULTS/$base.csv"
done
