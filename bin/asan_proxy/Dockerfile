ARG BASE_IMAGE
ARG proxy_version

FROM ${BASE_IMAGE}
ENV ISTIO_META_ISTIO_PROXY_SHA $proxy_version

# llvm - for symbolizer
# libc++-dev - asan builds use dynamically linked libc++
RUN apt-get update && apt-get install -y llvm libc++-dev 

ENV ASAN_SYMBOLIZER_PATH /usr/bin/llvm-symbolizer
ENV ASAN_OPTIONS halt_on_error=0

# proxy build packages envoy as a tgz with /usr/local/bin/envoy
# Automatically unzip it
ADD envoy.tgz /
