package filenamify

import (
	"math"
	"regexp"
	"strings"
)

const (
	// MaxFileNameLength ...
	MaxFileNameLength = 100
	// Replacement for special chars
	Replacement = "-"
)

// Filenamify convert a string to a valid safe filename
func Filenamify(str string) string {
	// remove empty
	str = strings.Join(strings.Fields(str), "")

	reControlCharsRegex := regexp.MustCompile("[\u0000-\u001f\u0080-\u009f]")

	reRelativePathRegex := regexp.MustCompile(`^\.+`)

	// https://stackoverflow.com/a/31976060/5685258
	forbiddenWindowsCharsRegex := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	reservedWindowsNamesRegex := regexp.MustCompile(`(?i)^(con|prn|aux|nul|com[0-9]|lpt[0-9])$`)

	// reserved word
	str = forbiddenWindowsCharsRegex.ReplaceAllString(str, Replacement)

	// continue
	str = reControlCharsRegex.ReplaceAllString(str, Replacement)
	str = reRelativePathRegex.ReplaceAllString(str, Replacement)

	// for repeat
	if len(Replacement) > 0 {
		str = trimRepeated(str, Replacement)

		if len(str) > 1 {
			str = stripOuter(str, Replacement)
		}
	}

	// for windows names
	if reservedWindowsNamesRegex.MatchString(str) {
		str = str + Replacement
	}

	// limit length
	strBuf := []rune(str)
	strBuf = strBuf[0:int(math.Min(float64(MaxFileNameLength), float64(len(strBuf))))]

	return string(strBuf)
}

func escapeStringRegexp(str string) string {
	// https://github.com/sindresorhus/escape-string-regexp/blob/master/index.js
	reg := regexp.MustCompile(`[|\\{}()[\]^$+*?.-]`)
	str = reg.ReplaceAllStringFunc(str, func(s string) string {
		return `\` + s
	})
	return str
}

func trimRepeated(str string, replacement string) string {
	reg := regexp.MustCompile(`(?:` + escapeStringRegexp(replacement) + `){2,}`)
	return reg.ReplaceAllString(str, replacement)
}

func stripOuter(input string, substring string) string {
	// https://github.com/sindresorhus/strip-outer/blob/master/index.js
	substring = escapeStringRegexp(substring)
	reg := regexp.MustCompile(`^` + substring + `|` + substring + `$`)
	return reg.ReplaceAllString(input, "")
}
