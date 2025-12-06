package slack

import (
	"html"
	"regexp"
	"strings"
)

// convertSlackToMarkdown converts Slack's mrkdwn format to standard Markdown
func convertSlackToMarkdown(text string) string {
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
	// Use negative lookbehind/lookahead to avoid matching * in middle of words or already converted **
	// Since Go regex doesn't support lookbehind, we'll use a simpler approach
	// Match *word* or *multiple words* but not already doubled **
	boldRe := regexp.MustCompile(`(?:^|[\s\n])(\*)([^\*\n]+)\*`)
	text = boldRe.ReplaceAllString(text, "${0:0}**$2**")
	// Clean up the start marker we preserved
	text = strings.ReplaceAll(text, "${0:0}", "")

	// Actually, let's use a simpler and more reliable approach:
	// Split on **, process each segment for *, then rejoin
	// This prevents converting * to ** if it's already part of **

	// Let me rewrite this more simply:
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

	// 5. Strikethrough ~text~ → ~~text~~ (same format in both)
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
