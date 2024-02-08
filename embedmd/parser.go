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
	"bufio"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"strings"
)

type commandRunner func(io.Writer, *command) error

func process(out io.Writer, in io.Reader, run commandRunner) error {
	s := &countingScanner{bufio.NewScanner(in), 0}

	state := parsingText
	var err error
	for state != nil {
		state, err = state(out, s, run)
		if err != nil {
			return fmt.Errorf("%d: %v", s.line, err)
		}
	}

	if err := s.Err(); err != nil {
		return fmt.Errorf("%d: %v", s.line, err)
	}
	return nil
}

type countingScanner struct {
	*bufio.Scanner
	line int
}

func (c *countingScanner) Scan() bool {
	b := c.Scanner.Scan()
	if b {
		c.line++
	}
	return b
}

type state func(io.Writer, *countingScanner, commandRunner) (state, error)

func parsingText(out io.Writer, s *countingScanner, run commandRunner) (state, error) {
	if !s.Scan() {
		return nil, nil // end of file, which is fine.
	}

	if s.line == 2 && s.Text() == "embed:" { // line numbers start with 1
		//parse until line with "---" whole file with yaml command
		return yamlParser{}.parse, nil
	}

	switch line := s.Text(); {
	case strings.HasPrefix(line, "[embedmd]:#"):
		return parsingCmd, nil
	case strings.HasPrefix(line, "```"):
		return codeParser{print: true}.parse, nil
	default:
		fmt.Fprintln(out, s.Text())
		return parsingText, nil
	}
}

func parsingCmd(out io.Writer, s *countingScanner, run commandRunner) (state, error) {
	line := s.Text()
	fmt.Fprintln(out, line)
	args := line[strings.Index(line, "#")+1:]
	cmd, err := parseCommand(args)
	if err != nil {
		return nil, err
	}
	if err := run(out, cmd); err != nil {
		return nil, err
	}
	if !s.Scan() {
		return nil, nil // end of file, which is fine.
	}
	if strings.HasPrefix(s.Text(), "```") {
		return codeParser{print: false}.parse, nil
	}
	fmt.Fprintln(out, s.Text())
	return parsingText, nil
}

type codeParser struct{ print bool }

func (c codeParser) parse(out io.Writer, s *countingScanner, run commandRunner) (state, error) {
	if c.print {
		fmt.Fprintln(out, s.Text())
	}
	if !s.Scan() {
		return nil, fmt.Errorf("unbalanced code section")
	}
	if !strings.HasPrefix(s.Text(), "```") {
		return c.parse, nil
	}

	// print the end of the code section if needed and go back to parsing text.
	if c.print {
		fmt.Fprintln(out, s.Text())
	}
	return parsingText, nil
}

type yamlParser struct {
	yaml []string
	drop bool
}

func (c yamlParser) parse(out io.Writer, s *countingScanner, run commandRunner) (state, error) {
	if c.drop {
		if !s.Scan() {
			return nil, nil
		}
	} else {
		fmt.Fprintln(out, s.Text())
		if !s.Scan() {
			return nil, fmt.Errorf("unbalanced yaml section")
		}
		if s.Text() == "---" {
			fmt.Fprintf(out, "%s\n\n", s.Text())
			c.drop = true

			cmd := &command{yamlMode: true}
			err := yaml.Unmarshal([]byte(strings.Join(c.yaml, "\n")), &cmd)
			if err != nil {
				return nil, err
			}

			if err := run(out, cmd); err != nil {
				return nil, err
			}
		} else {
			c.yaml = append(c.yaml, s.Text())
		}
	}
	return c.parse, nil
}
