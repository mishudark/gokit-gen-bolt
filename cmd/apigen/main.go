package main

import (
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type config struct {
	File string
}

type params struct {
	Package       string `yaml:"package"`
	Type          string `yaml:"struct_type"`
	OrgName       string `yaml:"org_name"`
	RepoName      string `yaml:"repo_name"`
	TemplatesPath string `yaml:"templates_path"`
}

func main() {
	var c config
	flag.StringVar(&c.File, "f", "", "the conf.yaml path")
	flag.Parse()

	blob, err := ioutil.ReadFile(c.File)
	if err != nil {
		log.Fatalln("can't find config file:", err)
	}

	var p params
	err = yaml.Unmarshal(blob, &p)
	if err != nil {
		log.Fatalln("config yaml error:", err)
	}

	walk(p)
}

func walk(p params) {
	outDir := p.Package

	err := filepath.Walk(p.TemplatesPath, func(path string, info os.FileInfo, err error) error {
		// replace original template path with new path
		newPath := strings.Replace(path, p.TemplatesPath, outDir, 1)
		// drop .in ext
		newPath = strings.Replace(newPath, ".go.in", ".go", 1)
		log.Println(newPath)

		if info.IsDir() {
			if err := os.Mkdir(newPath, 0755); err != nil {
				log.Fatalln(err)
			}

			return nil
		}

		f, err := os.Create(newPath)
		if err != nil {
			log.Fatalln("create file: ", err)
		}

		defer f.Close() // nolint: errcheck

		blob, err := ioutil.ReadFile(path)
		if err != nil {
			log.Println(err)
			return err
		}
		t := template.Must(template.New("queue").Parse(string(blob)))
		return t.Execute(f, p)
	})

	if err != nil {
		log.Fatalln(err)
	}
}
