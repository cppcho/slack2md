package formatter

import (
	"html"
	"regexp"
	"strings"
)

// MarkdownFormatter provides Slack to Markdown conversion
type MarkdownFormatter struct{}

// NewMarkdownFormatter creates a new MarkdownFormatter
func NewMarkdownFormatter() *MarkdownFormatter {
	return &MarkdownFormatter{}
}

// ConvertSlackToMarkdown converts Slack's mrkdwn format to standard Markdown
func (f *MarkdownFormatter) ConvertSlackToMarkdown(text string) string {
	// First, decode HTML entities (&gt; → >, &lt; → <, &amp; → &, etc.)
	text = html.UnescapeString(text)

	// Process in order to avoid conflicts between replacements

	// 1. Convert links with text: <url|text> → [text](url)
	linkWithTextRe := regexp.MustCompile(`<([^|>]+)\|([^>]+)>`)
	text = linkWithTextRe.ReplaceAllString(text, "[$2]($1)")

	// 2. Convert bare URLs: <url> → url (remove angle brackets)
	bareURLRe := regexp.MustCompile(`<(https?://[^>]+)>`)
	text = bareURLRe.ReplaceAllString(text, "$1")

	// 3. Convert bold: *text* → **text**
	// First, protect existing ** by replacing with a placeholder
	text = strings.ReplaceAll(text, "**", "\x00BOLD\x00")

	// Now convert * to **
	boldSimpleRe := regexp.MustCompile(`\*([^\*]+?)\*`)
	text = boldSimpleRe.ReplaceAllString(text, "**$1**")

	// Restore protected **
	text = strings.ReplaceAll(text, "\x00BOLD\x00", "**")

	// 4. Convert italic: _text_ → *text*
	// Similarly, protect existing * that we just created
	text = strings.ReplaceAll(text, "*", "\x00STAR\x00")

	italicRe := regexp.MustCompile(`_([^_]+?)_`)
	text = italicRe.ReplaceAllString(text, "*$1*")

	// Restore protected *
	text = strings.ReplaceAll(text, "\x00STAR\x00", "*")

	// 5. Strikethrough ~text~ → ~~text~~
	strikeRe := regexp.MustCompile(`~([^~]+?)~`)
	text = strikeRe.ReplaceAllString(text, "~~$1~~")

	// 6. Replace bullet characters with markdown bullets
	// Handle variation selectors (U+FE0E, U+FE0F) that may follow bullet characters
	bulletRe := regexp.MustCompile(`(?m)^(\s*)(\x{2022}|\x{25E6}|\x{25AA})[\x{FE0E}\x{FE0F}]?`)
	text = bulletRe.ReplaceAllString(text, "$1*")

	// 7. Code blocks and inline code are the same in both formats (`)
	// 8. Quotes (>) are the same in both formats
	// No conversion needed for these

	return text
}
