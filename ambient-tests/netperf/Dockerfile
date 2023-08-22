FROM networkstatic/netserver:f42eb5f94f31
RUN apt-get update -y && apt-get upgrade -y       \
    && apt-get install -y --no-install-recommends \
        netcat    \
        net-tools \
        tcpdump   \
        netperf   \
        less      \
        ncat      \
        termshark \
        python3   \
        && rm -rf /var/lib/apt/lists/*

# tcp echo port
EXPOSE 6789
# tcp ping port (python)
EXPOSE 9999
# data port for netperf
EXPOSE 35000
# config port netperf 
EXPOSE 12865
WORKDIR "/ambient-performance"

COPY . .
ENTRYPOINT ["./server-and-sleep.sh"]
