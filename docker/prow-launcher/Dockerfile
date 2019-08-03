FROM ubuntu:xenial as build_context

ENV DOCKER_VERSION=18.06.1
ENV VERSION_SUFFIX="~ce~3-0~ubuntu"

ADD install-docker.sh /tmp/

# Installing
RUN apt-get update
RUN apt-get -qqy install --no-install-recommends \
  apt-transport-https \
  ca-certificates \
  curl \
  lsb-release \
  software-properties-common \
  time \
  make

RUN /tmp/install-docker.sh

# Add entrypoint to start docker
ADD prow-runner.sh /usr/local/bin/entrypoint
RUN chmod +rx /usr/local/bin/entrypoint

FROM scratch
COPY --from=build_context / /

# Docker in Docker settings
VOLUME /var/lib/docker
EXPOSE 2375

ENV PATH /usr/local/go/bin:/opt/go/bin:/usr/lib/google-cloud-sdk/bin:/root/.local/bin:${PATH}

# Set CI variable which can be checked by test scripts to verify
# if running in the continuous integration environment.
ENV CI prow

ENTRYPOINT ["entrypoint"]
