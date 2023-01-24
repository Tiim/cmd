package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/gabriel-vasile/mimetype"
)

type StorjConfig struct {
	// the access grant from the storj.io dashboard
	AccessGrant string `toml:"access_grant"`
	// the name of the storage bucket
	BucketName string `toml:"bucket_name"`
	// the Prefix to use for the uploaded files
	Prefix string `toml:"bucket_prefix"`
}

type WebPConfig struct {
	Enabled   bool     `toml:"enabled"`
	MimeTypes []string `toml:"mime_types"`
	Quality   int      `toml:"quality"`
}

type Config struct {
	StorjConfig StorjConfig `toml:"storj"`
	WebPConfig  WebPConfig  `toml:"webp"`
	// the url format to use for the uploaded files
	FormatUrl string `toml:"format_url"`
}

var (
	initConf   = flag.Bool("init-conf", false, "initialize the config file")
	configFile = flag.String("config", getDefaultConfigFile(), "the config file to use")
)

var defaultConfig = Config{
	FormatUrl: "https://example.com/{{bucket}}/{{prefix}}/{{filename}}",
	WebPConfig: WebPConfig{
		Enabled:   true,
		MimeTypes: []string{"image/jpeg", "image/png", "image/gif"},
		Quality:   75,
	},
	StorjConfig: StorjConfig{AccessGrant: "",
		BucketName: "uplog",
		Prefix:     "folder",
	},
}

func main() {
	flag.Parse()
	if *initConf {
		err := writeConfig(*configFile, defaultConfig)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	config, err := loadConfig(*configFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = writeConfig(*configFile, config)
	if err != nil {
		fmt.Println("unable to update config", err)
	}

	inputFile := flag.Arg(0)
	if inputFile == "" {
		fmt.Println("no input file specified")
		return
	}
	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	mime, err := mimetype.DetectReader(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	file.Seek(0, io.SeekStart)
	reader, mime, err := EncodeWebP(file, mime, config.WebPConfig)
	if err != nil {
		fmt.Println(err)
		return
	}
	url, err := SaveMediaFile(reader, mime, config.StorjConfig.AccessGrant, config.StorjConfig.BucketName, config.StorjConfig.Prefix, config.FormatUrl)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("uploaded file to:")
	fmt.Println(url)
}

func writeConfig(file string, config Config) error {
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("could not create config directory: %v", err)
	}

	buf := bytes.NewBuffer(nil)
	err := toml.NewEncoder(buf).Encode(config)

	if err != nil {
		return fmt.Errorf("could not encode config: %v", err)
	}

	os.WriteFile(file, buf.Bytes(), 0660)

	if err != nil {
		return fmt.Errorf("could not write config file: %v", err)
	}
	return nil
}

func loadConfig(file string) (Config, error) {

	var config Config
	_, err := toml.DecodeFile(file, &config)
	if err != nil {
		return Config{}, fmt.Errorf("could not load config: %v", err)
	}
	return config, nil
}

func getDefaultConfigFile() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		dir = os.Getenv("HOME") + "/.config"
	}
	return dir + "/uplog/config.toml"
}
