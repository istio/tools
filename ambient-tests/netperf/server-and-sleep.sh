#! /bin/bash

set -eux
# Couldn't find a way to run netserver in the foreground.
netserver $@
echo "Started echo server."
ncat -e /bin/cat -k -l 6789 &
echo "Started netserver."
python ./tcp_ping/server.py &
echo "Started TCP ping server"
echo "Sleeping"
sleep 365d
