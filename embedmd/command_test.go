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
			cmd: command{Path: "code.go", Lang: "go", Start: ptr("/start/"), End: ptr("/end/"), Type: typeCode, IncludeStart: true, IncludeEnd: true}},
		{name: "start with replace",
			in:  "(code.go s/.*/b/ /start/ /end/)",
			cmd: command{Path: "code.go", Lang: "go", Start: ptr("/start/"), End: ptr("/end/"), Substitutions: []Substitution{{Pattern: ".*", Replacement: "b"}}, Type: typeCode, IncludeStart: true, IncludeEnd: true}},
		{name: "start with replace with space, escape /, unescape \n in replacement",
			in:  "(code.go s/.* /$embed:{newline}b\\/ / /start/ /end/)",
			cmd: command{Path: "code.go", Lang: "go", Start: ptr("/start/"), End: ptr("/end/"), Substitutions: []Substitution{{Pattern: ".* ", Replacement: "\nb/ "}}, Type: typeCode, IncludeStart: true, IncludeEnd: true}},
		{name: "only start",
			in:  "(code.go     /start/)",
			cmd: command{Path: "code.go", Lang: "go", Start: ptr("/start/"), Type: typeCode, IncludeStart: true, IncludeEnd: true}},
		{name: "only start with replace",
			in:  "(code.go s/.*/b/    /start/)",
			cmd: command{Path: "code.go", Lang: "go", Start: ptr("/start/"), Substitutions: []Substitution{{Pattern: ".*", Replacement: "b"}}, Type: typeCode, IncludeStart: true, IncludeEnd: true}},
		{name: "empty list",
			in:  "()",
			err: "missing file name"},
		{name: "file with no extension and no lang",
			in:  "(test)",
			err: "language is required when file has no extension"},
		{name: "surrounding blanks",
			in:  "   \t  (code.go)  \t  ",
			cmd: command{Path: "code.go", Lang: "go", Type: typeCode, IncludeStart: true, IncludeEnd: true}},
		{name: "all options",
			in: "(code.go noCode noStart noEnd trim trimSuffix:suffix trimPrefix:prefix template:template lang:md s/from/to/ /start/ /end/)",
			cmd: command{
				Path:         "code.go",
				Lang:         "md",
				Type:         typePlain,
				Start:        ptr("/start/"),
				End:          ptr("/end/"),
				IncludeStart: false,
				IncludeEnd:   false,
				Trim:         true,
				TrimPrefix:   "prefix",
				TrimSuffix:   "suffix",
				Template:     "template",
				Substitutions: []Substitution{{
					Pattern:     "from",
					Replacement: "to",
				}},
			}},
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
			cmd: command{Path: "test.md", Lang: "markdown", Type: typeCode, IncludeStart: true, IncludeEnd: true}},
		{name: "file name and language with replace",
			in:  "(test.md markdown s/.*/b/)",
			cmd: command{Path: "test.md", Lang: "markdown", Substitutions: []Substitution{{Pattern: ".*", Replacement: "b"}}, Type: typeCode, IncludeStart: true, IncludeEnd: true}},
		{name: "multi-line comments",
			in:  `(doc.go /\/\*/ /\*\//)`,
			cmd: command{Path: "doc.go", Lang: "go", Start: ptr(`/\/\*/`), End: ptr(`/\*\//`), Type: typeCode, IncludeStart: true, IncludeEnd: true}},
		{name: "using $ as end",
			in:  "(foo.go /start/ $)",
			cmd: command{Path: "foo.go", Lang: "go", Start: ptr("/start/"), End: ptr("$"), Type: typeCode, IncludeStart: true, IncludeEnd: true}},
		{name: "extra arguments",
			in: "(foo.go /start/ $ extra)", err: "too many arguments"},
		{name: "file name with directories",
			in:  "(foo/bar.go)",
			cmd: command{Path: "foo/bar.go", Lang: "go", Type: typeCode, IncludeStart: true, IncludeEnd: true}},
		{name: "url",
			in:  "(http://golang.org/sample.go)",
			cmd: command{Path: "http://golang.org/sample.go", Lang: "go", Type: typeCode, IncludeStart: true, IncludeEnd: true}},
		{name: "bad url",
			in:  "(http://golang:org:sample.go)",
			cmd: command{Path: "http://golang:org:sample.go", Lang: "go", Type: typeCode, IncludeStart: true, IncludeEnd: true}},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := parseCommand(tt.in)
			if !eqErr(t, tt.name, err, tt.err) {
				return
			}
			assert.Equal(t, tt.cmd, *cmd)
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
