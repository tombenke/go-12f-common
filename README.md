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

The Application object must implement the `apprun.LifecycleManager` interface in order to have its lifecycle managed.
Moreover the components inside the application may also implement this interface.

The `apprun.LifecycleManager` interface defines the following functions:

- `Startup()`: Starts the application and its internal components, by calling their `Startup()` method.
- `Shutdown()`: Shuts down the internal components, by calling their `Shutdown()` method, then shuts down the application as well.
- `Check()`:  It is called by the healthcheck API. If this function returns with nil that means the appication or component is healthy. If it returns with any error, that means the application or component is either sick, or yet not ready for working.

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
5. Starts the servive endpoints for liveness and health-check (live: `true`, ready: `false`).
6. Enters the STARTUP state: calls the Startup() method of the application object. The Application's `Startup()` method is responsible to start the internal services of the application, by calling the `Startup()` method of the internal components.
7. When the `Startup()` method of the application finished the component services are starting, and the application runner starts checking their readiness by calling their `Check()` method. When each of the internal services is ready, then the application is ready, and enters the RUN state.
8. When the application enters the RUN state, it registers the signal handler function for graceful-shutdown, then it keeps running its state until a kill or shutdown signal is not arrived.
9. When the application got a signal to shut down, it disables the readiness check, and enters the SHUTDOWN state.
10. Calls the `Shutdown()` memthod of the application, that is responsible to call further the `Shutdown()` function of its internal components.
11. When all internal components has been successfully stopped, the application terminates.

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
