test-example-01:
	go test -race -count=1 ./examples/example_01/... -test.v -test.run ^TestFeatures$

test-example-02:
	go test -race -count=1 ./examples/example_02/... -test.v -test.run ^TestFeatures$