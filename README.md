[![Build Status](https://travis-ci.org/campoy/embedmd.svg)](https://travis-ci.org/campoy/embedmd) [![Go Report Card](https://goreportcard.com/badge/github.com/campoy/embedmd)](https://goreportcard.com/report/github.com/campoy/embedmd)


# embedmd

Are you tired of copy pasting your code into your `README.md` file, just to
forget about it later on and have unsynced copies? Or even worse, code
that does not even compile?

Then `embedmd` is for you!

`embedmd` embeds files or fractions of files into Markdown files. It does
so by searching `embedmd` commands, which are a subset of the Markdown
syntax for comments. This means they are invisible when Markdown is
rendered, so they can be kept in the file as pointers to the origin of
the embedded text.

The command receives a list of Markdown files. If no list is given, the command
reads from the standard input.

The format of an `embedmd` command is:

```Markdown
[embedmd]:# (pathOrURL language /start regexp/ /end regexp/)
```

The embedded code will be extracted from the file at `pathOrURL`,
which can either be a relative path to a file in the local file
system (using always forward slashes as directory separator) or
a URL starting with `http://` or `https://`.
If the `pathOrURL` is a URL the tool will fetch the content in that URL.
The embedded content starts at the first line that matches `/start regexp/`
and finishes at the first line matching `/end regexp/`.

Omitting the the second regular expression will embed only the piece of text
that matches `/regexp/`:

```Markdown
[embedmd]:# (pathOrURL language /regexp/)
```

To embed the whole line matching a regular expression you can use:

```Markdown
[embedmd]:# (pathOrURL language /.*regexp.*/)
```

To embed from a point to the end you should use:

```Markdown
[embedmd]:# (pathOrURL language /start regexp/ $)
```

To perform substitutions, use `s/regex/to/`:

```Markdown
[embedmd]:# (pathOrURL language s/regex/to/ /start regexp/ $)
```

To embed a whole file, omit both regular expressions:

```Markdown
[embedmd]:# (pathOrURL language)
```

You can omit the language in any of the previous commands, and the extension
of the file will be used for the snippet syntax highlighting.

This works when the file extensions matches the name of the language (like Go
files, since `.go` matches `go`). However, this will fail with other files like
`.md` whose language name is `markdown`.

```Markdown
[embedmd]:# (file.ext)
```

You can use the following options to modify the behavior of `embedmd`:

[embedmd]:# (pathOrURL <flags> language s/regex/to/ /start regexp/ /end regexp/)
                                                                                 
Unary flags:
* `noCode`: Do not wrap the embedded content in a code block.
* `noStart`: Do not include the content that matches the start regular expression.
* `noEnd`: Do not include the content that matches the end regular expression.
* `trim`: Trim the content before embedding it.

Options in the form of `key:value`:
* `lang`: The language of the embedded content.
* `template`: A template to use to format the content. It uses Go's text/template package.
* `trimPrefix`: A string to trim from the start.
* `trimSuffix`: A string to trim from the end.

## Installation

> You can install Go by following [these instructions](https://golang.org/doc/install).

`embedmd` is written in Go, so if you have Go installed you can install it with
`go get`:

```
go get github.com/campoy/embedmd
```

This will download the code, compile it, and leave an `embedmd` binary
in `$GOPATH/bin`.

Eventually, and if there's enough interest, I will provide binaries for
every OS and architecture out there ... _eventually_.

## Usage:

Given the two files in [sample](sample):

*hello.go:*

[embedmd]:# (sample/hello.go)
```go
// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Hello, there, it is", time.Now())
}
```

*docs.md:*

[embedmd]:# (sample/docs.md Markdown /./ /embedmd.*time.*/)
```Markdown
# A hello world in Go

Go is very simple, here you can see a whole "hello, world" program.

[embedmd]:# (hello.go)

We can try to embed a file from a directory.

[embedmd]:# (test/hello.go /func main/ $)

You always start with a `package` statement like:

[embedmd]:# (hello.go /package.*/)

Followed by an `import` statement:

[embedmd]:# (hello.go /import/ /\)/)

You can also see how to get the current time:

[embedmd]:# (hello.go /time\.[^)]*\)/)
```

### YAML Mode

`embedmd` also supports YAML mode, which is useful for more complex
scenarios. For example, the following command will embed the content of
`go.mod` into the `docs.md` file, and then use a template to generate a
`go get` command:

```markdown
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


```

will result in the following:

```markdown
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

`` ` ```` ` ```` ` ``sh
go get "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp" \
  "go.opentelemetry.io/contrib/instrumentation/runtime" \
  "go.opentelemetry.io/otel" \
  "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp" \
  "go.opentelemetry.io/otel/exporters/otlp/otlptrace" \
  "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp" \
  "go.opentelemetry.io/otel/sdk" \
  "go.opentelemetry.io/otel/sdk/metric"
`` ` ```` ` ```` ` ``
```
         
All the options are optional, and the following is a list of all the
options available - these are the same options as with the embedded mode:

* `src`: The source of the content to embed. It can be a path or a URL - and can start with a mount point like `$lgtm` (see [flags](#flags)).
* `type`: The type of the content formatting. It can be `plain` or `code`.
* `lang`: The language of the content.
* `template`: A template to use to format the content. It uses Go's text/template package.
* `start`: A regular expression to match the start of the content to embed.
* `end`: A regular expression to match the end of the content to embed. If not provided, the content will be the `start` expression.
* `includeStart`: Whether to include the line that matches the `start` expression.
* `includeEnd`: Whether to include the line that matches the `end` expression.
* `trim`: Whether to trim the content (trim space at start and end).
* `trimPrefix`: A string to trim from the start.
* `trimSuffix`: A string to trim from the end. 
* `replace`: A list of replacements to perform on the content (see example above).

# Flags

* `-w`: Executing `embedmd -w docs.md` will modify `docs.md`
and add the corresponding code snippets, as shown in
[sample/result.md](sample/result.md).

* `-d`: Executing `embedmd -d docs.md` will display the difference
between the contents of `docs.md` and the output of
`embedmd docs.md`.

* `-m`: Register a mount point. For example, `embedmd -m $lgtm=https://raw.githubusercontent.com/lgtmco/lgtm/master` will allow you to use `$lgtm/examples/go.mod` as a mount point in the `src` field. The result will be the same as if you had used `https://raw.githubusercontent.com/lgtmco/lgtm/master/examples/go/go.mod`.

### Disclaimer

This is not an official Google product (experimental or otherwise), it is just
code that happens to be owned by Google.
