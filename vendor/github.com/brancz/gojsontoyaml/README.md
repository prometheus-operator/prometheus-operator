# gojsontoyaml

This is a small tool written in go to convert json to yaml reading from STDIN and writing to STDOUT. The heavy lifting is actually done by [ghodss/yaml](https://github.com/ghodss/yaml) and [gopkg.in/yaml.v2](http://gopkg.in/yaml.v2).

## Install

To install simply

```
go get github.com/brancz/gojsontoyaml
```

## Usage

Simply pipe a json string into `gojsontoyaml` and it will print the converted yaml string to STDOUT.

```
$ echo '{"test":"test string with\\nmultiple lines"}' | gojsontoyaml
test: |-
  test string with
  multiple lines
```

## Motivation

You may ask yourself why this was developed. The answer is simple, when I wrote this there was no simple to use binary for this purpose that supported yaml multiline strings. All alternatives out there that I tried kept line breaks in the string rather than making use of the yaml multiline strings.
