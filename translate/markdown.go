package translate

import (
	"context"
	"fmt"
	"strings"
)

func TranslateMarkdown(ctx context.Context, tr *Translator, input string, from, to string) (string, error) {
	lines := strings.Split(input, "\n")
	inCodeBlock := false
	var result strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		switch {
		case inCodeBlock:
			result.WriteString(line)
			result.WriteByte('\n')
			if isFenceLine(trimmed) {
				inCodeBlock = false
			}

		case isFenceLine(trimmed):
			result.WriteString(line)
			result.WriteByte('\n')
			inCodeBlock = true

		case trimmed == "":
			result.WriteByte('\n')

		case isThematicBreak(trimmed):
			result.WriteString(line)
			result.WriteByte('\n')

		default:
			translated, err := translateContent(ctx, tr, line, from, to)
			if err != nil {
				return "", err
			}
			result.WriteString(translated)
			result.WriteByte('\n')
		}
	}

	return strings.TrimRight(result.String(), "\n"), nil
}

func isFenceLine(s string) bool {
	return strings.HasPrefix(s, "```") || strings.HasPrefix(s, "~~~")
}

func isThematicBreak(s string) bool {
	if len(s) < 3 {
		return false
	}
	for _, r := range s {
		if r != '*' && r != '-' && r != '_' && r != ' ' {
			return false
		}
	}
	return true
}

func translateContent(ctx context.Context, tr *Translator, line, from, to string) (string, error) {
	prefix, body := splitPrefix(line)
	if body == "" {
		return line, nil
	}

	tokens := protectLinks(&body)
	if tokens != nil && onlyPlaceholders(body, len(tokens)) {
		restoreLinks(&body, tokens)
		return prefix + body, nil
	}

	translated, err := tr.Translate(ctx, body, from, to)
	if err != nil {
		return "", err
	}
	restoreLinks(&translated, tokens)

	return prefix + translated, nil
}

func onlyPlaceholders(s string, n int) bool {
	var want strings.Builder
	for i := 0; i < n; i++ {
		want.WriteString(fmt.Sprintf("0xMD%04x", i))
	}
	return s == want.String()
}

func protectLinks(s *string) []string {
	var tokens []string
	runes := []rune(*s)
	var b strings.Builder
	i := 0
	for i < len(runes) {
		if runes[i] == '[' && (i == 0 || runes[i-1] != '!') {
			start := i
			i++
			// find matching ] at nesting level 0 (skip over image syntax like [![alt](url)])
			depth := 0
			for i < len(runes) {
				if runes[i] == '[' {
					depth++
				} else if runes[i] == ']' {
					if depth == 0 {
						break
					}
					depth--
				}
				i++
			}
			if i < len(runes) && i+1 < len(runes) && runes[i+1] == '(' {
				i++ // skip ]
				i++ // skip (
				for i < len(runes) && runes[i] != ')' {
					i++
				}
				if i < len(runes) {
					i++ // skip )
				}
				tokens = append(tokens, string(runes[start:i]))
				b.WriteString(fmt.Sprintf("0xMD%04x", len(tokens)-1))
				continue
			}
		}
		b.WriteRune(runes[i])
		i++
	}
	*s = b.String()
	return tokens
}

func restoreLinks(s *string, tokens []string) {
	for i, t := range tokens {
		*s = strings.ReplaceAll(*s, fmt.Sprintf("0xMD%04x", i), t)
	}
}

func splitWhitespace(line string) (ws, rest string) {
	idx := 0
	for idx < len(line) && (line[idx] == ' ' || line[idx] == '\t') {
		idx++
	}
	return line[:idx], line[idx:]
}

func splitAtxHeading(ws, rest string) (prefix, body string, ok bool) {
	if rest[0] != '#' {
		return "", "", false
	}
	i := 1
	for i < len(rest) && rest[i] == '#' {
		i++
	}
	if i < len(rest) && rest[i] == ' ' {
		return ws + rest[:i+1], rest[i+1:], true
	}
	return ws, rest, true
}

func splitBlockquote(ws, rest string) (prefix, body string, ok bool) {
	if rest[0] != '>' {
		return "", "", false
	}
	if len(rest) > 1 && rest[1] == ' ' {
		return ws + "> ", rest[2:], true
	}
	return ws + ">", rest[1:], true
}

func splitUnorderedList(ws, rest string) (prefix, body string, ok bool) {
	if rest[0] != '-' && rest[0] != '*' && rest[0] != '+' {
		return "", "", false
	}
	if len(rest) > 1 && rest[1] == ' ' {
		return ws + rest[:2], rest[2:], true
	}
	return ws, rest, true
}

func splitOrderedList(ws, rest string) (prefix, body string, ok bool) {
	if rest[0] < '0' || rest[0] > '9' {
		return "", "", false
	}
	i := 0
	for i < len(rest) && rest[i] >= '0' && rest[i] <= '9' {
		i++
	}
	if i < len(rest) && rest[i] == '.' && i+1 < len(rest) && rest[i+1] == ' ' {
		return ws + rest[:i+2], rest[i+2:], true
	}
	return "", "", false
}

func splitPrefix(line string) (prefix, body string) {
	ws, rest := splitWhitespace(line)

	if rest == "" {
		return line, ""
	}

	if p, b, ok := splitAtxHeading(ws, rest); ok {
		return p, b
	}
	if p, b, ok := splitBlockquote(ws, rest); ok {
		return p, b
	}
	if p, b, ok := splitUnorderedList(ws, rest); ok {
		return p, b
	}
	if p, b, ok := splitOrderedList(ws, rest); ok {
		return p, b
	}
	return ws, rest
}
