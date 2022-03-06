# boksynt

_boksynt_ is a simple CLI tool for downloading and converting articles to the EPUB format so that they can be put on an
e-book reader or just read without any clutter and offline in your favorite EPUB reading software.

## Installation

1. Clone this repo.
2. You can immediately use it with `go run main.go`.

## Requirements

You need to have [pandoc](https://github.com/jgm/pandoc/) installed for _boksynt_ to work. Other dependencies are
specified in `go.mod` and should be handled automatically.

## Usage

The following flags are available:

```
-file string
      Path to file with URLs to articles to be fetched and converted
-output-dir string
      Directory where the final epub files should be placed (default "current/working/directory/")
```

The flow is:

1. Create a file with URLs that you want downloaded and converted.
2. Run the app and provide the path to your file to the `-file` flag.

## Development status

This is a work in progress, but the basic functionality is there. Third-party software ([Mozilla
Readability](https://github.com/mozilla/readability)) handles the actual parsing of websites, so that part should be
pretty mature.
