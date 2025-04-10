package files

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"go.uber.org/zap"
)

type JSONStructForFile struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type File struct {
	config config.Config
	logger zap.SugaredLogger
	UUID   int
}

func NewFile(config config.Config, logger zap.SugaredLogger) *File {
	f := &File{config: config, logger: logger}
	return f
}

func (f *File) Set(shortURL, origURL string) bool { //jsonStruct JSONStruct,

	jsonStruct := JSONStructForFile{
		ShortURL:    shortURL,
		OriginalURL: origURL,
	}
	f.UUID++
	jsonStruct.UUID = strconv.Itoa(f.UUID)

	err := os.MkdirAll(filepath.Dir(f.config.FilePath), os.ModePerm)

	if err != nil {
		return false
	}

	file, err := os.OpenFile(f.config.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		f.logger.Infow("Problem with open file")
		return false
	}

	defer file.Close()

	jsonData, err := json.Marshal(jsonStruct)
	if err != nil {
		f.logger.Errorw("Problem with marshal JSONStruct")
		return false
	}
	jsonData = append(jsonData, '\n')

	_, err = file.Write(jsonData)

	if err != nil {
		f.logger.Errorw("Problem with write into file")
		return false
	}
	return true
}

func (f *File) Get(shortURL string) (string, bool) {
	data := make(map[string]string)

	var jsonStrct JSONStructForFile
	err := os.MkdirAll(filepath.Dir(f.config.FilePath), os.ModePerm)

	if err != nil {
		return "", false
	}

	file, err := os.OpenFile(f.config.FilePath, os.O_RDONLY|os.O_CREATE, 06666)

	if err != nil {
		f.logger.Infow("Problem with open file")
		return "", false
	}

	f.logger.Infow("created file in direction: " + f.config.FilePath)

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		json.Unmarshal(scanner.Bytes(), &jsonStrct)
		data[jsonStrct.ShortURL] = jsonStrct.OriginalURL
	}

	f.UUID, _ = strconv.Atoi(jsonStrct.UUID)

	origURL, flag := data[shortURL]
	if !flag {
		f.logger.Infow("No short URL in File")
		return "", false
	}

	return origURL, true
}
