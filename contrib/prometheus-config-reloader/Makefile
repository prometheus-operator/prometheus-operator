all: build

FLAGS =
ENVVAR = GOOS=linux GOARCH=amd64 CGO_ENABLED=0
NAME = prometheus-config-reloader
REPO = quay.io/coreos/$(NAME)
TAG = v0.0.4
IMAGE = $(REPO):$(TAG)

build:
	$(ENVVAR) go build -o $(NAME) main.go

container: build
	docker build -t $(IMAGE) .

push: container
	docker push $(IMAGE)

clean:
	rm -f $(NAME)

.PHONY: all build container push clean
