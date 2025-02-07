go-12f-common
=============

[![Actions Status](https://github.com/tombenke/go-12f-common/workflows/ci/badge.svg)](https://github.com/tombenke/go-12f-common)

## About

The common packages of a [12-factor application](https://12factor.net/) written in Golang.

This repository holds those infrastructure-level modules,
that every application requires that follows the core [12-factor principles](https://12factor.net/).

This package can be used to create 12-factor applications that have built-in logging, configurability,
 and lifecycle management for their internal components including healthcheck and graceful shutdown.

Figure 1. shows the structure of a typical application:

![The ApplicationRunner instance](docs/ApplicationRunner.png)

Every application is made of an Application object, which may hold one or more components.

Both the application and its internal components can be configured a common way,
via CLI parameters, environment variables and config files.
The implementation of configurability is based on [Cobra](https://cobra.dev/) and [Viper](https://github.com/spf13/viper).

The application object and its components may have configuration objects. The configuration objects assigned to the corresponding component typically hold parameters, that directly belongs to that component. These component-level configuration objects can be integrated into the central configuration object of the application, that may have additional, application-level configuration parameters.

Every configuration object must implement the `apprun.Configurer` interface:

- `GetConfigFlagSet()`: is a factory function that receives a reference to the main `pflag.FlagSet` config aggregate object, to that it puts its own configuration parameters.

- `LoadConfig()`: resolves the actual values of the configuration object. It takes into account the parameter definitions, the CLI and environment variables and the default values as well.

Every application has a lifecycle. The Figure 2. shows the states of the application that goes through during its lifecycle:

![The application states](docs/application-states.png)

The application object must implement the `apprun.Application` interface in order to have its lifecycle managed.
Moreover the components inside the application must implement the `apprun.ComponentLifecycleManager` interface.

The `apprun.ComponentLifecycleManager` interface defines the following functions:

- `Startup()`: The component should initialize itself and start a loop in a goroutine if necessary. When the loop is started, the component should signal that it's ready and healthy.
- `Shutdown()`: Shuts down the component. If a loop has been started it should be graceful shut down and the component should signal that it's not ready anymore.
- `Check()`:  It is called by the healthcheck API. If this function returns no error, then the component is considered healthy.

There are additional hooks an application can subscribe to:

- `AfterStartup`: Called after the components are initialized and became healthy. Parts of the application that depends on the components should be initialized here.
- `BeforeShutdown`: Called before the components are being shut down.
- `Check`: The application can also signal that it's not healthy. This is completely optional, the application is considered healthy when all of it's components are healthy by default


The `apprun.MakeAndRun()` wrapper function manages the configuration and lifecycle of a complete application.
It only needs the application level configuration aggregate, and the constructor of the application object and the main package of the application can be as simple as this:

```go
package main

import (
	"github.com/tombenke/go-12f-common/apprun"
	"github.com/tombenke/go-12f-common/must"
)

func main() {

	// Make and run an application via ApplicationRunner
	must.Must(apprun.MakeAndRun(&Config{}, NewApplication))
}
```

The `apprun.MakeAndRun()` does the following:

1. Creates a new application-level configuration aggregate object, that holds those parameters, that every application must have (e.g. health-check port, log-level, etc.).
2. Resolves the configuration parameters to the application and its components.
3. Calls the constructor function of the application with the complete, resolved configuration aggregate object.
4. Set the log level and log format of the logger module,
5. Starts the service endpoints for liveness and health-check (live: `true`, ready: `false`).
6. Enters the STARTUP state: calls the `Startup()` method of the application's components.
7. Waits until all components become healthy or times out.
8. If provided, the application's `AfterStartup()` hook is called.
9. When the application enters the RUN state, it registers the signal handler function for graceful-shutdown, then it keeps running its state until a kill or shutdown signal is not arrived.
10. When the application got a signal to shut down, it disables the readiness check, and enters the SHUTDOWN state.
11. If provided, the application's `BeforeShutdown()` hook is called.
12. Calls `Shutdown()` on the components.
13. When all internal components has been successfully stopped, the application terminates.

The application-level configuration parameters of the logger:

The log level:
- cli parameter: `--log-level`.
- env. variable: `LOG_LEVEL`.
- type: String. One of `panic, fatal, error, warning, info, debug, trace`.
- default value: `info`.

Log format:
- cli parameter: `--log-format`.
- env. variable: `LOG_FORMAT`.
- type: String. One of `json, text`.
- default value: `json`.

The application-level configuration parameters of the health-check endpoints:

Health-Check Port:
- cli parameter: `--health-check-port`
- env. variable: `--HEALTH_CHECK_PORT`
- default: `8080`.
	
Liveness-Check Path:
- cli parameter: `--liveness-check-path`.
- env. variable: `LIVENESS_CHECK_PATH`.
- default: `"/live"`.

Readiness-Check Path
- cli parameter: `--readiness-check-path`.
- env. variable: `READINESS_CHECK_PATH`.
- default: `"/ready"`.

The system components may fork their own service processes as a goroutine, that run either until they decide to stop, or the application needs to shut down. So that The application has a central `sync.WaitGroup` to that the components' `Startup()` functions got a reference as a parameter. Every system that forks its own subprocess must `Add()` itself to this waitgroup, and make sure it will call the `Done()` on this central waitgroup when this subprocess terminates, so that the application can wait for all the running internal processes to join.

When the application shuts down, it will call the `Shutdown()` method of each system component. 
It is the system components' decision and responsibility how many subprocesses it forks, and how it will terminate its forked subprocesses but it must be able to do at at least, when its `Shutdown()` method is called.

A typical pattern to implement this, to have a local channel inside the system component, that it shares with the subprocesses that is forks, then the subprocesses will do their job, until this channel is not closed. The only think the system component has to do in its `Shutdown()` method is, to close this channel.

See the [examples/scheduler](examples/scheduler/) as a sample for more details in this topic.

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

There are examples about the usage of the package in the [examples/](examples/) directory:

- [examples/simple](examples/simple/): Is a bare-minimum 12-factor application, that is build on top of the go-12f-common package.
- [examples/scheduler](examples/scheduler/): Demonstrates how to implement concurrent, communication processes as system components.

Build the binaries of the examples:

```bash
    task build
```

Then run it:

```bash
    examples/simple/main
```

or

```bash
    examples/scheduler/main --time-step 5s -l debug
```

## References

- [12-factor principles](https://12factor.net/)
- [Health Check Response Format for HTTP APIs](https://datatracker.ietf.org/doc/html/draft-inadarei-api-health-check-06)
- [Cobra](https://cobra.dev/)
- [Viper](https://github.com/spf13/viper)
