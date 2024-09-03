package main

import (
    "bytes"
    "compress/gzip"
    "encoding/json"
    "fmt"
    "io"
    "mime/multipart"
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

    bytes, err := io.ReadAll(file)
    if err != nil {
        return config, err
    }

    err = json.Unmarshal(bytes, &config)
    return config, err
}

func convertToLinuxPath(winPath string) string {
    return strings.ReplaceAll(winPath, "\\", "/")
}

func checkForChanges(folderToWatch string, folder string, lastModTimes map[string]time.Time, serverURL string) {
    files, _ := os.ReadDir(folder)

    // Duyệt qua các file hiện tại trong thư mục
    for _, file := range files {
        if !file.IsDir() {
            filePath := filepath.Join(folder, file.Name())
            info, err := os.Stat(filePath)
            if err != nil {
                continue
            }
            fileModTime := info.ModTime()

            if lastModTime, exists := lastModTimes[filePath]; !exists || lastModTime != fileModTime {
                lastModTimes[filePath] = fileModTime
                fmt.Printf("Detected change in folder: %s\n", folder)
                fmt.Printf("File: %s, Last Modified: %s\n", file.Name(), fileModTime)
                err := gzipAndUploadFile(folderToWatch, filePath, folder, serverURL)
                if err != nil {
                    fmt.Printf("Error uploading file %s: %v\n", file.Name(), err)
                    continue
                }
            }
        } else {
            checkForChanges(folderToWatch, filepath.Join(folder, file.Name()), lastModTimes, serverURL)
        }
    }

    for filePath := range lastModTimes {
        if _, err := os.Stat(filePath); os.IsNotExist(err) {
            fmt.Printf("Detected not seen: %s\n", filePath)
            err := deleteFileOnServer(folderToWatch, filePath, folder, serverURL)
            if err == nil {
                delete(lastModTimes, filePath)
            }
        }
    }
}

func gzipAndUploadFile(folder,filePath, rootFolder, serverURL string) error {

    relativePath := strings.TrimLeft(filePath, folder)
    relativePath = convertToLinuxPath(relativePath) // Chuyển đổi sang dạng Linux

    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    var gzipBuffer bytes.Buffer
    gzipWriter := gzip.NewWriter(&gzipBuffer)
    if _, err := io.Copy(gzipWriter, file); err != nil {
        return err
    }
    gzipWriter.Close()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", filepath.Base(filePath)+".gz")
    if err != nil {
        return err
    }
    _, err = io.Copy(part, &gzipBuffer)
    if err != nil {
        return err
    }

    writer.Close()

    request, err := http.NewRequest("POST", serverURL+"/uploadfile/", body)
    if err != nil {
        return err
    }

    request.Header.Set("Content-Type", writer.FormDataContentType())
    request.Header.Set("X-Filename", relativePath+".gz")

    client := &http.Client{}
    response, err := client.Do(request)
    if err != nil {
        return err
    }
    defer response.Body.Close()

    if response.StatusCode != http.StatusOK {
        return fmt.Errorf("upload failed with status: %s", response.Status)
    }

    fmt.Printf("File %s uploaded successfully to %s\n", relativePath, serverURL)
    return nil
}


func deleteFileOnServer(folderToWatch,filePath, rootFolder, serverURL string) error {
    relativePath := strings.TrimLeft(filePath, folderToWatch)
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

    fmt.Printf("File %s deleted successfully from server\n", relativePath)  // Print thông tin file bị xóa thành công
    return nil
}

func main() {
    configPath := "config.json"
    config, err := loadConfig(configPath)
    if err != nil {
        fmt.Printf("Failed to load config: %v\n", err)
        return
    }

    folderToWatch := config.FolderToWatch
    serverURL := config.URL
    lastModTimes := make(map[string]time.Time)

    for {
        checkForChanges(folderToWatch,folderToWatch, lastModTimes, serverURL)
        time.Sleep(10 * time.Second)
    }
}
