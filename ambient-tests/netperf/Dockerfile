FROM networkstatic/netserver
RUN apt update -y && apt upgrade -y
RUN apt install -y netcat 
RUN apt install -y net-tools
RUN apt install -y tcpdump
RUN apt install -y netperf
RUN apt install -y less
RUN apt install -y ncat
RUN apt install -y termshark
RUN apt install -y python3

# tcp echo port
EXPOSE 6789
# tcp ping port (python)
EXPOSE 9999
# data port for netperf
EXPOSE 35000
# config port netperf 
EXPOSE 12865

COPY . .
ENTRYPOINT ./server-and-sleep.sh
