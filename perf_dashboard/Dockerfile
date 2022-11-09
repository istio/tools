############################################################
# Dockerfile to run a Django-based web application
# Based on an Ubuntu Image
############################################################

# Set the base image to use to Ubuntu
FROM ubuntu:jammy

RUN apt-get update \
  && apt-get install -y python3-pip python3-dev \
  && cd /usr/local/bin \
  && ln -s /usr/bin/python3 python \
  && pip3 install --upgrade pip

# The environment variable ensures that the python output is set straight
# to the terminal without buffering it first
ENV PYTHONUNBUFFERED 1

# Set environment variables
ENV CUR_RELEASE=release-1.15
ENV PROJECT_ID="service-mesh-296815"
ENV BUCKET_NAME=intel-istio-build
ENV BUCKET_PREFIX="perf/"
ENV BUCKET_DELIMITER="/"

# Create root directory for our project in the container
RUN mkdir /perf_dashboard

# Set the working directory to /perf_dashboard
WORKDIR /perf_dashboard

# Copy the current directory contents into the container at /perf_dashboard
COPY . /perf_dashboard/

# Copy and set application default credentials
COPY application_default_credentials.json /perf_dashboard
ENV GOOGLE_APPLICATION_CREDENTIALS=/perf_dashboard/application_default_credentials.json

# Install any needed packages specified in requirements.txt
RUN pip3 install -r requirements.txt

# uWSGI will listen on this port
EXPOSE 8000

COPY ./docker-entrypoint.sh /

RUN chmod u+x docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]
