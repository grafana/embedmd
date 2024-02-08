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

