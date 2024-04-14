package config

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

const (
	nameRegex = "\\[(?P<name>.*)\\]"
)

type CredConfig map[string][]string

func LoadCreds(file string) (creds CredConfig, err error) {
	var name string
	var credLines []string
	var f *os.File

	creds = make(CredConfig)
	reg := regexp.MustCompile(nameRegex)
	if _, err = os.Stat(file); err != nil {
		os.Create(file)
		return creds, nil
	}
	f, err = os.Open(file)
	if err != nil {
		return
	}

	r := bufio.NewReader(f)

	for line, _, _ := r.ReadLine(); line != nil; line, _, _ = r.ReadLine() {
		stringLine := string(line)
		if matches := reg.FindStringSubmatch(stringLine); matches != nil {
			if len(credLines) > 0 {
				creds[name] = credLines
				credLines = []string{}
			}
			name = matches[1] // The name is always this index
		} else if stringLine != "" {
			credLines = append(credLines, stringLine)
		}
	}
	if len(credLines) > 0 {
		creds[name] = credLines
	}
	return creds, err
}

func SaveCreds(file string, creds CredConfig) error {
	f, _ := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	for name, lines := range creds {
		f.WriteString(fmt.Sprintf("[%s]\n", name))
		for _, line := range lines {
			f.WriteString(fmt.Sprintf("%s\n", line))
		}
	}
	return nil
}
