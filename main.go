package main

import (
	"bytes"
	"io/ioutil"
	"path/filepath"

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

func main() {
	files, err := ioutil.ReadDir(exportDir)
	check(err)

	fmap := sprig.TxtFuncMap()
	t := template.Must(template.New("test").Funcs(fmap).Parse(tpl))

	for _, file := range files {
		content, err := ioutil.ReadFile(filepath.Join(exportDir, file.Name()))
		check(err)

		// Ensure we escape cson triple-single quotes.
		content = bytes.Replace([]byte(content), []byte("'''"), []byte("\\'''"), -1)

		// @todo Dynamically get Created, Updated, Folder, & Title.
		vars := map[string]interface{}{
			"Created": "cxxx",
			"Updated": "uxxx",
			"Folder":  "fxxx",
			"Title":   "txxx",
			"Content": string(content),
		}

		var buffer bytes.Buffer
		err = t.Execute(&buffer, vars)
		check(err)

		err = ioutil.WriteFile(filepath.Join(importDir, file.Name()), buffer.Bytes(), 0644)
		check(err)
	}

}
