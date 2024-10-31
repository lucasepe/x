package env

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
)

func Load(in io.Reader) (err error) {
	scanner := bufio.NewScanner(in)

	for scanner.Scan() {
		key, val := parseLine(scanner.Text())
		if len(key) == 0 || len(val) == 0 {
			continue
		}

		err2 := os.Setenv(key, val)
		if err2 != nil {
			err = errors.Join(err, err2)
			continue
		}
	}

	err = errors.Join(err, scanner.Err())

	return
}

func parseLine(s string) (key, val string) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return
	}

	if s[0] == '#' {
		return
	}

	for j := 0; j < len(s); j++ {
		if s[j] == '=' {
			key = s[:j]
			val = s[j+1:]
			return
		}
	}

	return
}
