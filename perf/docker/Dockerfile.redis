FROM python:alpine

RUN pip3 install -q redis prometheus_client

ENV PORT 6379
ENV ADDRESS redis-master

ENV SLAVE_PORT 6379
ENV SLAVE_ADDRESS redis-slave

ADD redis/client.py /client.py
ADD prom_client.py /prom_client.py

CMD ["python3", "-u", "/client.py"]
