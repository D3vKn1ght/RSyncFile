package main

import (
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "time"
)

func checkForChanges(folder string, lastModTimes map[string]time.Time) {
    files, _ := ioutil.ReadDir(folder)

    for _, file := range files {
        if !file.IsDir() {
            filePath := filepath.Join(folder, file.Name())
            fileModTime := file.ModTime()

            if lastModTimes[filePath].Before(fileModTime) {
                lastModTimes[filePath] = fileModTime
                if err := createTempAndUpload(filePath); err != nil {
                    log.Printf("Failed to upload file %s: %v\n", filePath, err)
                }
            }
        }
    }
}

func createTempAndUpload(filePath string) error {
    tmpFilePath := filePath + ".tmp"

    if err := copyFile(filePath, tmpFilePath); err != nil {
        return err
    }
    defer os.Remove(tmpFilePath)

    if err := uploadFile(tmpFilePath); err != nil {
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

func uploadFile(filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    request, err := http.NewRequest("POST", "http://localhost:8000/uploadfile/", file)
    if err != nil {
        return err
    }

    request.Header.Set("Content-Type", "application/octet-stream")
    request.Header.Set("filename", filepath.Base(filePath))

    client := &http.Client{}
    response, err := client.Do(request)
    if err != nil {
        return err
    }
    defer response.Body.Close()

    if response.StatusCode != http.StatusOK {
        return err
    }

    return nil
}

func main() {
    folderToWatch := "./folder_to_watch"
    lastModTimes := make(map[string]time.Time)

    for {
        checkForChanges(folderToWatch, lastModTimes)
        time.Sleep(10 * time.Second)
    }
}
