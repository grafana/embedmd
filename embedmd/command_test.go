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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseCommand(t *testing.T) {
	tc := []struct {
		name string
		in   string
		cmd  command
		err  string
	}{
		{name: "start to end",
			in:  "(code.go /start/ /end/)",
			cmd: command{Path: "code.go", Lang: "go", Start: ptr("/start/"), End: ptr("/end/")}},
		{name: "start with replace",
			in:  "(code.go s/.*/b/ /start/ /end/)",
			cmd: command{Path: "code.go", Lang: "go", Start: ptr("/start/"), End: ptr("/end/"), Substitutions: []Substitution{{Pattern: ".*", Replacement: "b"}}}},
		{name: "start with replace with space, escape /, unescape \n in replacement",
			in:  "(code.go s/.* /$embed:{newline}b\\/ / /start/ /end/)",
			cmd: command{Path: "code.go", Lang: "go", Start: ptr("/start/"), End: ptr("/end/"), Substitutions: []Substitution{{Pattern: ".* ", Replacement: "\nb/ "}}}},
		{name: "only start",
			in:  "(code.go     /start/)",
			cmd: command{Path: "code.go", Lang: "go", Start: ptr("/start/")}},
		{name: "only start with replace",
			in:  "(code.go s/.*/b/    /start/)",
			cmd: command{Path: "code.go", Lang: "go", Start: ptr("/start/"), Substitutions: []Substitution{{Pattern: ".*", Replacement: "b"}}}},
		{name: "empty list",
			in:  "()",
			err: "missing file name"},
		{name: "file with no extension and no lang",
			in:  "(test)",
			err: "language is required when file has no extension"},
		{name: "surrounding blanks",
			in:  "   \t  (code.go)  \t  ",
			cmd: command{Path: "code.go", Lang: "go"}},
		{name: "all flags",
			in:  "(code.go noCode noStart noEnd)",
			cmd: command{Path: "code.go", Lang: "go", Type: typePlain, IncludeStart: false, IncludeEnd: false}},
		{name: "no parenthesis",
			in:  "{code.go}",
			err: "argument list should be in parenthesis"},
		{name: "only left parenthesis",
			in:  "(code.go",
			err: "argument list should be in parenthesis"},
		{name: "regexp not closed",
			in:  "(code.go /start)",
			err: "unbalanced /"},
		{name: "end regexp not closed",
			in:  "(code.go /start/ /end)",
			err: "unbalanced /"},
		{name: "file name and language",
			in:  "(test.md markdown)",
			cmd: command{Path: "test.md", Lang: "markdown"}},
		{name: "file name and language with replace",
			in:  "(test.md markdown s/.*/b/)",
			cmd: command{Path: "test.md", Lang: "markdown", Substitutions: []Substitution{{Pattern: ".*", Replacement: "b"}}}},
		{name: "multi-line comments",
			in:  `(doc.go /\/\*/ /\*\//)`,
			cmd: command{Path: "doc.go", Lang: "go", Start: ptr(`/\/\*/`), End: ptr(`/\*\//`)}},
		{name: "using $ as end",
			in:  "(foo.go /start/ $)",
			cmd: command{Path: "foo.go", Lang: "go", Start: ptr("/start/"), End: ptr("$")}},
		{name: "extra arguments",
			in: "(foo.go /start/ $ extra)", err: "too many arguments"},
		{name: "file name with directories",
			in:  "(foo/bar.go)",
			cmd: command{Path: "foo/bar.go", Lang: "go"}},
		{name: "url",
			in:  "(http://golang.org/sample.go)",
			cmd: command{Path: "http://golang.org/sample.go", Lang: "go"}},
		{name: "bad url",
			in:  "(http://golang:org:sample.go)",
			cmd: command{Path: "http://golang:org:sample.go", Lang: "go"}},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := parseCommand(tt.in)
			if !eqErr(t, tt.name, err, tt.err) {
				return
			}

			want, got := tt.cmd, *cmd
			if want.Path != got.Path {
				t.Errorf("case [%s]: expected file %q; got %q", tt.name, want.Path, got.Path)
			}
			if want.Lang != got.Lang {
				t.Errorf("case [%s]: expected language %q; got %q", tt.name, want.Lang, got.Lang)
			}
			assert.Equal(t, want.Substitutions, got.Substitutions)
			if !eqPtr(want.Start, got.Start) {
				t.Errorf("case [%s]: expected start %v; got %v", tt.name, str(want.Start), str(got.Start))
			}
			if !eqPtr(want.End, got.End) {
				t.Errorf("case [%s]: expected end %v; got %v", tt.name, str(want.End), str(got.End))
			}
		})
	}
}

func ptr(s string) *string { return &s }

func str(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func eqPtr(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func eqErr(t *testing.T, id string, err error, msg string) bool {
	if msg == "" {
		assert.NoError(t, err)
		return true
	} else {
		assert.EqualError(t, err, msg)
		return false
	}
}
