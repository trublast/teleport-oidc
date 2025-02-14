/*
Copyright 2019-2021 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"crypto/x509"
	"fmt"
	"testing"

	"github.com/gravitational/trace"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestUserMessageFromError(t *testing.T) {
	// Behavior is different in debug
	priorLevel := logrus.GetLevel()
	logrus.SetLevel(logrus.InfoLevel)
	t.Cleanup(func() {
		logrus.SetLevel(priorLevel)
	})

	tests := []struct {
		comment   string
		inError   error
		outString string
	}{
		{
			comment:   "outputs x509-specific unknown authority message",
			inError:   trace.Wrap(x509.UnknownAuthorityError{}),
			outString: "WARNING:\n\n  The proxy you are connecting to has presented a",
		},
		{
			comment:   "outputs x509-specific invalid certificate message",
			inError:   trace.Wrap(x509.CertificateInvalidError{}),
			outString: "WARNING:\n\n  The certificate presented by the proxy is invalid",
		},
		{
			comment:   "outputs user message as provided",
			inError:   trace.Errorf("bad thing occurred"),
			outString: "\x1b[31mERROR: \x1b[0mbad thing occurred",
		},
	}

	for _, tt := range tests {
		message := UserMessageFromError(tt.inError)
		require.Contains(t, message, tt.outString)
	}
}

func TestUnixShellQuote(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "emptyString",
			in:   "",
			out:  "",
		},
		{
			name: "noQuote",
			in:   "foo",
			out:  "foo",
		},
		{
			name: "bang",
			in:   "foo!",
			out:  "'foo!'",
		},
		{
			name: "variable",
			in:   "foo$BAR",
			out:  "'foo$BAR'",
		},
		{
			name: "semicolon",
			in:   "foo;bar",
			out:  "'foo;bar'",
		},
		{
			name: "singleQuoteStart",
			in:   "'foo",
			out:  "''\"'\"'foo'",
		},
		{
			name: "singleQuoteMid",
			in:   "foo'bar",
			out:  "'foo'\"'\"'bar'",
		},
		{
			name: "singleQuoteEnd",
			in:   "foo'",
			out:  "'foo'\"'\"''",
		},
		{
			name: "singleQuotesSurrounding",
			in:   "'foo'",
			out:  "''\"'\"'foo'\"'\"''",
		},
		{
			name: "space",
			in:   "foo bar",
			out:  "'foo bar'",
		},
		{
			name: "path",
			in:   "/usr/local/bin",
			out:  "/usr/local/bin",
		},
		{
			name: "commandSubstitution",
			in:   "$(ls -la)",
			out:  "'$(ls -la)'",
		},
		{
			name: "backticks",
			in:   "`echo foo`",
			out:  "'`echo foo`'",
		},
		{
			name: "doubleQuotes",
			in:   "foo\"bar",
			out:  "'foo\"bar'",
		},
		{
			name: "brackets",
			in:   "[1,2,3]",
			out:  "'[1,2,3]'",
		},
		{
			name: "parentheses",
			in:   "(1+2)",
			out:  "'(1+2)'",
		},
		{
			name: "braceExpansion",
			in:   "{a,b}",
			out:  "'{a,b}'",
		},
		{
			name: "escapeCharacters",
			in:   "foo\\bar",
			out:  "'foo\\bar'",
		},
		{
			name: "wildcards",
			in:   "*",
			out:  "'*'",
		},
		{
			name: "pipe",
			in:   "foo | bar",
			out:  "'foo | bar'",
		},
		{
			name: "andOperator",
			in:   "foo && bar",
			out:  "'foo && bar'",
		},
		{
			name: "newline",
			in:   "foo\nbar",
			out:  "'foo\\nbar'",
		},
		{
			name: "carriageReturn",
			in:   "foo\rbar",
			out:  "'foo\\rbar'",
		},
		{
			name: "tab",
			in:   "foo\tbar",
			out:  "'foo\tbar'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.out, UnixShellQuote(tt.in))
		})
	}
}

// TestEscapeControl tests escape control
func TestEscapeControl(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in  string
		out string
	}{
		{
			in:  "hello, world!",
			out: "hello, world!",
		},
		{
			in:  "hello,\nworld!",
			out: `"hello,\nworld!"`,
		},
		{
			in:  "hello,\r\tworld!",
			out: `"hello,\r\tworld!"`,
		},
	}

	for i, tt := range tests {
		require.Equal(t, tt.out, EscapeControl(tt.in), fmt.Sprintf("test case %v", i))
	}
}

// TestAllowWhitespace tests escape control that allows (some) whitespace characters.
func TestAllowWhitespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in  string
		out string
	}{
		{
			in:  "hello, world!",
			out: "hello, world!",
		},
		{
			in:  "hello,\nworld!",
			out: "hello,\nworld!",
		},
		{
			in:  "\thello, world!",
			out: "\thello, world!",
		},
		{
			in:  "\t\thello, world!",
			out: "\t\thello, world!",
		},
		{
			in:  "hello, world!\n",
			out: "hello, world!\n",
		},
		{
			in:  "hello, world!\n\n",
			out: "hello, world!\n\n",
		},
		{
			in:  string([]byte{0x68, 0x00, 0x68}),
			out: "\"h\\x00h\"",
		},
		{
			in:  string([]byte{0x68, 0x08, 0x68}),
			out: "\"h\\bh\"",
		},
		{
			in:  string([]int32{0x00000008, 0x00000009, 0x00000068}),
			out: "\"\\b\"\th",
		},
		{
			in:  string([]int32{0x00000090}),
			out: "\"\\u0090\"",
		},
		{
			in:  "hello,\r\tworld!",
			out: `"hello,\r"` + "\tworld!",
		},
		{
			in:  "hello,\n\r\tworld!",
			out: "hello,\n" + `"\r"` + "\tworld!",
		},
		{
			in:  "hello,\t\n\r\tworld!",
			out: "hello,\t\n" + `"\r"` + "\tworld!",
		},
	}

	for i, tt := range tests {
		require.Equal(t, tt.out, AllowWhitespace(tt.in), fmt.Sprintf("test case %v", i))
	}
}
