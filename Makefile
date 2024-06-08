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