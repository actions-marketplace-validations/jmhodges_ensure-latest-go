package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/jmhodges/yaml.v2"
)

func updateTravisFiles(travisfilePaths map[string]bool, goVers string) ([]fileContent, error) {
	var files []fileContent
	for fp, _ := range travisfilePaths {
		// O_RDWR so we can ensure we can write to the file without doing a
		// bunch of work first
		f, err := os.OpenFile(fp, os.O_RDWR, 0644)
		if err != nil {
			return nil, fmt.Errorf("unable to open Dockerfile %#v for reading: %w", fp, err)
		}
		defer f.Close()
		origFileContents, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, fmt.Errorf("unable to read contents of Dockerfile %#v: %s", fp, err)
		}

		contentsToWrite, err := updateSingleTravisFile(fp, origFileContents, goVers)
		if err != nil {
			return nil, fmt.Errorf("unable to parse YAML travis config file %#v: %s", fp, err)
		}
		if contentsToWrite != nil {
			files = append(files, fileContent{origFP: fp, contentsToWrite: contentsToWrite})
		}
	}
	return files, nil
}

func updateSingleTravisFile(fp string, origFileContents []byte, goVers string) ([]byte, error) {
	var ty yaml.MapSlice
	err := yaml.Unmarshal(origFileContents, &ty)
	if err != nil {
		return nil, err
	}
	log.Println("FIXME updateSingleTravisFile 1", ty)
	i, goVersions, err := findMapItem(ty, "go")
	if err != nil {
		return nil, err
	}
	if i == -1 {
		return origFileContents, nil
	}
	log.Println("FIXME updateSingleTravisFile 110", i, goVersions, err)
	var fileContentsUpdated bool
	switch oldGoVers := goVersions.(type) {
	case string:
		if oldGoVers != goVers {
			ty[i].Value = goVers
			fileContentsUpdated = true
		}
	case []interface{}:
		versions := make(map[string]bool)
		var out []string
		log.Println("FIXME updateSingleTravisFile 30", oldGoVers, goVers)

		for _, oldVersInt := range oldGoVers {
			oldVers, ok := oldVersInt.(string)
			if !ok {
				return nil, fmt.Errorf("unknown type in 'go' array in travis config file %#v: %s", fp, err)
			}
			if !versions[oldVers] {
				out = append(out, oldVers)
				versions[oldVers] = true
			}
		}
		if !versions[goVers] {
			fileContentsUpdated = true
			if len(versions) == 2 {
				ty[i].Value = yaml.MapItem{Key: "go", Value: goVers}
			} else {
				ty[i].Value = append(out, goVers)
			}
		}
	default:
		return nil, fmt.Errorf("unknown type for 'go' value in travis config file %#v: %s", fp, err)
	}
	if fileContentsUpdated {
		return yamlMarshal(ty)
	}
	return origFileContents, nil
}
