package sqllexer

import (
	"regexp"
	"strings"
)

type SQLObfuscatorConfig struct {
	ReplaceDigits bool
}

type SQLObfuscator struct {
	config *SQLObfuscatorConfig
}

func NewSQLObfuscator(config *SQLObfuscatorConfig) *SQLObfuscator {
	return &SQLObfuscator{config: config}
}

// Obfuscate takes an input SQL string and returns an obfuscated SQL string.
// The obfuscator replaces all literal values with a single placeholder
func (o *SQLObfuscator) Obfuscate(input string) string {
	var obfuscatedSQL string

	lexer := NewSQLLexer(input)
	for token := range lexer.ScanAllTokens() {
		switch token.Type {
		case NUMBER:
			obfuscatedSQL += "?"
		case STRING:
			obfuscatedSQL += "?"
		case INCOMPLETE_STRING:
			obfuscatedSQL += "?"
		case IDENT:
			if strings.EqualFold(token.Value, "null") || strings.EqualFold(token.Value, "true") || strings.EqualFold(token.Value, "false") {
				obfuscatedSQL += "?" // replace null, true, false with ?
			} else {
				if o.config.ReplaceDigits {
					// regex to replace digits in identifier
					// we try to avoid using regex as much as possible,
					// as regex isn't the most performant,
					// but it's the easiest to implement and maintain
					digits_regex := regexp.MustCompile(`\d+`)
					obfuscatedSQL += digits_regex.ReplaceAllString(token.Value, "?")
				} else {
					obfuscatedSQL += token.Value
				}
			}
		case COMMENT:
			// replace single line comment with multi line comment
			// this is done because the obfuscated SQL is collaped into a single line
			// and we don't want the single line comment to to mask the rest of the SQL
			commentContent := strings.TrimPrefix(token.Value, "--")
			obfuscatedSQL += "/*" + commentContent + " */"
		case MULTILINE_COMMENT:
			// replace newlines and tabs in multiline comment with whitespace
			obfuscatedSQL += collapseWhitespace(token.Value)
		case DOLLAR_QUOTED_STRING:
			obfuscatedSQL += "?"
		case DOLLAR_QUOTED_FUNCTION:
			// obfuscate the content of dollar quoted function
			quotedFunc := strings.TrimPrefix(token.Value, "$func$")
			quotedFunc = strings.TrimSuffix(quotedFunc, "$func$")
			obfuscatedSQL += "$func$" + o.Obfuscate(quotedFunc) + "$func$"
		case ERROR | UNKNOWN:
			// if we encounter an error or unknown token, we just append the value
			obfuscatedSQL += collapseWhitespace(token.Value)
		default:
			obfuscatedSQL += token.Value
		}
	}

	return strings.TrimSpace(obfuscatedSQL)
}
