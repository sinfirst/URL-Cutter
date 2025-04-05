package files

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"github.com/sinfirst/URL-Cutter/internal/app/storage"
)

type JSONStructForBD struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type File struct {
	config config.Config
	UUID   int
}

func NewFile(cfg config.Config, stg *storage.MapStorage) *File {
	f := &File{config: cfg}
	f.ReadFile(stg)
	return f
}

func (f *File) UpdateFile(jsonStruct JSONStructForBD) {

	f.UUID++
	jsonStruct.UUID = strconv.Itoa(f.UUID)

	err := os.MkdirAll(filepath.Dir(f.config.FilePath), os.ModePerm)

	if err != nil {
		return
	}

	file, err := os.OpenFile(f.config.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return
	}
	defer file.Close()

	jsonData, err := json.Marshal(jsonStruct)

	if err != nil {
		return
	}
	jsonData = append(jsonData, '\n')

	_, err = file.Write(jsonData)

	if err != nil {
		return
	}
}

func (f *File) ReadFile(strg storage.Storage) {

	var jsonStrct JSONStructForBD
	err := os.MkdirAll(filepath.Dir(f.config.FilePath), os.ModePerm)

	if err != nil {
		return
	}

	file, err := os.OpenFile(f.config.FilePath, os.O_RDONLY|os.O_CREATE, 06666)

	if err != nil {
		return
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		json.Unmarshal(scanner.Bytes(), &jsonStrct)
		strg.Set(jsonStrct.ShortURL, jsonStrct.OriginalURL)
	}

	f.UUID, err = strconv.Atoi(jsonStrct.UUID)
	if err != nil {
		return
	}
}
