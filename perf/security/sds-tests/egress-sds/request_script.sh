#!/bin/bash
# shellcheck disable=SC2086
sleep 30
num_curl=0
num_succeed=0
while true; do
  resp_code=$(curl -s  -o /dev/null -w "%{http_code}\n" URL_TO_REPLACE)
  if [ ${resp_code} = 200 ]; then
    num_succeed=$((num_succeed+1))
  else
    echo "$(date +"%Y-%m-%d %H:%M:%S:%3N") curl to URL_TO_REPLACE failed, response code $resp_code"
  fi
  num_curl=$((num_curl+1))
  echo "$(date +"%Y-%m-%d %H:%M:%S:%3N") Out of ${num_curl} curl, ${num_succeed} succeeded."
  sleep .5
done
