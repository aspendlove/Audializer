package main

import (
	"audializer/def"
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/vansante/go-ffprobe.v2"
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
	if err := yaml.Unmarshal(configFile, &config); err != nil {
		log.Fatalf("Cannot read config file: %v", err)
	}

	for _, folder := range config.Folders {
		destRoot := folder.Destination

		err := filepath.WalkDir(folder.Source, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(folder.Source, path)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Join(destRoot, rel), 0775); err != nil {
				return fmt.Errorf("Cannot create new path: %v", err)
			}
			return nil
		})
		if err != nil {
			log.Fatalf("Error while creating directories: %v", err)
		}

		err = filepath.WalkDir(folder.Source, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}

			filename := entry.Name()
			ext := filepath.Ext(filename)
			if !def.ALLOWED_EXTENSIONS[strings.ToLower(ext)] {
				return nil
			}

			rel, err := filepath.Rel(folder.Source, path)
			if err != nil {
				return err
			}
			outputFile := filepath.Join(destRoot, strings.TrimSuffix(rel, ext)+".mp3")

			// check if file already exists by seeing if a stat succeeds
			if _, err := os.Stat(outputFile); err == nil {
				return nil
			}

			stream, err := getAudioStream(path)
			if err != nil {
				log.Fatalf("Cannot get audio stream from file: %v", err)
			}

			args := []string{
				// "-hide_banner",
				// "-loglevel", "error",
				"-i", path,
				"-vn",
				"-map", fmt.Sprintf("0:%d", stream),
				"-c:a", "libmp3lame",
				"-q:a", "0",
				"-f", "mp3",
				outputFile,
			}

			cmd := exec.Command("ffmpeg", args...)

			cmd.Stdout = os.Stderr
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				if err := os.Remove(outputFile); err != nil {
					log.Default().Printf("Cannot delete failed file %s\n", outputFile)
				}
				return fmt.Errorf("Transcode for %s failed: %v", path, err)
			}

			return nil
		})
		if err != nil {
			log.Fatalf("Error while transcoding files: %v", err)
		}
	}
}

func getAudioStream(path string) (int, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	fileReader, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("Error opening test file: %v", err)
	}
	defer fileReader.Close()

	data, err := ffprobe.ProbeReader(ctx, fileReader)
	if err != nil {
		return 0, fmt.Errorf("Error getting data: %v", err)
	}
	audio := data.StreamType(ffprobe.StreamAudio)
	if len(audio) == 0 {
		return 0, fmt.Errorf("no audio streams in %s", path)
	}

	pick := audio[0]
	for _, s := range audio {
		if s.TagList["Language"] == "eng" {
			pick = s
			break
		}
	}
	return pick.Index, nil
}
