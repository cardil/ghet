# Those are callable targets
TASKS = $(shell go run ./build/ --list)

.PHONY: $(TASKS)
$(TASKS):
	@go run ./build/ $(ARGS) $@

.PHONY: help
help:
	@go run ./build/ --help
