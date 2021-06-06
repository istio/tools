#!/bin/bash

apt-get install -y make build-essential libssl-dev zlib1g-dev \
        libbz2-dev libreadline-dev libsqlite3-dev wget curl llvm \
        libncurses5-dev libncursesw5-dev xz-utils tk-dev
wget https://www.python.org/ftp/python/3.8.10/Python-3.8.10.tgz
tar zxvf Python-3.8.10.tgz
cd Python-3.8.10 && ./configure --with-ensurepip=install && make -j 16 && make install && ln -s  /usr/local/bin/python3 /usr/bin/python
cd .. && rm -rf ./Python-3.8.10 && rm -rf ./Python-3.8.10.tgz
