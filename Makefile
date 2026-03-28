.PHONY: all build build-leanschool build-receipt-reader build-timetable-service tidy generate publish-model clean release

LEANSCHOOL_MODEL_VERSION ?= v0.0.1

all: build

build: build-leanschool build-receipt-reader build-timetable-service

build-leanschool:
	cd leanschool && go build -o ../bin/leanschool ./cmd/...

build-receipt-reader:
	cd receipt-reader && go build -o ../bin/receipt-reader ./cmd/...

build-timetable-service:
	cd timetable-service && go build -o ../bin/timetable-service ./cmd/...

tidy:
	cd leanschool-model && go mod tidy
	cd leanschool && go mod tidy
	cd receipt-reader && go mod tidy
	cd timetable-service && go mod tidy

generate:
	cd leanschool-model && go generate ./...
	cd leanschool && go generate ./...
	cd leanschool-ui && npm run generate:api


publish-model:
	cd leanschool-model && GONOSUMCHECK=* GOFLAGS=-mod=mod go mod tidy
	git tag $(LEANSCHOOL_MODEL_VERSION) leanschool-model
	git push origin $(LEANSCHOOL_MODEL_VERSION)

clean:
	rm -rf bin/

release:
	@echo "Enter version number (e.g., v1.0.0):"
	@read -p "Version: " VERSION && \
	git tag $$VERSION && \
	git push origin $$VERSION && \
	echo "Released version $$VERSION"
