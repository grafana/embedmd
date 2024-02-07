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
	"errors"
	"path/filepath"
	"strings"
)

type substitution struct {
	pattern     string
	replacement string
}

type parseField struct {
	subs  *substitution
	plain string
}

type command struct {
	path, lang    string
	start, end    *string
	substitutions []substitution
}

func parseCommand(s string) (*command, error) {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '(' || s[len(s)-1] != ')' {
		return nil, errors.New("argument list should be in parenthesis")
	}

	args, err := fields(s[1 : len(s)-1])
	if err != nil {
		return nil, err
	}
	if len(args) == 0 {
		return nil, errors.New("missing file name")
	}

	cmd := &command{path: args[0].plain}
	args = args[1:]
	if len(args) > 0 && args[0].plain != "" && args[0].plain[0] != '/' {
		cmd.lang, args = args[0].plain, args[1:]
	} else {
		ext := filepath.Ext(cmd.path[1:])
		if len(ext) == 0 {
			return nil, errors.New("language is required when file has no extension")
		}
		cmd.lang = ext[1:]
	}

	for {
		if len(args) > 0 && args[0].subs != nil {
			cmd.substitutions = append(cmd.substitutions, *args[0].subs)
			args = args[1:]
		} else {
			break
		}
	}

	switch {
	case len(args) == 1:
		cmd.start = &args[0].plain
	case len(args) == 2:
		cmd.start, cmd.end = &args[0].plain, &args[1].plain
	case len(args) > 2:
		return nil, errors.New("too many arguments")
	}

	return cmd, nil
}

// fields returns a list of the groups of text separated by blanks,
// keeping all text surrounded by / as a group.
func fields(s string) ([]parseField, error) {
	var args []parseField

	for s = strings.TrimSpace(s); len(s) > 0; s = strings.TrimSpace(s) {
		if strings.HasPrefix(s, "s/") {
			// parse substitution pattern s/pattern/replacement/, / can be escaped with \
			patternLen := nextSlash(s[2:])
			if patternLen < 0 {
				return nil, errors.New("unbalanced /")
			}
			subsLen := nextSlash(s[patternLen+3:])
			if subsLen < 0 {
				return nil, errors.New("unbalanced /")
			}

			l := patternLen + subsLen + 4
			args, s = append(args, parseField{subs: &substitution{
				pattern:     unescapeSlash(s[2 : patternLen+2]),
				replacement: unescapeSlash(s[patternLen+3 : l-1]),
			}}), s[l:]
		} else if s[0] == '/' {
			sep := nextSlash(s[1:])
			if sep < 0 {
				return nil, errors.New("unbalanced /")
			}
			args, s = append(args, parseField{plain: s[:sep+2]}), s[sep+2:]
		} else {
			sep := strings.IndexByte(s[1:], ' ')
			if sep < 0 {
				return append(args, parseField{plain: s}), nil
			}
			args, s = append(args, parseField{plain: s[:sep+1]}), s[sep+1:]
		}
	}

	return args, nil
}

func unescapeSlash(s2 string) string {
	return strings.ReplaceAll(s2, "\\/", "/")
}

// nextSlash will find the index of the next unescaped slash in a string.
func nextSlash(s string) int {
	for sep := 0; ; sep++ {
		i := strings.IndexByte(s[sep:], '/')
		if i < 0 {
			return -1
		}
		sep += i
		if sep == 0 || s[sep-1] != '\\' {
			return sep
		}
	}
}
