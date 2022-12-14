############################################################
# Dockerfile to run a Django-based web application
# Based on an Ubuntu Image
############################################################

# Set the base image to use to Ubuntu
FROM ubuntu:jammy

# Pip version
ENV PIP_INSTALL_VERSION=22.3.1

RUN apt-get update \
  && apt-get install -y --no-install-recommends \
  python3-pip python3-dev \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

# Use WORKDIR to switch to a directory
WORKDIR /usr/local/bin

RUN ln -s /usr/bin/python3 python \
  && pip3 install --no-cache-dir --upgrade pip==${PIP_INSTALL_VERSION}

# Set the working directory to /perf_dashboard
WORKDIR /perf_dashboard

# The environment variable ensures that the python output is set straight
# to the terminal without buffering it first
ENV PYTHONUNBUFFERED 1

# Set environment variables
ARG BUCKET_NAME_ARG="istio-build"
ARG PROJECT_ID_ARG="istio-testing"

ENV PROJECT_ID=${PROJECT_ID_ARG}
ENV BUCKET_NAME=${BUCKET_NAME_ARG}
ENV CUR_RELEASE=release-1.17
ENV DOWNLOAD_DATASET_DAYS=60
ENV BUCKET_PREFIX="perf/"
ENV BUCKET_DELIMITER="/"

# Copy the current directory contents into the container at /perf_dashboard
COPY . /perf_dashboard/

# Install any needed packages specified in requirements.txt
RUN pip3 install --no-cache-dir --requirement requirements.txt

# uWSGI will listen on this port
EXPOSE 8000

COPY ./docker-entrypoint.sh /

RUN chmod u+x docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]
