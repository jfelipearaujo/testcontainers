TAG := $(shell git describe --tags --abbrev=0 2>/dev/null)
VERSION := $(shell echo $(TAG) | sed 's/v//')

tag:
	@if [ -z "$(TAG)" ]; then \
        echo "No previous version found. Creating v1.0.0 tag..."; \
        git tag v1.0.0; \
    else \
        echo "Previous version found: $(VERSION)"; \
        read -p "Bump major version (M/m), minor version (R/r), or patch version (P/p)? " choice; \
        if [ "$$choice" = "M" ] || [ "$$choice" = "m" ]; then \
            echo "Bumping major version..."; \
			major=$$(echo $(VERSION) | cut -d'.' -f1); \
            major=$$(expr $$major + 1); \
            new_version=$$major.0.0; \
		elif [ "$$choice" = "R" ] || [ "$$choice" = "r" ]; then \
            echo "Bumping minor version..."; \
			minor=$$(echo $(VERSION) | cut -d'.' -f2); \
            minor=$$(expr $$minor + 1); \
            new_version=$$(echo $(VERSION) | cut -d'.' -f1).$$minor.0; \
		elif [ "$$choice" = "P" ] || [ "$$choice" = "p" ]; then \
            echo "Bumping patch version..."; \
			patch=$$(echo $(VERSION) | cut -d'.' -f3); \
            patch=$$(expr $$patch + 1); \
            new_version=$$(echo $(VERSION) | cut -d'.' -f1).$$(echo $(VERSION) | cut -d'.' -f2).$$patch; \
        else \
            echo "Invalid choice. Aborting."; \
            exit 1; \
        fi; \
        echo "Creating tag for version v$$new_version..."; \
        git tag v$$new_version; \
    fi

gen-docs:
	@echo "Generating docs..."
	gomarkdoc -o README.md ./pkg/...

run-tests:
	@echo "Running tests..."
	make run-tests-01
	make run-tests-02
	make run-tests-03
	make run-tests-04
	make run-tests-05	
	make run-tests-06

run-tests-01:
	@echo "Running tests for example 01..."
	@go test -race -count=1 ./examples/example_01/... -timeout=300s -test.v -test.run ^TestFeatures$

run-tests-02:
	@echo "Running tests for example 02..."
	@go test -race -count=1 ./examples/example_02/... -timeout=300s -test.v -test.run ^TestFeatures$

run-tests-03:
	@echo "Running tests for example 03..."
	@go test -race -count=1 ./examples/example_03/... -timeout=300s -test.v -test.run ^TestFeatures$

run-tests-04:
	@echo "Running tests for example 04..."
	@go test -race -count=1 ./examples/example_04/... -timeout=300s -test.v -test.run ^TestFeatures$

run-tests-05:
	@echo "Running tests for example 05..."
	@go test -race -count=1 ./examples/example_05/... -timeout=300s -test.v -test.run ^TestFeatures$

run-tests-06:
	@echo "Running tests for example 06..."
	@go test -race -count=1 ./examples/example_06/... -timeout=300s -test.v -test.run ^TestFeatures$