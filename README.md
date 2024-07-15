go-12f-common
=============

[![Actions Status](https://github.com/tombenke/go-12f-common/workflows/ci/badge.svg)](https://github.com/tombenke/go-12f-common)

## About

The common packages of a [12-factor application](https://12factor.net/) written in Golang.

This repository holds those infrastructure-level modules,
that every application requires that follows the core [12-factor principles](https://12factor.net/).

This package can be used to create 12-factor applications that have built-in logging, configurability,
graceful shutdown, healthcheck and lifecycle management for their internal components.

Both the application and its internal components can be configured a common way,
via CLI parameters, environment variables and config files.
The implementation of configurability is based on [Cobra](https://cobra.dev/) and [Viper](https://github.com/spf13/viper).

![The application states](docs/application-states.png)

![The ApplicationRunner instance](docs/ApplicationRunner.png)

## Development

Clone the repository, then install the dependencies and the development tools:

```bash
task install
```

List the tasks:

```bash
task list
```

## The examples

There are examples about the usage of the package in the [examples/](examples/) directory.

Build the binaries of the examples:

```bash
    task build
```

Then run it:

```bash
    cd examples/simple
    ./main
```

## References

- [Health Check Response Format for HTTP APIs](https://datatracker.ietf.org/doc/html/draft-inadarei-api-health-check-06)
