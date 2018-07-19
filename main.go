package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"text/template"

	"github.com/Masterminds/sprig"
)

const (
	// @todo Allow configurating export and import dirs.
	exportDir = "export"
	importDir = "import"
	tpl       = `createdAt: "{{.Created}}"
updatedAt: ""{{.Updated}}""
type: "MARKDOWN_NOTE"
folder: "{{.Folder}}"
title: "{{.Title}}"
content: '''
{{.Content | indent 2}}
'''
tags: []
isStarred: false
isTrashed: false
`
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func debug(thing interface{}) {
	fmt.Printf("%#v", thing)
}

func main() {
	fileInfos, err := ioutil.ReadDir(exportDir)
	check(err)

	fmap := sprig.TxtFuncMap()
	t := template.Must(template.New("test").Funcs(fmap).Parse(tpl))

	for _, fileInfo := range fileInfos {
		// No need to import empty files.
		if fileInfo.Size() == 0 {
			continue
		}

		title, _ := getTitle(fileInfo)
		check(err)
		// If the title is an empty string, it means the file only contains
		// empty lines or spaces. No need to import these either.
		if title == "" {
			continue
		}

		content, err := ioutil.ReadFile(filepath.Join(exportDir, fileInfo.Name()))
		check(err)

		// Ensure we escape cson triple-single quotes.
		content = bytes.Replace([]byte(content), []byte("'''"), []byte("\\'''"), -1)

		// @todo Dynamically get Created, Updated, Folder.
		vars := map[string]interface{}{
			"Created": "cxxx",
			"Updated": "uxxx",
			"Folder":  "fxxx",
			"Title":   title,
			"Content": string(content),
		}

		var buffer bytes.Buffer
		err = t.Execute(&buffer, vars)
		check(err)

		err = ioutil.WriteFile(filepath.Join(importDir, fileInfo.Name()), buffer.Bytes(), 0644)
		check(err)
	}

}

// If there are only empty spaces and/or newlines, return an empty string.
func getTitle(fileInfo os.FileInfo) (string, error) {
	file, err := os.Open(filepath.Join(exportDir, fileInfo.Name()))
	if err != nil {
		return "", err
	}
	// Default to an empty string, in case there's no content after trimming.
	title := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Strip leading or trailing spaces.
		trimmed := strings.TrimSpace(scanner.Text())
		// Continue through empty lines until we reach one with content.
		if trimmed == "" {
			continue
		} else {
			title = trimmed
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return title, nil
}
