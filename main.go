package main

import (
	"audializer/def"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	flags := flag.NewFlagSet("flags", flag.ExitOnError)
	configPath := flags.String("config-file", "config.yaml", "Path to configuration file")
	flags.Parse(os.Args[1:])
	configFile, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Cannot open config file: %v", err)
	}
	var config def.Config
	yaml.Unmarshal(configFile, config)

	var dirCreateFunc fs.WalkDirFunc = func(path string, entry fs.DirEntry, err error) error {
		if !entry.IsDir() {
			return nil
		}
		split := filepath.SplitList(path)
		if len(split) < 2 {
			return nil
		}
		newPathParts := append([]string{config.OutputFolder}, split[1:]...)
		newPath := filepath.Join(newPathParts...)
		os.MkdirAll(newPath, 0775)
		return nil
	}

	var walkFunc fs.WalkDirFunc = func(path string, entry fs.DirEntry, err error) error {
		if entry.IsDir() {
			return nil
		}

		filename := entry.Name()
		ext := filepath.Ext(filename)
		name, _ := strings.CutSuffix(filename, "."+ext)

		if !def.ALLOWED_MIMETYPES[ext] {
			return nil
		}

		outputFilename := fmt.Sprintf("%s.mp3", name)
		outputFile := filepath.Join(config.OutputFolder, outputFilename)

		// check if file already exists by seeing if a stat succeeds
		if _, err := os.Stat(outputFile); err == nil {
			return nil
		}

		args := []string{
			"-i",
			filename,
			"-q:a", "0",
			"-map", "a",
			outputFile,
		}

		exec.Command("ffmpeg", args...)

		return nil
	}

	for _, folder := range config.Folders {
		err := filepath.WalkDir(folder, dirCreateFunc)
		if err != nil {
			log.Fatalf("Error while creating directories: %v", err)
		}
		err = filepath.WalkDir(folder, walkFunc)
		if err != nil {
			log.Fatalf("Error while transcoding files: %v", err)
		}
	}
}
