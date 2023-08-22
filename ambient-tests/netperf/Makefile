CR=stjinxuan.azurecr.io
NAME=netperf
CR_NAME=$(CR)/$(NAME)
CONTROL_PORT=12865
DATA_PORTS=35000-35004

.PHONY: build push run push-cr push-local

default: build

build:
	docker build . -t $(CR_NAME) -t $(NAME)

push-cr:
	docker push $(CR_NAME)

push-local:
	kind load docker-image $(CR_NAME)

push: push-cr

run:
	docker run --rm --name $(NAME) $(NAME) -P
