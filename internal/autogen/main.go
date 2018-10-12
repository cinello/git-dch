package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

const (
	versionFilename = "../../../.version"
	hashFilename    = "../../../.version"
)

var (
	packageName   = flag.String("p", "main", "package name")
	gitCommand    = "git"
	gitVersionTag = []string{"tag", "--list", "[0-9999].[0-9999].[0-9999]", "--sort=-version:refname"}
	gitCommitHash = []string{"rev-parse", "--short", "HEAD"}
)

func valueFromGit(command string, parameters []string) (value string) {
	cmd := exec.Command(command, parameters...)
	if out, err := cmd.Output(); err == nil {
		value = strings.Split(strings.TrimSpace(string(out)), "\n")[0]
	}
	return
}

func valueFromFile(file string) (value string) {
	f, err := os.OpenFile(file, os.O_RDONLY, 0666)
	if err != nil {
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}

	value = string(data)
	return
}

func main() {
	flag.Parse()

	var version string
	version = valueFromGit(gitCommand, gitVersionTag)
	if version == "" {
		version = valueFromFile(versionFilename)
	}
	if version == "" {
		version = "0.0.0"
	}

	var hash string
	hash = valueFromGit(gitCommand, gitCommitHash)
	if hash == "" {
		hash = valueFromFile(hashFilename)
	}

	f, err := os.Create("autogendata.go")
	die(err)
	defer f.Close()

	packageTemplate.Execute(f, struct {
		Package   string
		Timestamp string
		Commit    string
		Version   string
	}{
		Package:   *packageName,
		Timestamp: time.Now().Format(time.UnixDate),
		Commit:    hash,
		Version:   version,
	})
}

func die(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var packageTemplate = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots at
// {{ .Timestamp }}

package {{ .Package }}

import "time"

var (
	buildDate, _  = time.Parse(time.UnixDate, "{{ .Timestamp }}")
	commitHash = "{{ .Commit }}"
	version    = "{{ .Version }}"
)
`))
