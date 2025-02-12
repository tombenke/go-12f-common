# scheduler example

This application demonstrates how to create system components, that are running as concurrent processes, and communicate with each other.

The two components are:

- [timer](timer/): Emits an event every `time-step` period, that triggers the sending of the actual time value to the worker process.
- [worker](worker): Receives the time message, and prints it.

Get help on parameters:

```bash
$ examples/scheduler/scheduler --help
Usage:
   [flags]

Flags:
      --health-check-port uint              The HTTP port of the healthcheck endpoints (default 8080)
  -h, --help                                help for this command
      --liveness-check-path string          The path of the liveness check endpoint (default "/live")
  -f, --log-format string                   The log format: json | text (default "json")
  -l, --log-level string                    The log level: panic | fatal | error | warning | info | debug | trace (default "info")
      --otel-exporter-prometheus-port int   the port used by the Prometheus exporter (default 9464)
      --otel-metrics-exporter string        Selects the exporter to use for metrics: otlp | prometheus | console | none (default "none")
      --otel-traces-exporter string         Selects the exporter to use for tracing: otlp | console | none (default "none")
      --otel-traces-sampler string          Specifies the Sampler used to sample traces by the SDK.
                                            One of: always_on | always_off | traceidratio | parentbased_always_on | parentbased_always_off |
                                            parentbased_traceidratio | parentbased_jaeger_remote | jaeger_remote | xray (default "parentbased_always_on")
      --otel-traces-sampler-arg string      Specifies arguments, if applicable, to the sampler defined in by --otel-traces-sampler
      --readiness-check-path string         The path of the readiness check endpoint (default "/ready")
      --time-step string                    The size of a time-step (default "60s")
```

Run the application with debug log level:

```bash
examples/scheduler/scheduler -l debug --time-step 1s
{"time":"2025-02-12T13:02:08.271287963+01:00","level":"INFO","msg":"Creating Application","config":{}}
{"time":"2025-02-12T13:02:08.271339452+01:00","level":"DEBUG","msg":"Starting 12f application","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","config":{"LogLevel":"debug","LogFormat":"json","HealthCheckPort":8080,"LivenessCheckPath":"/live","ReadinessCheckPath":"/ready","OtelConfig":{"OtelTracesExporter":"none","OtelTracesSampler":"parentbased_always_on","OtelTracesSamplerArg":"","OtelMetricsExporter":"none","OtelExporterPrometheusPort":9464}}}
{"time":"2025-02-12T13:02:08.271390149+01:00","level":"INFO","msg":"Starting up","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"HealthCheck"}
{"time":"2025-02-12T13:02:08.271393596+01:00","level":"DEBUG","msg":"Adding endpoint","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"HealthCheck","path":"/live"}
{"time":"2025-02-12T13:02:08.271406945+01:00","level":"DEBUG","msg":"Adding endpoint","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"HealthCheck","path":"/ready"}
{"time":"2025-02-12T13:02:08.281672975+01:00","level":"INFO","msg":"Checking if server started...","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"HealthCheck"}
{"time":"2025-02-12T13:02:08.282821195+01:00","level":"DEBUG","msg":"Liveness check","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de"}
{"time":"2025-02-12T13:02:08.282972989+01:00","level":"DEBUG","msg":"Checking response","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"HealthCheck","statusCode":200}
{"time":"2025-02-12T13:02:08.283029301+01:00","level":"INFO","msg":"Server is up and running","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"HealthCheck"}
{"time":"2025-02-12T13:02:08.283074594+01:00","level":"INFO","msg":"HealthCheck is up and running!","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"HealthCheck"}
{"time":"2025-02-12T13:02:08.283114789+01:00","level":"INFO","msg":"Starting up","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Otel"}
{"time":"2025-02-12T13:02:08.283213292+01:00","level":"INFO","msg":"Startup Metrics","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Otel","exporter":"none"}
{"time":"2025-02-12T13:02:08.283228107+01:00","level":"INFO","msg":"Startup Tracing","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Otel","exporter":"none"}
{"time":"2025-02-12T13:02:08.283242953+01:00","level":"DEBUG","msg":"Startup","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Timer","config":{"TimeStep":"1s"}}
{"time":"2025-02-12T13:02:08.283287358+01:00","level":"DEBUG","msg":"Starting ticker","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Timer","duration":"1s"}
{"time":"2025-02-12T13:02:08.283313104+01:00","level":"DEBUG","msg":"Startup","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Worker","config":{}}
{"time":"2025-02-12T13:02:08.283366982+01:00","level":"INFO","msg":"Check","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Timer"}
{"time":"2025-02-12T13:02:08.283378577+01:00","level":"INFO","msg":"Check","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Worker"}
{"time":"2025-02-12T13:02:08.308800342+01:00","level":"INFO","msg":"Check","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Timer"}
{"time":"2025-02-12T13:02:08.308834809+01:00","level":"INFO","msg":"Check","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Worker"}
{"time":"2025-02-12T13:02:08.308861895+01:00","level":"INFO","msg":"AfterStartup","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","app":"Application"}
{"time":"2025-02-12T13:02:08.30886956+01:00","level":"DEBUG","msg":"BuildInfo","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","app":"Application","AppName":"examples/scheduler/scheduler","Version":"v1.0.0-27-gee7dfc5"}
{"time":"2025-02-12T13:02:09.283957796+01:00","level":"DEBUG","msg":"Tick","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Timer","currentTime":"2025-02-12T13:02:09.283300118+01:00"}
{"time":"2025-02-12T13:02:09.284107676+01:00","level":"DEBUG","msg":"Tick","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Worker","currentTime":"2025-02-12T13:02:09.283300118+01:00"}
{"time":"2025-02-12T13:02:10.284320355+01:00","level":"DEBUG","msg":"Tick","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Timer","currentTime":"2025-02-12T13:02:10.283299952+01:00"}
{"time":"2025-02-12T13:02:10.28443593+01:00","level":"DEBUG","msg":"Tick","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Worker","currentTime":"2025-02-12T13:02:10.283299952+01:00"}
{"time":"2025-02-12T13:02:11.283794436+01:00","level":"DEBUG","msg":"Tick","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Timer","currentTime":"2025-02-12T13:02:11.283300246+01:00"}
{"time":"2025-02-12T13:02:11.283842365+01:00","level":"DEBUG","msg":"Tick","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Worker","currentTime":"2025-02-12T13:02:11.283300246+01:00"}
^C{"time":"2025-02-12T13:02:12.175901823+01:00","level":"DEBUG","msg":"Got signal","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","signal":2}
{"time":"2025-02-12T13:02:12.176038127+01:00","level":"INFO","msg":"GsdCallback called","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de"}
{"time":"2025-02-12T13:02:12.17606861+01:00","level":"INFO","msg":"BeforeShutdown","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","app":"Application"}
{"time":"2025-02-12T13:02:12.176085565+01:00","level":"DEBUG","msg":"Shutdown","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Worker"}
{"time":"2025-02-12T13:02:12.176099381+01:00","level":"DEBUG","msg":"Shutdown","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Timer"}
{"time":"2025-02-12T13:02:12.176112303+01:00","level":"INFO","msg":"Shutdown","component":"Otel"}
{"time":"2025-02-12T13:02:12.176124545+01:00","level":"INFO","msg":"Shutdown","component":"Otel.Metrics"}
{"time":"2025-02-12T13:02:12.176113343+01:00","level":"DEBUG","msg":"Tick","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Worker","currentTime":"0001-01-01T00:00:00Z"}
{"time":"2025-02-12T13:02:12.176142444+01:00","level":"DEBUG","msg":"Shutting down","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Timer"}
{"time":"2025-02-12T13:02:12.176155355+01:00","level":"DEBUG","msg":"Tick","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Worker","currentTime":"0001-01-01T00:00:00Z"}
{"time":"2025-02-12T13:02:12.176161173+01:00","level":"DEBUG","msg":"Stopped","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Timer"}
{"time":"2025-02-12T13:02:12.176167731+01:00","level":"DEBUG","msg":"Shutting down","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Worker"}
{"time":"2025-02-12T13:02:12.17619469+01:00","level":"DEBUG","msg":"Stopped","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"Worker"}
{"time":"2025-02-12T13:02:12.176133789+01:00","level":"INFO","msg":"Shutdown","component":"Otel.Tracer"}
{"time":"2025-02-12T13:02:12.176234406+01:00","level":"INFO","msg":"Shutdown","component":"HealthCheck"}
{"time":"2025-02-12T13:02:12.17634077+01:00","level":"INFO","msg":"Server closed","appId":"0e0be49c-caef-4aa4-8bea-e33cdc61c0de","component":"HealthCheck"}
```

This application also demonstrates how to use the OTEL metrics, and how to use a counter meter instrument.

Run the scheduler application with `console` metric exporter to test how the OTEL configuration parameters are working:

```bash
OTEL_SERVICE_NAME=hubcontrol:scheduler OTEL_RESOURCE_ATTRIBUTES=service.instance.id=b9d7402f-358c-4909-8e2f-66b3d2f5a6a8 ./examples/scheduler/scheduler --otel-metrics-exporter console --time-step 5s

{"time":"2025-02-10T18:13:45.812187525+01:00","level":"INFO","msg":"Creating Application","config":{}}
{"time":"2025-02-10T18:13:45.812249813+01:00","level":"INFO","msg":"Starting 12f application","appId":"16dd0c1f-6985-498a-a3ed-566ee905b391"}
{"time":"2025-02-10T18:13:45.812294432+01:00","level":"INFO","msg":"Starting up","appId":"16dd0c1f-6985-498a-a3ed-566ee905b391","component":"HealthCheck"}
{"time":"2025-02-10T18:13:45.822586044+01:00","level":"INFO","msg":"Checking if server started...","appId":"16dd0c1f-6985-498a-a3ed-566ee905b391","component":"HealthCheck"}
{"time":"2025-02-10T18:13:45.823355568+01:00","level":"INFO","msg":"Server is up and running","appId":"16dd0c1f-6985-498a-a3ed-566ee905b391","component":"HealthCheck"}
{"time":"2025-02-10T18:13:45.82336999+01:00","level":"INFO","msg":"HealthCheck is up and running!","appId":"16dd0c1f-6985-498a-a3ed-566ee905b391","component":"HealthCheck"}
{"time":"2025-02-10T18:13:45.823384289+01:00","level":"INFO","msg":"Starting up","appId":"16dd0c1f-6985-498a-a3ed-566ee905b391","component":"Otel"}
{"time":"2025-02-10T18:13:45.823427264+01:00","level":"INFO","msg":"Startup Metrics","appId":"16dd0c1f-6985-498a-a3ed-566ee905b391","component":"Otel","exporter":"console"}
{"time":"2025-02-10T18:13:45.82348072+01:00","level":"INFO","msg":"Startup Tracer","appId":"16dd0c1f-6985-498a-a3ed-566ee905b391","component":"Otel"}
{"time":"2025-02-10T18:13:45.823540939+01:00","level":"INFO","msg":"Check","appId":"16dd0c1f-6985-498a-a3ed-566ee905b391","component":"Timer"}
{"time":"2025-02-10T18:13:45.823546078+01:00","level":"INFO","msg":"Check","appId":"16dd0c1f-6985-498a-a3ed-566ee905b391","component":"Worker"}
{"time":"2025-02-10T18:13:45.848795071+01:00","level":"INFO","msg":"Check","appId":"16dd0c1f-6985-498a-a3ed-566ee905b391","component":"Timer"}
{"time":"2025-02-10T18:13:45.848827478+01:00","level":"INFO","msg":"Check","appId":"16dd0c1f-6985-498a-a3ed-566ee905b391","component":"Worker"}
{"time":"2025-02-10T18:13:45.848843553+01:00","level":"INFO","msg":"AfterStartup","appId":"16dd0c1f-6985-498a-a3ed-566ee905b391","app":"Application"}

{"Resource":[{"Key":"service.instance.id","Value":{"Type":"STRING","Value":"b9d7402f-358c-4909-8e2f-66b3d2f5a6a8"}},{"Key":"service.name","Value":{"Type":"STRING","Value":"hubcontrol:scheduler"}},{"Key":"service.version","Value":{"Type":"STRING","Value":"v1.0.0-24-g7f69aad"}},{"Key":"telemetry.sdk.language","Value":{"Type":"STRING","Value":"go"}},{"Key":"telemetry.sdk.name","Value":{"Type":"STRING","Value":"opentelemetry"}},{"Key":"telemetry.sdk.version","Value":{"Type":"STRING","Value":"1.34.0"}}],"ScopeMetrics":[]}

{"Resource":[{"Key":"service.instance.id","Value":{"Type":"STRING","Value":"b9d7402f-358c-4909-8e2f-66b3d2f5a6a8"}},{"Key":"service.name","Value":{"Type":"STRING","Value":"hubcontrol:scheduler"}},{"Key":"service.version","Value":{"Type":"STRING","Value":"v1.0.0-24-g7f69aad"}},{"Key":"telemetry.sdk.language","Value":{"Type":"STRING","Value":"go"}},{"Key":"telemetry.sdk.name","Value":{"Type":"STRING","Value":"opentelemetry"}},{"Key":"telemetry.sdk.version","Value":{"Type":"STRING","Value":"1.34.0"}}],"ScopeMetrics":[{"Scope":{"Name":"worker-run-count","Version":"","SchemaURL":"","Attributes":null},"Metrics":[{"Name":"run","Description":"The number of times the worker run","Unit":"","Data":{"DataPoints":[{"Attributes":[],"StartTime":"2025-02-10T18:13:45.823523323+01:00","Time":"2025-02-10T18:13:51.823641272+01:00","Value":1}],"Temporality":"CumulativeTemporality","IsMonotonic":true}}]}]}

^C{"time":"2025-02-10T18:13:52.983347391+01:00","level":"INFO","msg":"GsdCallback called","appId":"16dd0c1f-6985-498a-a3ed-566ee905b391"}
{"time":"2025-02-10T18:13:52.98338766+01:00","level":"INFO","msg":"BeforeShutdown","appId":"16dd0c1f-6985-498a-a3ed-566ee905b391","app":"Application"}
{"time":"2025-02-10T18:13:52.98340359+01:00","level":"INFO","msg":"Shutdown","component":"Otel"}
{"time":"2025-02-10T18:13:52.983418198+01:00","level":"INFO","msg":"Shutdown","component":"Otel.Metrics"}
{"Resource":[{"Key":"service.instance.id","Value":{"Type":"STRING","Value":"b9d7402f-358c-4909-8e2f-66b3d2f5a6a8"}},{"Key":"service.name","Value":{"Type":"STRING","Value":"hubcontrol:scheduler"}},{"Key":"service.version","Value":{"Type":"STRING","Value":"v1.0.0-24-g7f69aad"}},{"Key":"telemetry.sdk.language","Value":{"Type":"STRING","Value":"go"}},{"Key":"telemetry.sdk.name","Value":{"Type":"STRING","Value":"opentelemetry"}},{"Key":"telemetry.sdk.version","Value":{"Type":"STRING","Value":"1.34.0"}}],"ScopeMetrics":[{"Scope":{"Name":"worker-run-count","Version":"","SchemaURL":"","Attributes":null},"Metrics":[{"Name":"run","Description":"The number of times the worker run","Unit":"","Data":{"DataPoints":[{"Attributes":[],"StartTime":"2025-02-10T18:13:45.823523323+01:00","Time":"2025-02-10T18:13:52.983440356+01:00","Value":1}],"Temporality":"CumulativeTemporality","IsMonotonic":true}}]}]}
{"time":"2025-02-10T18:13:52.983496393+01:00","level":"INFO","msg":"Shutdown","component":"Otel.Tracer"}
{"time":"2025-02-10T18:13:52.983502236+01:00","level":"INFO","msg":"Shutdown","component":"HealthCheck"}
{"time":"2025-02-10T18:13:52.983547219+01:00","level":"INFO","msg":"Server closed","appId":"16dd0c1f-6985-498a-a3ed-566ee905b391","component":"HealthCheck"}
```

