
NAME = cf-conduit
GOBUILD = go build
ALL_ARCH = amd64 386
ALL_GOOS = windows linux darwin

.PHONY: install
install:
	mkdir -p bin
	$(GOBUILD) -o bin/$(NAME)
	cf install-plugin -f bin/$(NAME)

.PHONY: dist
dist:
	mkdir -p bin
	for arch in $(ALL_ARCH); do \
		for platform in $(ALL_GOOS); do \
			CGO_ENABLED=0 GOOS=$$platform ARCH=$$arch $(GOBUILD) -o bin/$(NAME).$$platform.$$arch; \
		done; \
	done

.PHONY: clean
clean:
	rm -rf bin

.PHONY: test
test:
	go vet
	go test $(go list ./... | grep -v /vendor/)
