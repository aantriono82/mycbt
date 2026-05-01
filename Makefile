SHELL := /usr/bin/env bash

.PHONY: test-backend test-backend-integration

test-backend:
	./scripts/test-backend.sh

test-backend-integration:
	./scripts/test-backend.sh integration
