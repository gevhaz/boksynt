package main

import (
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
        }

        fmt.Println(article.Title)
    }
}
