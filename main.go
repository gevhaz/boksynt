package main

import (
    "errors"
    "flag"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"

    readability "github.com/go-shiori/go-readability"
)

func main() {
    colorError := "\033[31m"
    colorOk := "\033[32m"
    colorWarning := "\033[33m"
    colorReset := "\033[0m"

    currentDirectory, err := os.Getwd()
    if err != nil {
        log.Fatal(err)
    }

    var urlsFilePath string
    var outputDir string

    flag.StringVar(&urlsFilePath, "file", "", "Path to file with URLs to articles to be fetched and converted")
    flag.StringVar(&outputDir, "output-dir", currentDirectory, "Directory where the final epub files should be placed")

    flag.Parse()

    urlData, err := os.ReadFile(urlsFilePath)
    if err != nil {
        log.Fatal(err)
    }

    urls := strings.Split(strings.Trim(string(urlData), "\n"), "\n")

    tempDir, err := os.MkdirTemp("", "boksynt")
    if err != nil {
        log.Fatal(err)
    }
    defer os.RemoveAll(tempDir)

    for i, url := range urls {
        if i > 0 {
            fmt.Println("")
        }
        log.Printf("Processing URL number %01d: %s\n", i+1, url)

        article, err := readability.FromURL(url, 30 * time.Second)
        if err != nil {
            log.Fatalf("Failed to parse url: %s\n%v\n", url, err)
            continue
        }

        articleSafeName := strings.ReplaceAll(strings.ToLower(article.Title), " ", "_")
        epubName := articleSafeName + ".epub"

        if fileExists(epubName) {
            log.Printf(colorWarning + "A file with the name '%s' already exists, skipping" + colorReset, epubName)
            continue
        }

        coverImagePath := filepath.Join(tempDir, articleSafeName + ".jpg")
        err = downloadFile(article.Image, coverImagePath)
        if err != nil {
            log.Fatal(err)
        }

        htmlPath := filepath.Join(tempDir, articleSafeName + ".html")
        htmlFile, err := os.Create(htmlPath)
        if err != nil {
            log.Fatal(err)
        }
        defer htmlFile.Close()
        htmlFile.WriteString(article.Content)

        cmd := exec.Command(
            "pandoc",
			"-f",
			"html",
			"-t",
			"epub",
			"-o",
            epubName,
            "--metadata",
            fmt.Sprintf("title: %s", article.Title),
            "--metadata",
            fmt.Sprintf("author: %s", article.Byline),
            "--epub-cover-image",
            coverImagePath,
            htmlPath,
        )

        err = cmd.Run()
        if err != nil {
            log.Fatalf(colorError + "Error converting %s\n" + colorReset, article.Title)
            log.Fatal(err)
        } else {
            log.Printf(colorOk + "Successfully converted %s\n" + colorReset, article.Title)
        }
    }
}

func downloadFile(url string, filename string) error {
    response, err := http.Get(url)
    if err != nil {
        return err
    }

    defer response.Body.Close()

    if response.StatusCode != 200 {
        return errors.New(fmt.Sprintf("Request did not succeed. Response code: %d", response.StatusCode))
    }

    file, err := os.Create(filename)
    if err != nil {
        return err
    }

    defer file.Close()
    _, err = io.Copy(file, response.Body)
    if err != nil {
        return nil
    }

    return nil
}

// Takes a string and checks whether the path it represents is a file or
// directory, in which case true is returned, otherwise false is returned.
func fileExists(path string) bool {
    _, err := os.Stat(path)
    if os.IsNotExist(err) {
        return false
    }
    return true
}
