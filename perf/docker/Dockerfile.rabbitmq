FROM python:alpine

RUN pip3 install -q pika prometheus_client

ENV ADDRESS rabbitmq
ENV USERNAME istio

ADD rabbitmq/client.py /client.py
ADD prom_client.py /prom_client.py

CMD ["python3", "-u", "/client.py"]
