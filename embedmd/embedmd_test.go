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
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const content = `
package main

import "fmt"

func main() {
        fmt.Println("hello, test")
}
`

func TestExtract(t *testing.T) {
	tc := []struct {
		name           string
		start, end     *string
		noStart, noEnd bool
		out            string
		err            string
	}{
		{name: "no limits",
			out: content},
		{name: "only one line",
			start: ptr("/func main.*\n/"), out: "func main() {\n"},
		{name: "from package to end",
			start: ptr("/package main/"), end: ptr("$"), out: content[1:]},
		{name: "from package to end - skip start",
			start: ptr("/package main/"), noStart: true, end: ptr("$"), out: content[13:]},
		{name: "not matching",
			start: ptr("/gopher/"), err: "could not match \"/gopher/\""},
		{name: "part of a line",
			start: ptr("/fmt.P/"), end: ptr("/hello/"), out: "fmt.Println(\"hello"},
		{name: "part of a line - skip end",
			start: ptr("/fmt.P/"), end: ptr("/hello/"), noEnd: true, out: "fmt.Println(\""},
		{name: "function call",
			start: ptr("/fmt\\.[^()]*/"), out: "fmt.Println"},
		{name: "from fmt to end of line",
			start: ptr("/fmt.P.*\n/"), out: "fmt.Println(\"hello, test\")\n"},
		{name: "from func to end of next line",
			start: ptr("/func/"), end: ptr("/Println.*\n/"), out: "func main() {\n        fmt.Println(\"hello, test\")\n"},
		{name: "from func to }",
			start: ptr("/func main/"), end: ptr("/}/"), out: "func main() {\n        fmt.Println(\"hello, test\")\n}"},

		{name: "bad start regexp",
			start: ptr("/(/"), err: "error parsing regexp: missing closing ): `(`"},
		{name: "bad regexp",
			start: ptr("something"), err: "missing slashes (/) around \"something\""},
		{name: "bad end regexp",
			start: ptr("/fmt.P/"), end: ptr("/)/"), err: "error parsing regexp: unexpected ): `)`"},

		{name: "start and end of line ^$",
			start: ptr("/^func main/"), end: ptr("/}$/"), out: "func main() {\n        fmt.Println(\"hello, test\")\n}"},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			b, err := extract([]byte(content),
				&command{
					start:   tt.start,
					end:     tt.end,
					noStart: tt.noStart,
					noEnd:   tt.noEnd,
				})
			if tt.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.err)
			}
			assert.Equal(t, tt.out, string(b))
		})
	}
}

func TestExtractFromFile(t *testing.T) {
	tc := []struct {
		name    string
		cmd     command
		baseDir string
		files   map[string][]byte
		out     string
		err     string
	}{
		{
			name:  "extract the whole file",
			cmd:   command{path: "code.go", lang: "go"},
			files: map[string][]byte{"code.go": []byte(content)},
			out:   "```go\n" + string(content) + "```\n",
		},
		{
			name:  "no code",
			cmd:   command{path: "code.go", lang: "go", noCode: true},
			files: map[string][]byte{"code.go": []byte(content)},
			out:   content,
		},
		{
			name:    "extract the whole from a different directory",
			cmd:     command{path: "code.go", lang: "go"},
			baseDir: "sample",
			files:   map[string][]byte{"sample/code.go": []byte(content)},
			out:     "```go\n" + string(content) + "```\n",
		},
		{
			name:  "added line break",
			cmd:   command{path: "code.go", lang: "go", start: ptr("/fmt\\.Println/")},
			files: map[string][]byte{"code.go": []byte(content)},
			out:   "```go\nfmt.Println\n```\n",
		},
		{
			name: "missing file",
			cmd:  command{path: "code.go", lang: "go"},
			err:  "could not read code.go: file does not exist",
		},
		{
			name:  "unmatched regexp",
			cmd:   command{path: "code.go", lang: "go", start: ptr("/potato/")},
			files: map[string][]byte{"code.go": []byte(content)},
			err:   "could not extract content from code.go: could not match \"/potato/\"",
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			e := embedder{
				baseDir: tt.baseDir,
				Fetcher: fakeFileProvider(tt.files),
			}

			w := new(bytes.Buffer)
			err := e.runCommand(w, &tt.cmd)
			if !eqErr(t, tt.name, err, tt.err) {
				return
			}
			if w.String() != tt.out {
				t.Errorf("case [%s]: expected output\n%q\n; got \n%q\n", tt.name, tt.out, w.String())
			}
		})
	}
}

type fakeFileProvider map[string][]byte

func (c fakeFileProvider) Fetch(dir, path string) ([]byte, error) {
	if f, ok := c[filepath.Join(dir, path)]; ok {
		return f, nil
	}
	return nil, os.ErrNotExist
}

func TestProcess(t *testing.T) {
	tc := []struct {
		name  string
		in    string
		dir   string
		files map[string][]byte
		urls  map[string][]byte
		out   string
		err   string
		diff  bool
	}{
		{
			name: "missing file",
			in: "# This is some markdown\n" +
				"[embedmd]:# (code.go)\n" +
				"Yay!\n",
			err: "2: could not read code.go: file does not exist",
		},
		{
			name: "generating code for first time",
			in: "# This is some markdown\n" +
				"[embedmd]:# (code.go)\n" +
				"Yay!\n",
			files: map[string][]byte{"code.go": []byte(content)},
			out: "# This is some markdown\n" +
				"[embedmd]:# (code.go)\n" +
				"```go\n" +
				string(content) +
				"```\n" +
				"Yay!\n",
		},
		{
			name: "generating code for first time with base dir",
			dir:  "sample",
			in: "# This is some markdown\n" +
				"[embedmd]:# (code.go)\n" +
				"Yay!\n",
			files: map[string][]byte{"sample/code.go": []byte(content)},
			out: "# This is some markdown\n" +
				"[embedmd]:# (code.go)\n" +
				"```go\n" +
				string(content) +
				"```\n" +
				"Yay!\n",
		},
		{
			name: "replacing existing code",
			in: "# This is some markdown\n" +
				"[embedmd]:# (code.go)\n" +
				"```go\n" +
				string(content) +
				"```\n" +
				"Yay!\n",
			files: map[string][]byte{"code.go": []byte(content)},
			out: "# This is some markdown\n" +
				"[embedmd]:# (code.go)\n" +
				"```go\n" +
				string(content) +
				"```\n" +
				"Yay!\n",
		},
		{
			name: "embedding code from a URL",
			in: "# This is some markdown\n" +
				"[embedmd]:# (https://fakeurl.com/main.go)\n" +
				"Yay!\n",
			urls: map[string][]byte{"https://fakeurl.com/main.go": []byte(content)},
			out: "# This is some markdown\n" +
				"[embedmd]:# (https://fakeurl.com/main.go)\n" +
				"```go\n" +
				string(content) +
				"```\n" +
				"Yay!\n",
		},
		{
			name: "embedding code from a URL not found",
			in: "# This is some markdown\n" +
				"[embedmd]:# (https://fakeurl.com/main.go)\n" +
				"Yay!\n",
			err: "2: could not read https://fakeurl.com/main.go: status Not Found",
		},
		{
			name: "embedding code from a bad URL",
			in: "# This is some markdown\n" +
				"[embedmd]:# (https://fakeurl.com\\main.go)\n" +
				"Yay!\n",
			err: "2: could not read https://fakeurl.com\\main.go: parse \"https://fakeurl.com\\\\main.go\": invalid character \"\\\\\" in host name",
		},
		{
			name: "ignore commands in code blocks",
			in: "# This is some markdown\n" +
				"```markdown\n" +
				"[embedmd]:# (nothing.md)\n" +
				"```\n" +
				"Yay!\n",
			out: "# This is some markdown\n" +
				"```markdown\n" +
				"[embedmd]:# (nothing.md)\n" +
				"```\n" +
				"Yay!\n",
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			cp := mixedContentProvider{tt.files, tt.urls}
			if tt.diff {
				cp.files["file.md"] = []byte(tt.in)
			}
			opts := []Option{WithFetcher(cp)}
			if tt.dir != "" {
				opts = append(opts, WithBaseDir(tt.dir))
			}
			err := Process(&out, strings.NewReader(tt.in), opts...)
			if !eqErr(t, tt.name, err, tt.err) {
				return
			}
			if tt.out != out.String() {
				t.Errorf("case [%s]: expected output:\n###\n%s\n###; got###\n%s\n###", tt.name, tt.out, out.String())
			}
		})
	}
}

func TestReplace(t *testing.T) {
	tc := []struct {
		name  string
		value string
		subs  []substitution
		out   string
	}{
		{
			name:  "one line with single",
			value: "func main() {",
			subs: []substitution{{
				pattern:     "\\(",
				replacement: "[",
			}},
			out: "func main[) {",
		},
		{
			name:  "one line with multiple",
			value: "func main() {",
			subs: []substitution{{
				pattern:     "[()]",
				replacement: "[",
			}},
			out: "func main[[ {",
		},
		{
			name:  "use variables",
			value: "func main() {",
			subs: []substitution{{
				pattern:     "func (\\S+) {",
				replacement: "$1",
			}},
			out: "main()",
		},
		{
			name:  "multi line with multiple",
			value: content,
			subs: []substitution{{
				pattern:     "[()]",
				replacement: "[",
			}},
			out: `
package main

import "fmt"

func main[[ {
        fmt.Println["hello, test"[
}
`,
		},
		{
			name:  "multi line match",
			value: content,
			subs: []substitution{{
				pattern:     "main\n\n",
				replacement: "foo",
			}},
			out: `
package fooimport "fmt"

func main() {
        fmt.Println("hello, test")
}
`,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			b, err := replace([]byte(tt.value), tt.subs)
			assert.NoError(t, err)
			assert.Equal(t, tt.out, string(b))
		})
	}
}

type mixedContentProvider struct {
	files, urls map[string][]byte
}

func (c mixedContentProvider) Fetch(dir, path string) ([]byte, error) {
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		path = filepath.Join(dir, filepath.FromSlash(path))
		if f, ok := c.files[path]; ok {
			return f, nil
		}
		return nil, os.ErrNotExist
	}

	_, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	if b, ok := c.urls[path]; ok {
		return b, nil
	}
	return nil, fmt.Errorf("status Not Found")
}
