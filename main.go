package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"text/template"

	"github.com/Masterminds/sprig"
	homedir "github.com/mitchellh/go-homedir"
)

const (
	// @todo Allow users to specify Boost storage path. I'm not sure this is the
	// same on every OS installation.
	boostDir = "~/Boostnote"
	// @todo Allow configurating export and import dirs.
	exportDir = "export"
	importDir = "import"
	// The other type is "SNIPPET_NOTE", however that requires a file name
	// and/or language "mode", which Simplenote exports don't have. If there's
	// a use-case for bulk importing Simplenotes that are all snippets and all
	// of the same language then we can make this configurable.
	noteType = "MARKDOWN_NOTE"
	// Note this template includes a newline purposefully, as this is expected
	// by Boostnote.
	tpl = `createdAt: "{{.Created}}"
updatedAt: ""{{.Updated}}""
type: "{{.Type}}"
folder: "{{.Folder}}"
title: "{{.Title}}"
content: '''
{{.Content | indent 2}}
'''
tags: [
  "simplenote-import"
]
isStarred: false
isTrashed: false
`
	// Mimick Boostnote expected time format. Similar to a cross between
	// time.RFC3339 and time.StampMilli.
	boostFormat = "2006-01-02T15:04:05.000Z"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func debug(thing interface{}) {
	fmt.Printf("%#v\n", thing)
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

		// Note creation time is not stored in most Linux file systems.
		// We could hack something together for Windows, MacOS, Ext4 etc, but
		// Go standard library does not support it. So let's just use updated
		// time as created time as well.
		modTime := fileInfo.ModTime()
		updated := modTime.Format(boostFormat)

		boostStoragePath, err := getBoostStoragePath()
		check(err)
		folder, err := getBoostFolderID(boostStoragePath)
		check(err)

		content, err := ioutil.ReadFile(filepath.Join(exportDir, fileInfo.Name()))
		check(err)

		// Ensure we escape cson triple-single quotes.
		content = bytes.Replace([]byte(content), []byte("'''"), []byte("\\'''"), -1)

		vars := map[string]interface{}{
			"Created": updated,
			"Updated": updated,
			"Type":    noteType,
			"Folder":  folder,
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

// We only care about the key.
type boostConfig struct {
	Folders []struct {
		Key string `json:"key"`
	} `json:"folders"`
}

// Support ~ expansion for home directory config.
func getBoostStoragePath() (string, error) {
	boostStoragePath, err := homedir.Expand(boostDir)
	if err != nil {
		return "", err
	}
	return boostStoragePath, nil
}

func getBoostFolderID(storagePath string) (string, error) {
	file, err := ioutil.ReadFile(filepath.Join(storagePath, "boostnote.json"))
	if err != nil {
		return "", err
	}

	var config boostConfig
	err = json.Unmarshal(file, &config)
	if err != nil {
		return "", err
	}

	// @todo Allow users to specify which folder they wish to import into.
	// For now, just use the key of the first folder listed in the config file.
	folder := config.Folders[0].Key

	return folder, nil
}
