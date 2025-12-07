MODULES := $(shell find ./ -name "go.mod" -exec dirname {} \;)
BUILD_DIR := build

.PHONY: build fmt test tidy

build:
	mkdir -p $(BUILD_DIR)
	for mod in $(MODULES); do \
		echo "Building $$mod"; \
		name=$$(basename $$mod); \
		(cd $$mod && go build -o $(shell pwd)/$(BUILD_DIR)/$$name .); \
	done

fmt:
	for mod in $(MODULES); do \
		echo "Formatting $$mod"; \
		(cd $$mod && go fmt ./...); \
	done

test:
	for mod in $(MODULES); do \
		echo "Testing $$mod"; \
		(cd $$mod && go test ./...); \
	done

tidy:
	for mod in $(MODULES); do \
		echo "Tidying $$mod"; \
		(cd $$mod && go mod tidy); \
	done

clean:
	rm build/*