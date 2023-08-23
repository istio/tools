CR=stjinxuan.azurecr.io
NAME=ambient-performance
CR_NAME=$(CR)/$(NAME)

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
