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

var file_help string = "Path to file with URLs to articles to be fetched and converted"
var output_help string = "Directory where the final epub files should be placed"
var tag_help string = "Comma separated list of tags that should be added to articles"
var flag_usage string = fmt.Sprintf(`Usage of Boksynt:
  -f, --file string
		%s
  -o, --output-dir string
		%s
  -t, --tag string
		%s
`, file_help, output_help, tag_help)

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
	var articleTag string

	flag.StringVar(&urlsFilePath, "file", "", file_help)
	flag.StringVar(&urlsFilePath, "f", "", file_help)
	flag.StringVar(&outputDir, "output-dir", currentDirectory, output_help)
	flag.StringVar(&outputDir, "o", currentDirectory, output_help)
	flag.StringVar(&articleTag, "tag", "", tag_help)
	flag.StringVar(&articleTag, "t", "", tag_help)
	flag.Usage = func() { fmt.Print(flag_usage) }

	flag.Parse()

	if outputDir != currentDirectory {
		err = os.Mkdir(outputDir, 0755)
		if os.IsExist(err) {
			log.Printf("Writing output into existing directory: %s", outputDir)
		} else if err != nil {
			log.Fatal(err)
		} else {
			log.Printf("Successfully created directory '%s'", outputDir)
		}
	}

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
		fmt.Println("")

		log.Printf("Processing URL number %01d: %s\n", i+1, url)

		article, err := readability.FromURL(url, 30*time.Second)
		if err != nil {
			log.Fatalf("Failed to parse url: %s\n%v\n", url, err)
			continue
		} else if len(article.Content) < 200 {
			log.Print(colorWarning + "Downloaded content is shorter than 200 characters so the page probably was not parsed properly:" + colorReset)
			log.Print(colorWarning + "URL: " + url + colorReset)
			continue
		}

		articleSafeName := strings.ReplaceAll(strings.ToLower(article.Title), " ", "_")
		articleSafeName = strings.ReplaceAll(articleSafeName, "/", "_")
		epubPath := filepath.Join(outputDir, articleSafeName+".epub")

		if fileExists(epubPath) {
			log.Printf(colorWarning+"A file with the name '%s' already exists, skipping"+colorReset, epubPath)
			continue
		}

		coverImagePath := filepath.Join(tempDir, articleSafeName+".jpg")
		if article.Image != "" {
			err = downloadFile(article.Image, coverImagePath)
			if err != nil {
				log.Fatal(err)
			}
		}

		htmlPath := filepath.Join(tempDir, articleSafeName+".html")
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
			epubPath,
			"--metadata",
			fmt.Sprintf("title: %s", article.Title),
			"--metadata",
			fmt.Sprintf("author: %s", article.Byline),
			"--metadata",
			fmt.Sprintf("subject: %s", articleTag),
			"--epub-cover-image",
			coverImagePath,
			htmlPath,
		)

		if len(article.Image) == 0 {
			cmd = exec.Command(
				"pandoc",
				"-f",
				"html",
				"-t",
				"epub",
				"-o",
				epubPath,
				"--metadata",
				fmt.Sprintf("title: %s", article.Title),
				"--metadata",
				fmt.Sprintf("author: %s", article.Byline),
				"--metadata",
				fmt.Sprintf("subject: %s", articleTag),
				htmlPath,
			)
		}

		err = cmd.Run()
		if err != nil {
			log.Fatalf(colorError+"Error converting '%s'"+colorReset+"\n", article.Title)
			log.Fatal(err)
		} else {
			log.Printf(colorOk+"Successfully converted '%s'"+colorReset+"\n", article.Title)
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
