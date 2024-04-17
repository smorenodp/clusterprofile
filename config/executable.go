package config

import (
	"os"
	"os/exec"
	"strings"
	"text/template"
)

type ExecFile struct {
	Profile string
	Creds   []string
	Banner  Banner
}

type Banner struct {
	Enable  bool
	Command string
	Args    string
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

const (
	templateFile string = `
#!/bin/bash
{{ if .Banner.Enable }}
{{ .Banner.Command }} {{ .Banner.Args }} {{ .Profile }} && echo
{{ end }}
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

func CreateExecFile(execFile string, profile string, creds []string, banner Banner) {
	createDirectory(execFile)
	f, _ := os.OpenFile(execFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	tpl, _ := template.New("exec").Parse(templateFile)
	exec := ExecFile{Profile: profile, Creds: creds}
	if banner.Enable && commandExists(banner.Command) {
		exec.Banner = banner
	}
	tpl.Execute(f, exec)
}

func GenerateExportContent(profile string, creds []string, banner Banner) {
	tpl, _ := template.New("exec").Parse(templateFile)
	exec := ExecFile{Profile: profile, Creds: creds}
	if banner.Enable && commandExists(banner.Command) {
		exec.Banner = banner
	}
	tpl.Execute(os.Stdout, exec)
}
