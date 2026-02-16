package util

import "testing"

func TestShellQuote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string returns single-quoted empty",
			input: "",
			want:  "''",
		},
		{
			name:  "safe alphanumeric string unchanged",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "safe alphanumeric string unchanged",
			input: "hello\nworld",
			want:  "'hello\nworld'",
		},
		{
			name:  "safe chars unchanged",
			input: "file@host:path/to/dir,v=1+2",
			want:  "file@host:path/to/dir,v=1+2",
		},
		{
			name:  "safe percent and equals",
			input: "key=value%20",
			want:  "key=value%20",
		},
		{
			name:  "string with spaces is quoted",
			input: "hello world",
			want:  "'hello world'",
		},
		{
			name:  "string with tab is quoted",
			input: "hello\tworld",
			want:  "'hello\tworld'",
		},
		{
			name:  "string with single quote is escaped",
			input: "it's",
			want:  "'it'\"'\"'s'",
		},
		{
			name:  "string with double quote is quoted",
			input: `say "hello"`,
			want:  `'say "hello"'`,
		},
		{
			name:  "string with semicolon is quoted",
			input: "cmd; rm -rf /",
			want:  "'cmd; rm -rf /'",
		},
		{
			name:  "string with dollar sign is quoted",
			input: "$HOME",
			want:  "'$HOME'",
		},
		{
			name:  "string with backtick is quoted",
			input: "`whoami`",
			want:  "'`whoami`'",
		},
		{
			name:  "string with exclamation mark is quoted",
			input: "hello!",
			want:  "'hello!'",
		},
		{
			name:  "string with parentheses is quoted",
			input: "echo (test)",
			want:  "'echo (test)'",
		},
		{
			name:  "string with pipe is quoted",
			input: "cmd | other",
			want:  "'cmd | other'",
		},
		{
			name:  "string with backslash is quoted",
			input: `path\to\file`,
			want:  `'path\to\file'`,
		},
		{
			name:  "string with multiple single quotes",
			input: "it's a 'test'",
			want:  "'it'\"'\"'s a '\"'\"'test'\"'\"''",
		},
		{
			name:  "numeric string unchanged",
			input: "12345",
			want:  "12345",
		},
		{
			name:  "path-like string unchanged",
			input: "/usr/local/bin/app",
			want:  "/usr/local/bin/app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShellQuote(tt.input)
			if got != tt.want {
				t.Errorf("ShellQuote(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestShellStripUnsafe(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string returns empty",
			input: "",
			want:  "",
		},
		{
			name:  "printable ASCII unchanged",
			input: "hello world 123 !@#",
			want:  "hello world 123 !@#",
		},
		{
			name:  "strips null byte",
			input: "hello\x00world",
			want:  "helloworld",
		},
		{
			name:  "strips bell character",
			input: "hello\aworld",
			want:  "helloworld",
		},
		{
			name:  "strips backspace",
			input: "hello\bworld",
			want:  "helloworld",
		},
		{
			name:  "strips escape character",
			input: "hello\x1bworld",
			want:  "helloworld",
		},
		{
			name:  "strips ANSI escape sequence",
			input: "hello\x1b[31mred\x1b[0mworld",
			want:  "hello[31mred[0mworld",
		},
		{
			name:  "strips tab and newline",
			input: "hello\tworld\n",
			want:  "helloworld",
		},
		{
			name:  "strips vertical tab and form feed",
			input: "hello\vworld\f!",
			want:  "helloworld!",
		},
		{
			name:  "preserves unicode printable characters",
			input: "héllo wörld 日本語",
			want:  "héllo wörld 日本語",
		},
		{
			name:  "strips multiple control characters",
			input: "\x00\x01\x02hello\x03\x04",
			want:  "hello",
		},
		{
			name:  "all control characters stripped leaves empty",
			input: "\x00\x01\x02\x03",
			want:  "",
		},
		{
			name:  "strips carriage return and newline",
			input: "line1\r\nline2",
			want:  "line1line2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShellStripUnsafe(tt.input)
			if got != tt.want {
				t.Errorf("ShellStripUnsafe(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
