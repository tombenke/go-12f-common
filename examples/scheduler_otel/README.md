# scheduler example

This application demonstrates how to create system components, that are running as concurrent processes, and communicate with each other.

The two components are:

- [timer](timer/): Emits an event every `time-step` period, that triggers the sending of the actual time value to the worker process.
- [worker](worker): Receives the time message, and prints it.

Get help on parameters:

```bash
$ examples/scheduler/main --help
Usage:
   [flags]

Flags:
      --health-check-port uint        The HTTP port of the healthcheck endpoints (default 8080)
  -h, --help                          help for this command
      --liveness-check-path string    The path of the liveness check endpoint (default "/live")
  -f, --log-format string             The log format: json | text (default "json")
  -l, --log-level string              The log level: panic | fatal | error | warning | info | debug | trace (default "info")
      --readiness-check-path string   The path of the readiness check endpoint (default "/ready")
      --time-step string              The size of a time-step (default "60s")
```

Run the application with debug log level:

```bash
$ examples/scheduler/main -l debug --time-step 1s
INFO[0000] Application.Config: {worker:{} timer:{TimeStep:1s}} 
{"level":"debug","msg":"ar.config: \u0026{LogLevel:debug LogFormat:json HealthCheckPort:8080 LivenessCheckPath:/live ReadinessCheckPath:/ready OtelConfig:{ServiceName:undefined TracesSampler:always_off TracesSamplerArg: TracesExporter:otlp MetricsExporter:otlp LogsExporter:otlp}}","time":"2024-07-18T20:53:38.613821926+02:00"}
{"level":"info","msg":"ApplicationRunner Run","time":"2024-07-18T20:53:38.61391491+02:00"}
{"level":"info","msg":"HealthCheck services Startup","time":"2024-07-18T20:53:38.613926468+02:00"}
{"level":"debug","msg":"HealthCheck add endpoint: /live, 0x847860","time":"2024-07-18T20:53:38.613938559+02:00"}
{"level":"debug","msg":"HealthCheck add endpoint: /ready, 0x847880","time":"2024-07-18T20:53:38.613956958+02:00"}
{"level":"info","msg":"Checking if HealthCheck server started...","time":"2024-07-18T20:53:38.624368657+02:00"}
{"level":"info","msg":"HealthCheck is up and running!","time":"2024-07-18T20:53:38.625959927+02:00"}
{"level":"info","msg":"Application Startup","time":"2024-07-18T20:53:38.625995373+02:00"}
{"level":"debug","msg":"Timer: Startup","time":"2024-07-18T20:53:38.626009255+02:00"}
{"level":"debug","msg":"Timer: config: \u0026{TimeStep:1s}","time":"2024-07-18T20:53:38.626055825+02:00"}
{"level":"debug","msg":"Worker: Startup","time":"2024-07-18T20:53:38.626125151+02:00"}
{"level":"debug","msg":"Worker: config: \u0026{}","time":"2024-07-18T20:53:38.626139938+02:00"}
{"level":"info","msg":"Application Check","time":"2024-07-18T20:53:38.626157605+02:00"}
{"level":"debug","msg":"Timer: tick: 2024-07-18 20:53:39.626220455 +0200 CEST m=+1.014114515","time":"2024-07-18T20:53:39.626388996+02:00"}
{"level":"debug","msg":"Worker: currentTime: 2024-07-18 20:53:39.626220455 +0200 CEST m=+1.014114515","time":"2024-07-18T20:53:39.626560592+02:00"}
{"level":"debug","msg":"Timer: tick: 2024-07-18 20:53:40.626468731 +0200 CEST m=+2.014362792","time":"2024-07-18T20:53:40.626564888+02:00"}
{"level":"debug","msg":"Worker: currentTime: 2024-07-18 20:53:40.626468731 +0200 CEST m=+2.014362792","time":"2024-07-18T20:53:40.626637807+02:00"}
{"level":"debug","msg":"Timer: tick: 2024-07-18 20:53:41.626702184 +0200 CEST m=+3.014596245","time":"2024-07-18T20:53:41.626751011+02:00"}
{"level":"debug","msg":"Worker: currentTime: 2024-07-18 20:53:41.626702184 +0200 CEST m=+3.014596245","time":"2024-07-18T20:53:41.62681458+02:00"}
^C{"level":"debug","msg":"Got 'interrupt' signal","time":"2024-07-18T20:53:42.465091173+02:00"}
{"level":"info","msg":"ApplicationRunner GsdCallback called","time":"2024-07-18T20:53:42.465143349+02:00"}
{"level":"info","msg":"Application Shutdown","time":"2024-07-18T20:53:42.465196688+02:00"}
{"level":"debug","msg":"Timer: Shutdown","time":"2024-07-18T20:53:42.465215522+02:00"}
{"level":"debug","msg":"Worker: Shutdown","time":"2024-07-18T20:53:42.465238456+02:00"}
{"level":"info","msg":"HealthCheck services Shutdown","time":"2024-07-18T20:53:42.465256048+02:00"}
{"level":"debug","msg":"Timer: Shutting down","time":"2024-07-18T20:53:42.465325546+02:00"}
{"level":"debug","msg":"Timer: Stopped","time":"2024-07-18T20:53:42.465411476+02:00"}
{"level":"info","msg":"HealthCheck server is closed","time":"2024-07-18T20:53:42.465351746+02:00"}
{"level":"debug","msg":"Worker: Shutting down","time":"2024-07-18T20:53:42.465357017+02:00"}
{"level":"debug","msg":"Worker: Stopped","time":"2024-07-18T20:53:42.465480977+02:00"}
```
