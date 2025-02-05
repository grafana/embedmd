// Copyright 2016 Google Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to writing, software distributed
// under the License is distributed on a "AS IS" BASIS, WITHOUT WARRANTIES OR
// CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.

package embedmd

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"io"
	"strings"
	"testing"
)

const yamlCommand = `embed:
  src: https://raw.githubusercontent.com/grafana/docker-otel-lgtm/73272e8995e9c5460d543d0b909317d5877c3855/examples/go/go.mod
  type: plain
  start: require (
  end: )
  includeStart: false
  includeEnd: false
  trim: true
  trimSuffix: \
  template: |
    ` + "```" + `sh
    go get {{ .Content }}
    ` + "```" + `
  replace:
    - pattern: \s+(\S+) \S+
      replacement: |
        $1" \
`

type yamlReceived struct {
	Embed *command `yaml:"embed"`
}

func TestParser(t *testing.T) {
	tc := []struct {
		name string
		in   string
		out  string
		run  commandRunner
		err  string
	}{
		{
			name: "empty file",
			in:   "",
			out:  "",
		},
		{
			name: "just text",
			in:   "one\ntwo\nthree\n",
			out:  "one\ntwo\nthree\n",
		},
		{
			name: "yaml command",
			in:   "---\n" + yamlCommand + "\nheadless: true\n---\none\ntwo\nthree\n",
			out:  "---\n" + yamlCommand + "\nheadless: true\n---\n\nreceived:\n" + yamlCommand,
			run: func(w io.Writer, cmd *command) error {
				fmt.Fprint(w, "received:\n")
				encoder := yaml.NewEncoder(w)
				encoder.SetIndent(2)
				err := encoder.Encode(yamlReceived{Embed: cmd})
				assert.NoError(t, err)
				err = encoder.Close()
				assert.NoError(t, err)

				return nil
			},
		},
		{
			name: "a command",
			in:   "one\n[embedmd]:# (code.go)",
			out:  "one\n[embedmd]:# (code.go)\nOK\n",
			run: func(w io.Writer, cmd *command) error {
				if cmd.Path != "code.go" {
					return fmt.Errorf("bad command")
				}
				fmt.Fprint(w, "OK\n")
				return nil
			},
		},
		{
			name: "a command then some text",
			in: `one
[embedmd]:# (code.go)
` + "```" + `go
main() {
	fmt.Println("hello")
}
` + "```" + `
Yay`,
			out: "one\n[embedmd]:# (code.go)\nOK\nYay\n",
			run: func(w io.Writer, cmd *command) error {
				if cmd.Path != "code.go" {
					return fmt.Errorf("bad command")
				}
				fmt.Fprint(w, "OK\n")
				return nil
			},
		},
		{
			name: "a bad command",
			in:   "one\n[embedmd]:# (code\n",
			err:  "2: argument list should be in parenthesis",
		},
		{
			name: "an ignored command",
			in:   "one\n```\n[embedmd]:# (code.go)\n```\n",
			out:  "one\n```\n[embedmd]:# (code.go)\n```\n",
		},
		{
			name: "unbalanced code section",
			in:   "one\n```\nsome code\n",
			err:  "3: unbalanced code section",
		},
		{
			name: "two contiguous code sections",
			in:   "\n```go\nhello\n```\n```go\nbye\n```\n",
			out:  "\n```go\nhello\n```\n```go\nbye\n```\n",
		},
		{
			name: "two non contiguous code sections",
			in:   "```go\nhello\n```\n\n```go\nbye\n```\n",
			out:  "```go\nhello\n```\n\n```go\nbye\n```\n",
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			err := process(&out, strings.NewReader(tt.in), tt.run)
			if tt.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.out, out.String())
			} else {
				assert.EqualError(t, err, tt.err)
			}
		})
	}
}
