GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
GIT_BRANCH_CLEAN := $(shell echo $(GIT_BRANCH) | sed -e "s/[^[:alnum:]]/-/g")
DOCKER_IMAGE := cool-kubernetes$(if $(GIT_BRANCH_CLEAN),:$(GIT_BRANCH_CLEAN))

DOCKERFILE ?= Dockerfile
DOCKER_BUILD_ARGS ?= --force-rm -f "$(DOCKERFILE)"

DOCKER_BUILD := docker build $(DOCKER_BUILD_ARGS) -t "$(DOCKER_IMAGE)" . 
DOCKER_RUN := docker run --rm -it 

.PHONY: all
all: push

.PHONY: build
build:
	$(DOCKER_BUILD)

.PHONY: push
push: build
	docker tag $(DOCKER_IMAGE) alejandrox1/cool-kubernetes:v1
	docker push alejandrox1/cool-kubernetes:v1
