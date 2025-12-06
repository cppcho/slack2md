package slack

import "testing"

func TestConvertSlackToMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Bold conversions
		{
			name:     "bold single word",
			input:    "*hello*",
			expected: "**hello**",
		},
		{
			name:     "bold multiple words",
			input:    "*hello world*",
			expected: "**hello world**",
		},
		{
			name:     "bold with punctuation",
			input:    "*hello, world!*",
			expected: "**hello, world!**",
		},
		{
			name:     "multiple bold segments",
			input:    "*first* and *second*",
			expected: "**first** and**second**",
		},

		// Italic conversions
		{
			name:     "italic single word",
			input:    "_hello_",
			expected: "*hello*",
		},
		{
			name:     "italic multiple words",
			input:    "_hello world_",
			expected: "*hello world*",
		},
		{
			name:     "multiple italic segments",
			input:    "_first_ and _second_",
			expected: "*first* and *second*",
		},

		// Strikethrough conversions
		{
			name:     "strikethrough single word",
			input:    "~deleted~",
			expected: "~~deleted~~",
		},
		{
			name:     "strikethrough multiple words",
			input:    "~not needed anymore~",
			expected: "~~not needed anymore~~",
		},

		// Combined formatting
		{
			name:     "bold and italic",
			input:    "*bold* and _italic_",
			expected: "**bold** and *italic*",
		},
		{
			name:     "bold italic and strikethrough",
			input:    "*bold* _italic_ ~strike~",
			expected: "**bold** *italic* ~~strike~~",
		},
		{
			name:     "nested-like formatting",
			input:    "*bold with _italic_ inside*",
			expected: "**bold with *italic* inside**",
		},

		// Link conversions
		{
			name:     "link with text",
			input:    "<https://example.com|Example Site>",
			expected: "[Example Site](https://example.com)",
		},
		{
			name:     "bare URL",
			input:    "<https://example.com>",
			expected: "https://example.com",
		},
		{
			name:     "http bare URL",
			input:    "<http://example.com>",
			expected: "http://example.com",
		},
		{
			name:     "multiple links",
			input:    "Visit <https://example.com|Example> or <https://test.com|Test>",
			expected: "Visit [Example](https://example.com) or [Test](https://test.com)",
		},
		{
			name:     "link with bold text",
			input:    "<https://example.com|*Example*>",
			expected: "[**Example**](https://example.com)",
		},

		// HTML entity conversions
		{
			name:     "ampersand entity",
			input:    "Tom &amp; Jerry",
			expected: "Tom & Jerry",
		},
		{
			name:     "less than entity",
			input:    "x &lt; y",
			expected: "x < y",
		},
		{
			name:     "greater than entity",
			input:    "x &gt; y",
			expected: "x > y",
		},
		{
			name:     "multiple entities",
			input:    "&lt;div&gt; &amp; &lt;/div&gt;",
			expected: "<div> & </div>",
		},
		{
			name:     "quote entity",
			input:    "&quot;hello&quot;",
			expected: "\"hello\"",
		},

		// Edge cases
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "plain text no formatting",
			input:    "just plain text",
			expected: "just plain text",
		},
		{
			name:     "whitespace only",
			input:    "   \t\n  ",
			expected: "   \t\n  ",
		},
		{
			name:     "single asterisk",
			input:    "text * text",
			expected: "text * text",
		},
		{
			name:     "single underscores in text",
			input:    "text with underscores",
			expected: "text with underscores",
		},
		{
			name:     "single tilde",
			input:    "path/to/~user",
			expected: "path/to/~user",
		},
		{
			name:     "unmatched asterisk",
			input:    "*unmatched",
			expected: "*unmatched",
		},
		{
			name:     "unmatched underscore",
			input:    "_unmatched",
			expected: "_unmatched",
		},
		{
			name:     "already markdown bold",
			input:    "**already bold**",
			expected: "**already bold**",
		},
		{
			name:     "code block with backticks",
			input:    "`code here`",
			expected: "`code here`",
		},
		{
			name:     "multiline code block",
			input:    "```\ncode block\n```",
			expected: "```\ncode block\n```",
		},
		{
			name:     "blockquote",
			input:    "> quoted text",
			expected: "> quoted text",
		},

		// Complex real-world examples
		{
			name:     "complex message with multiple formats",
			input:    "*Hey team!* Check out <https://example.com|this link> and review _the docs_. ~Old info~ removed.",
			expected: "**Hey team!** Check out [this link](https://example.com) and review *the docs*. ~~Old info~~ removed.",
		},
		{
			name:     "message with entities and formatting",
			input:    "*Important:* x &gt; 0 &amp; y &lt; 100",
			expected: "**Important:** x > 0 & y < 100",
		},
		{
			name:     "multiple lines with formatting",
			input:    "*Line 1*\n_Line 2_\n~Line 3~",
			expected: "**Line 1**\n*Line 2*\n~~Line 3~~",
		},

		// Edge cases with special characters
		{
			name:     "asterisks in middle of word",
			input:    "ex*am*ple",
			expected: "ex**am**ple",
		},
		{
			name:     "formatting with numbers",
			input:    "*123* _456_ ~789~",
			expected: "**123** *456* ~~789~~",
		},
		{
			name:     "formatting with special chars",
			input:    "*hello!* _world?_ ~test..~",
			expected: "**hello!** *world?* ~~test..~~",
		},

		// User mentions and channel references (passed through as-is)
		{
			name:     "user mention",
			input:    "<@U1234567> mentioned you",
			expected: "<@U1234567> mentioned you",
		},
		{
			name:     "channel reference",
			input:    "Posted in <#C1234567>",
			expected: "Posted in <#C1234567>",
		},
		{
			name: "bulleted list with formatting",
			input: `• aaa
    ◦ bbb
        ▪︎ ccc
            • ddd
                ◦ eee
                ◦ fff`,
			expected: `* aaa
    * bbb
        * ccc
            * ddd
                * eee
                * fff`,
		},
		{
			name:     "do not replace bullets if it is not a list",
			input:    `zzz • aaa ◦ bbb ▪︎ ccc`,
			expected: `zzz • aaa ◦ bbb ▪︎ ccc`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertSlackToMarkdown(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertSlackToMarkdown(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestConvertSlackToMarkdown_MarkdownPreservation verifies behavior with existing markdown
func TestConvertSlackToMarkdown_MarkdownPreservation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already markdown bold preserved",
			input:    "**bold text**",
			expected: "**bold text**",
		},
		{
			name:     "already markdown link preserved",
			input:    "[link text](https://example.com)",
			expected: "[link text](https://example.com)",
		},
		// Note: The formatter is designed to convert Slack format to Markdown,
		// so it will convert markdown-style formatting in the input as well.
		// This is expected behavior since Slack messages use different syntax.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertSlackToMarkdown(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertSlackToMarkdown(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
