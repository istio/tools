# This is the only version that is available
# hadolint ignore=DL3007
FROM networkstatic/netserver:latest
RUN apt-get update -y && apt-get upgrade -y       \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        netcat    \
        net-tools \
        tcpdump   \
        netperf   \
        less      \
        ncat      \
        termshark \
        python3   \
        curl \
        && rm -rf /var/lib/apt/lists/*

RUN curl -OL https://github.com/fortio/fortio/releases/download/v1.17.0/fortio_1.17.0_amd64.deb \
    && dpkg -i fortio_1.17.0_amd64.deb

# tcp echo port
EXPOSE 6789
# data port for netperf
EXPOSE 35000
# config port netperf 
EXPOSE 12865
# python http server
EXPOSE 8000
# fortio server
EXPOSE 8080
EXPOSE 8078
EXPOSE 8079

WORKDIR "/ambient-performance"

COPY . .
ENTRYPOINT ["./server-and-sleep.sh"]
