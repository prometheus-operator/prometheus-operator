package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/ghodss/yaml"
)

func main() {
	yamltojson := flag.Bool("yamltojson", false, "Convert yaml to json instead of the default json to yaml.")
	flag.Parse()

	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal(err)
	}
	if fi.Mode()&os.ModeNamedPipe == 0 {
		log.Fatal("no data to read from stdin")
	}

	inBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	var outBytes []byte
	if *yamltojson {
		outBytes, err = yaml.YAMLToJSON(inBytes)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		outBytes, err = yaml.JSONToYAML(inBytes)
		if err != nil {
			log.Fatal(err)
		}
	}

	os.Stdout.Write(outBytes)
}
