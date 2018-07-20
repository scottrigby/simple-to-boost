package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/davecgh/go-spew/spew"
	"github.com/gofrs/uuid"
	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
)

const (
	// The other type is "SNIPPET_NOTE", however that requires a file name
	// and/or language "mode", which Simplenote exports don't have. If there's
	// a use-case for bulk importing Simplenotes that are all snippets and all
	// of the same language then we can make this configurable.
	noteType = "MARKDOWN_NOTE"
	// Note this template includes a newline purposefully, as this is expected
	// by Boostnote.
	tpl = `createdAt: "{{.Created}}"
updatedAt: "{{.Updated}}"
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
isTrashed: {{.Trashed}}
`
	// Mimick Boostnote expected time format. Similar to a cross between
	// time.RFC3339 and time.StampMilli.
	boostFormat = "2006-01-02T15:04:05.000Z"
	charset     = "abcdefghijklmnopqrstuvwxyz" + "0123456789"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func debug(thing interface{}) {
	spew.Dump(thing)
}

func main() {
	exportDir, err := getSimplenoteExportDir()
	check(err)

	boostStoragePath, err := getBoostStoragePath()
	check(err)

	folder, err := getBoostFolderID(boostStoragePath)
	check(err)

	fileInfos, err := ioutil.ReadDir(exportDir)
	check(err)

	fmap := sprig.TxtFuncMap()
	t := template.Must(template.New("test").Funcs(fmap).Parse(tpl))

	for _, fileInfo := range fileInfos {
		// No need to import empty files.
		if fileInfo.Size() == 0 {
			continue
		}

		title, _ := getTitle(fileInfo, exportDir)
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

		content, err := ioutil.ReadFile(filepath.Join(exportDir, fileInfo.Name()))
		check(err)

		// Ensure we escape cson triple-single quotes.
		content = bytes.Replace([]byte(content), []byte("'''"), []byte("\\'''"), -1)

		// Simplenote export prefixes trashed note file names with "trashed-".
		var trashed = "false"
		if strings.HasPrefix(fileInfo.Name(), "trash-") {
			trashed = "true"
		}

		vars := map[string]interface{}{
			"Created": updated,
			"Updated": updated,
			"Type":    noteType,
			"Folder":  folder,
			"Title":   title,
			"Content": string(content),
			"Trashed": trashed,
		}

		var buffer bytes.Buffer
		err = t.Execute(&buffer, vars)
		check(err)

		// Ensure note file names match Boostnote expectations.
		uuid, err := uuid.NewV4()
		check(err)
		newName := uuid.String() + ".cson"

		err = ioutil.WriteFile(filepath.Join(boostStoragePath, "notes", newName), buffer.Bytes(), 0644)
		check(err)
	}

	fmt.Println("Imported! Quit and reopen Boost to see your files.")
}

func getSimplenoteExportDir() (string, error) {
	validate := func(input string) error {
		input, err := homedir.Expand(input)
		if err != nil {
			return err
		}

		if _, err := os.Stat(input); os.IsNotExist(err) {
			return errors.New("Path does not exist")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Simplenote export directory",
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	result, err = homedir.Expand(result)
	if err != nil {
		return "", err
	}

	return result, err
}

// If there are only empty spaces and/or newlines, return an empty string.
func getTitle(fileInfo os.FileInfo, exportDir string) (string, error) {
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

// Support ~ expansion for home directory config.
func getBoostStoragePath() (string, error) {
	validate := func(input string) error {
		input, err := homedir.Expand(input)
		if err != nil {
			return err
		}
		if _, err := os.Stat(input); os.IsNotExist(err) {
			return errors.New("Path does not exist")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Boost storage directory",
		Validate: validate,
		Default:  "~/Boostnote",
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	boostStoragePath, err := homedir.Expand(result)
	if err != nil {
		return "", err
	}
	return boostStoragePath, nil
}

type boostConfig struct {
	Folders []folder `json:"folders"`
	Version string   `json:"version"`
}

type folder struct {
	Key   string `json:"key"`
	Color string `json:"color"`
	Name  string `json:"name"`
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

	// Build a list of folder names for the user prompt (they don't care about
	// the folder key).
	nameList := []string{}
	// Build a map keyed by name so we can get the folder "key" from the user's
	// selection.
	reverseKeyList := map[string]string{}
	for _, folder := range config.Folders {
		nameList = append(nameList, folder.Name)
		reverseKeyList[folder.Name] = folder.Key
	}
	// Add an additional option to automatically create a new Simplenote import
	// folder.
	// @todo Only prompt for a new folder if one by this name doesn't exist.
	nameList = append(nameList, "Create new folder")
	// Let people exit at this point without errors.
	nameList = append(nameList, "Exit")

	prompt := promptui.Select{
		Label: "Select folder",
		Items: nameList,
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	key := reverseKeyList[result]

	if result == "Create new folder" {
		// Add new folder to JSON config.
		key = randString(20)
		new := folder{
			Key: key,
			// See Boostnote folder color options at https://github.com/BoostIO/Boostnote/blob/master/browser/lib/consts.js
			Color: "#2BA5F7",
			Name:  "Simplenote import",
		}
		config.Folders = append(config.Folders, new)

		// Match Boostnote config JSON file 2 space indentation.
		newConfig, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return "", err
		}
		err = ioutil.WriteFile(filepath.Join(storagePath, "boostnote.json"), newConfig, 0644)
		if err != nil {
			return "", err
		}
	}

	if result == "Exit" {
		os.Exit(0)
		return "", nil
	}

	return key, nil
}

var seededRand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func randStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func randString(length int) string {
	return randStringWithCharset(length, charset)
}
