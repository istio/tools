FROM docker:stable as docker

FROM gcr.io/istio-testing/build-tools:master-2019-12-15T16-17-48 as build-tools

FROM ubuntu:xenial
ARG user
ARG group
ARG uid=1000
ARG gid=1000

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

COPY --from=build-tools /usr/bin/hadolint /usr/bin/
COPY --from=build-tools /usr/bin/helm /usr/bin/
COPY --from=build-tools /usr/bin/hugo /usr/bin/
COPY --from=build-tools /usr/bin/kind /usr/bin/
COPY --from=build-tools /usr/bin/protoc /usr/bin/
COPY --from=build-tools /usr/bin/shellcheck /usr/bin/
COPY --from=build-tools /usr/bin/yq /usr/bin/
COPY --from=build-tools /go/ /usr/local/go/
COPY --from=build-tools /usr/local/ /usr/local/
COPY --from=build-tools /node_modules/ /node_modules/

RUN chmod 755 /usr/local/go/pkg /usr/local/go/src

# Install development packages.
RUN apt-get update && apt-get -qqy install --no-install-recommends \
autoconf=2.69-9 \
autotools-dev=20150820.1 \
bash-completion=1:2.1-4.2ubuntu1.1 \
build-essential=12.1ubuntu2 \
ca-certificates=20170717~16.04.2 \
curl=7.47.0-1ubuntu2.14 \
git=1:2.7.4-0ubuntu1.7 \
iptables=1.6.0-2ubuntu3 \
jq=1.5+dfsg-1ubuntu0.1 \
libtool=2.4.6-0.1 \
lsb-release=9.20160110ubuntu0.2 \
make=4.1-6 \
python3=3.5.1-3 \
sudo=1.8.16-0ubuntu1.8 \
tmux=2.1-3build1 \
unzip=6.0-20ubuntu1 \
vim=2:7.4.1689-3ubuntu1.3 \
wget=1.17.1-1ubuntu1.5 \
xz-utils=5.1.1alpha+20120614-2ubuntu2 \
&& rm -rf /var/lib/apt/lists/*

# Create user and allow sudo without password.
RUN addgroup --quiet --gid $gid $group \
&& adduser --quiet --disabled-password --gecos ",,,," --uid $uid --ingroup $group $user \
&& echo "${user} ALL=(ALL:ALL) NOPASSWD: ALL" > /etc/sudoers.d/$user

# Install Docker CLI.
COPY --from=docker /usr/local/bin/docker /usr/local/bin/docker

# Install gcloud and kubectl.
RUN echo "deb http://packages.cloud.google.com/apt cloud-sdk-$(lsb_release -c -s) main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list \
&& curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - \
&& apt-get update && apt-get -qqy install --no-install-recommends \
google-cloud-sdk=272.0.0-0 \
kubectl=1.17.0-00 \
&& rm -rf /var/lib/apt/lists/*

# Install bash completion files.
RUN /usr/bin/kind completion bash > /etc/bash_completion.d/kind \
&& /usr/bin/helm completion bash > /etc/bash_completion.d/helm \
&& /usr/bin/kubectl completion bash > /etc/bash_completion.d/kubectl \
&& curl -s -Lo - https://raw.githubusercontent.com/docker/cli/master/contrib/completion/bash/docker > /etc/bash_completion.d/docker

USER $user

# Fix the Docker socket access rights at login time to allow non-root access.
RUN echo "sudo chmod o+rw /var/run/docker.sock" >> /home/${user}/.bashrc

# Setup Go for the user.
RUN echo "# Go environment." >> /home/${user}/.bashrc \
&& echo "export GOROOT=/usr/local/go" >> /home/${user}/.bashrc \
&& echo "export GOPATH=~/go" >> /home/${user}/.bashrc \
&& echo "export PATH=\$GOROOT/bin:\$GOPATH/out/linux_amd64/release:\$GOPATH/bin:\$PATH" >> /home/${user}/.bashrc \
&& echo "export GO111MODULE=on" >> /home/${user}/.bashrc \
&& mkdir -p /home/${user}/go

WORKDIR /home/$user/go
ENTRYPOINT ["/bin/bash", "-c"]
