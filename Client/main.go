package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	URL           string `json:"url"`
	FolderToWatch string `json:"folder_to_watch"`
}

func loadConfig(configPath string) (Config, error) {
	var config Config

	file, err := os.Open(configPath)
	if err != nil {
		return config, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(bytes, &config)
	return config, err
}

func convertToLinuxPath(winPath string) string {
	return strings.ReplaceAll(winPath, "\\", "/")
}

func checkForChanges(folder string, lastModTimes map[string]time.Time, serverURL string) {
	files, _ := ioutil.ReadDir(folder)
	seenFiles := make(map[string]bool)

	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(folder, file.Name())
			fileModTime := file.ModTime()
			seenFiles[filePath] = true

			if lastModTimes[filePath].Before(fileModTime) {
				lastModTimes[filePath] = fileModTime
				if err := createTempAndUpload(filePath, folder, serverURL); err != nil {
					log.Printf("Failed to upload file %s: %v\n", filePath, err)
				}
			}
		} else {
			checkForChanges(filepath.Join(folder, file.Name()), lastModTimes, serverURL)
		}
	}

	for filePath := range lastModTimes {
		if !seenFiles[filePath] {
			if err := deleteFileOnServer(filePath, folder, serverURL); err != nil {
				log.Printf("Failed to delete file %s on server: %v\n", filePath, err)
			}
			delete(lastModTimes, filePath)
		}
	}
}

func createTempAndUpload(filePath, rootFolder, serverURL string) error {
	tmpFilePath := filePath + ".tmp"

	if err := copyFile(filePath, tmpFilePath); err != nil {
		return err
	}
	defer os.Remove(tmpFilePath)

	relativePath := strings.TrimPrefix(tmpFilePath, rootFolder)
	relativePath = convertToLinuxPath(relativePath)

	if err := gzipAndUploadFile(tmpFilePath, relativePath, serverURL); err != nil {
		return err
	}

	os.Rename(tmpFilePath, filePath)
	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	return err
}

func gzipAndUploadFile(filePath, relativePath, serverURL string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipFilePath := filePath + ".gz"
	gzipFile, err := os.Create(gzipFilePath)
	if err != nil {
		return err
	}
	defer os.Remove(gzipFilePath)
	defer gzipFile.Close()

	gzipWriter := gzip.NewWriter(gzipFile)
	if _, err := io.Copy(gzipWriter, file); err != nil {
		return err
	}
	gzipWriter.Close()

	gzipFile, err = os.Open(gzipFilePath)
	if err != nil {
		return err
	}
	defer gzipFile.Close()

	request, err := http.NewRequest("POST", serverURL+"/uploadfile/", gzipFile)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/gzip")
	request.Header.Set("filename", relativePath)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed with status: %s", response.Status)
	}

	return nil
}

func deleteFileOnServer(filePath, rootFolder, serverURL string) error {
	relativePath := strings.TrimPrefix(filePath, rootFolder)
	relativePath = convertToLinuxPath(relativePath)

	request, err := http.NewRequest("DELETE", serverURL+"/deletefile/", nil)
	if err != nil {
		return err
	}

	request.Header.Set("filename", relativePath)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed with status: %s", response.Status)
	}

	return nil
}

func main() {
	configPath := "config.json"

	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	folderToWatch := config.FolderToWatch
	serverURL := config.URL

	lastModTimes := make(map[string]time.Time)

	for {
		checkForChanges(folderToWatch, lastModTimes, serverURL)
		time.Sleep(10 * time.Second)
	}
}
