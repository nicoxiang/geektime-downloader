package file

import (
	"math"
	"regexp"
	"strings"
)

// MaxFileNameLength ...
const MaxFileNameLength = 100

// Filenamify convert a string to a valid safe filename
func Filenamify(str string) string {
	// remove empty
	str = strings.Join(strings.Fields(str), "")

	replacement := "-"

	reControlCharsRegex := regexp.MustCompile("[\u0000-\u001f\u0080-\u009f]")

	reRelativePathRegex := regexp.MustCompile(`^\.+`)

	// https://github.com/sindresorhus/filename-reserved-regex/blob/master/index.js
	filenameReservedRegex := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	filenameReservedWindowsNamesRegex := regexp.MustCompile(`(?i)^(con|prn|aux|nul|com[0-9]|lpt[0-9])$`)

	// reserved word
	str = filenameReservedRegex.ReplaceAllString(str, replacement)

	// continue
	str = reControlCharsRegex.ReplaceAllString(str, replacement)
	str = reRelativePathRegex.ReplaceAllString(str, replacement)

	// for repeat
	if len(replacement) > 0 {
		str = trimRepeated(str, replacement)

		if len(str) > 1 {
			str = stripOuter(str, replacement)
		}
	}

	// for windows names
	if filenameReservedWindowsNamesRegex.MatchString(str) {
		str = str + replacement
	}

	// limit length
	strBuf := []byte(str)
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