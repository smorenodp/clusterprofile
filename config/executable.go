package config

import (
	"os"
	"strings"
	"text/template"
)

type ExecFile struct {
	Profile string
	Creds   []string
}

const (
	templateFile string = `
#!/bin/bash
echo "Loading {{ .Profile }} credentials"
{{ range .Creds }}
{{ . }}
{{ end }}
`
)

func createDirectory(file string) {
	lastInd := strings.LastIndex(file, "/")
	if lastInd != -1 {
		os.MkdirAll(file[:lastInd], os.ModePerm)
	}
}

func CreateExecFile(execFile string, profile string, creds []string) {
	createDirectory(execFile)
	f, _ := os.OpenFile(execFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	tpl, _ := template.New("exec").Parse(templateFile)
	tpl.Execute(f, ExecFile{profile, creds})
}

func GenerateExportContent(profile string, creds []string) {
	tpl, _ := template.New("exec").Parse(templateFile)
	tpl.Execute(os.Stdout, ExecFile{profile, creds})
}
