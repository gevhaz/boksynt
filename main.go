package main

import (
    "os/exec"
    "flag"
    "fmt"
    "log"
    "os"
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
