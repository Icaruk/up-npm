package npmrc

import (
	"errors"
	"regexp"
	"strings"
)

func ParseNpmrc(str string) (token string, err error) {

	/*
		See https://docs.npmjs.com/cli/v10/configuring-npm/npmrc#auth-related-configuration

		Example:
			# comment
			; comment
			//registry.npmjs.org/:_authToken=npm_ABCdef123456
	*/

	// Read "str" line by line
	lines := strings.Split(str, "\n")

	for _, line := range lines {

		cleanLine := strings.TrimSpace(line)

		// Skip empty lines
		if cleanLine == "" {
			continue
		}

		// Skip comments
		if strings.HasPrefix(cleanLine, "#") || strings.HasPrefix(cleanLine, ";") {
			continue
		}

		// Check if starts with "//"
		if strings.HasPrefix(cleanLine, "//") {

			re := regexp.MustCompile(`^\/\/registry.npmjs.org/:_authToken=(.*)$`)

			matches := re.FindStringSubmatch(cleanLine)
			if matches == nil {
				continue
			}

			return matches[1], nil

		}

	}

	return "", errors.New("no token found")

}
