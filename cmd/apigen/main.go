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

type params struct {
	Package       string `yaml:"package"`
	Type          string `yaml:"struct_type"`
	OrgName       string `yaml:"org_name"`
	RepoName      string `yaml:"repo_name"`
	TemplatesPath string `yaml:"templates_path"`
}

type config struct {
	TemplatesPath string `yaml:"templates_path"`
	Github        struct {
		Org  string `yaml:"org"`
		Repo string `yaml:"repo"`
	} `yaml:"github"`
	Components []struct {
		Package string `yaml:"package"`
		Type    string `yaml:"struct_type"`
	} `yaml:"components"`
}

func main() {
	var fileConfig string
	flag.StringVar(&fileConfig, "f", "", "the conf.yaml path")
	flag.Parse()

	blob, err := ioutil.ReadFile(fileConfig)
	if err != nil {
		log.Fatalln("can't find config file:", err)
	}

	var conf config
	err = yaml.Unmarshal(blob, &conf)
	if err != nil {
		log.Fatalln("config yaml error:", err)
	}

	for _, item := range conf.Components {
		p := params{
			Package:       item.Package,
			Type:          item.Type,
			OrgName:       conf.Github.Org,
			RepoName:      conf.Github.Repo,
			TemplatesPath: conf.TemplatesPath,
		}

		walk(p)
	}
}

func walk(p params) {
	outDir := p.Package

	err := filepath.Walk(p.TemplatesPath, func(path string, info os.FileInfo, err error) error {
		// replace original template path with new path
		newPath := strings.Replace(path, p.TemplatesPath, outDir, 1)
		// drop .in ext
		newPath = strings.Replace(newPath, ".go.in", ".go", 1)
		// replace __package__ with package name
		newPath = strings.Replace(newPath, "__package__", p.Package, 1)

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
