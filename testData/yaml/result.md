---
embed:
  src: $lgtm/examples/go/go.mod
  type: plain
  template: |
    ```sh
    go get {{ .Content }}
    ```
  start: "require \\("
  end: "\\)"
  includeStart: false
  includeEnd: false
  trim: true
  trimSuffix: \
  replace:
    - pattern: \s+(\S+) \S+
      replacement: |-
        "$1" \
          
headless: true
description: Instrument Go dependencies
---

```sh
go get "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp" \
  "go.opentelemetry.io/contrib/instrumentation/runtime" \
  "go.opentelemetry.io/otel" \
  "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp" \
  "go.opentelemetry.io/otel/exporters/otlp/otlptrace" \
  "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp" \
  "go.opentelemetry.io/otel/sdk" \
  "go.opentelemetry.io/otel/sdk/metric"
```
