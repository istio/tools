# Docker image for config changer.
FROM gcr.io/google-containers/hyperkube:v1.12.1
RUN apt-get update \
  && apt-get install -y python3-pip=9.0.1-2+deb9u1 python3-dev=3.5.3-1 \
    --no-install-recommends \
  && ln -s /usr/bin/python3 /usr/local/bin/python \
  && pip3 install --upgrade pip==19.2.3 \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*
