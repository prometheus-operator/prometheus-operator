# liche

[![Circle CI](https://img.shields.io/circleci/project/github/raviqqe/liche.svg?style=flat-square)](https://circleci.com/gh/raviqqe/liche)
[![Codecov](https://img.shields.io/codecov/c/github/raviqqe/liche.svg?style=flat-square)](https://codecov.io/gh/raviqqe/liche)
[![Go Report Card](https://goreportcard.com/badge/github.com/raviqqe/liche?style=flat-square)](https://goreportcard.com/report/github.com/raviqqe/liche)
[![License](https://img.shields.io/github/license/raviqqe/liche.svg?style=flat-square)](LICENSE)

[![asciicast](https://asciinema.org/a/148896.png)](https://asciinema.org/a/148896)

`liche` is a command to check links' connectivity in Markdown and HTML files.
It checks all `a` and `img` tags in specified files.

## Installation

```sh
go get -u github.com/raviqqe/liche
```

## Usage

```sh
> liche --help
Link checker for Markdown and HTML

Usage:
	liche [-c <num-requests>] [-d <directory>] [-r] [-t <timeout>] [-x <regex>] [-v] <filenames>...

Options:
	-c, --concurrency <num-requests>  Set max number of concurrent HTTP requests. [default: 512]
	-d, --document-root <directory>  Set document root directory for absolute paths.
	-r, --recursive  Search Markdown and HTML files recursively
	-t, --timeout <timeout>  Set timeout for HTTP requests in seconds. Disabled by default.
	-x, --exclude <regex>  Regex of links to exclude from checking.
	-v, --verbose  Be verbose.
```

## Examples

```sh
> liche file.md
> liche file1.md file2.md
> liche file.html
> liche file1.html file2.html
> liche -r directory # Search all Markdown and HTML files recursively.
```

## Supported properties

- File extensions: `.md`, `.html`, `.htm`
- HTML tags: `a`, `img`
- HTML attributes: `href`, `src`
- URL schemes: `http`, `https`

Also supports relative and absolute paths.
(Absolute paths need `--document-root` option.)

## License

[MIT](LICENSE)
