# Example 06

This example shows how to use the testcontainers package interacting directly with a custom API and a Postgres container interaction with each other using a Network.

## How to run

To run this example, you need to have Go installed on your system.

Once you have Go installed, you can run the example by following these steps:

1. Clone the repository:

```bash
git clone https://github.com/jfelipearaujo/testcontainers.git
```

2. Run the example with Makefile:

```bash
make run-tests-06
```

or via the command line:

```bash
go test -race -count=1 ./examples/example_06/... -timeout=300s -test.v -test.run ^TestFeatures$
```