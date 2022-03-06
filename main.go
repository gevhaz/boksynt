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
    "strings"
    "time"

    readability "github.com/go-shiori/go-readability"
)

func main() {
    var urls_file_path string
    flag.StringVar(&urls_file_path, "file", "", "Path to file with URLs to articles to be fetched and converted")
    flag.Parse()

    url_data, err := os.ReadFile(urls_file_path)
    if err != nil {
        log.Fatal(err)
    }

    urls := strings.Split(strings.Trim(string(url_data), "\n"), "\n")

    for i, url := range urls {
        fmt.Printf("Processing URL number %01d: %s\n", i+1, url)

        article, err := readability.FromURL(url, 30 * time.Second)
        if err != nil {
            log.Fatalf("Failed to parse url: %s\n%v\n", url, err)
            continue
        }

        cover_image_file_name := strings.ReplaceAll(article.Title, " ", "_") + ".jpg"
        err = downloadFile(article.Image, cover_image_file_name)
        if err != nil {
            log.Fatal(err)
        }

        cmd := exec.Command(
            "pandoc",
			"-f",
			"html",
			"-t",
			"epub",
			"-o",
			"article.epub",
            "--metadata",
            fmt.Sprintf("title: %s", article.Title),
            "--metadata",
            fmt.Sprintf("author: %s", article.Byline),
            "--epub-cover-image",
            cover_image_file_name,
			"article.html",
        )

        html_file, err := os.Create("article.html")
        if err != nil {
            log.Fatal(err)
        }
        defer html_file.Close()
        html_file.WriteString(article.Content)

        err = cmd.Run()
        if err != nil {
            fmt.Printf("Error converting %s\n", article.Title)
            log.Fatal(err)
        } else {
            fmt.Printf("Successfully converted %s\n", article.Title)
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
