package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"gopkg.in/yaml.v3"
)

const hashFile = "config/hashes.yml"

var GlobalHashes map[string]HashConfig

type HashConfig struct {
	MD5Digest  string `yaml:"md5digest"`
	SHA1Digest string `yaml:"sha1digest"`
}

func InitHashes() error {
	currentDir, err := os.Getwd()
	if err != nil {
		logger.Errorf("Failed opening current directory")
		return err
	}
	fileName := filepath.Join(currentDir, hashFile)
	yamlFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		logger.Fatalf("Error reading hash config file: %v", err)
		return err
	}

	err = yaml.Unmarshal(yamlFile, &GlobalHashes)
	if err != nil {
		logger.Fatalf("Error unmarshaling parameter yaml %v", err)
		return err
	}
	return nil
}
