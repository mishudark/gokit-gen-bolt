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
	PackageAbbr   string
	Package       string
	PkgClient     string
	Type          string
	OrgName       string
	RepoName      string
	TemplatesPath string
	Common        []string
}

type config struct {
	TemplatesPath string `yaml:"templates_path"`
	Github        struct {
		Org  string `yaml:"org"`
		Repo string `yaml:"repo"`
	} `yaml:"github"`
	Common     []string `yaml:"common"`
	PkgClient  string   `yaml:"pkg_client"`
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
		letters := []rune(item.Package)
		abbr := ""
		if len(letters) > 0 {
			abbr = string(letters[0])
		}

		p := params{
			PackageAbbr:   abbr,
			PkgClient:     conf.PkgClient,
			Package:       item.Package,
			Type:          item.Type,
			OrgName:       conf.Github.Org,
			RepoName:      conf.Github.Repo,
			TemplatesPath: conf.TemplatesPath,
			Common:        conf.Common,
		}

		walk(p)
	}
}

func walk(p params) {
	outDir := p.Package
	commonPrefix := ""
	commonPath := ""

	err := filepath.Walk(p.TemplatesPath, func(path string, info os.FileInfo, err error) error {
		// replace original template path with new path
		newPath := strings.Replace(path, p.TemplatesPath, outDir, 1)
		// drop .in ext
		newPath = strings.Replace(newPath, ".go.in", ".go", 1)
		// replace __package__ with package name
		newPath = strings.Replace(newPath, "__package__", p.Package, 1)

		if info.IsDir() {
			for _, item := range p.Common {
				// check if is a common component, creates the dir in the root
				// FIXME: improve checking if its a real dir
				// currently match ie:
				// item = 'bolt'
				// newPath = /some/boltBar
				// clearly this is an error
				if strings.Contains(newPath, item) {
					commonPrefix = strings.Replace(newPath, item, "", 1) + item
					commonPath = item
					newPath = item

					break
				}

				commonPrefix = ""
				commonPath = ""
			}

			if err := os.Mkdir(newPath, 0755); err != nil {
				log.Println(err)
			}

			return nil
		}

		// replace prefix for common components
		if strings.Contains(newPath, commonPrefix) {
			newPath = strings.Replace(newPath, commonPrefix, "", 1)
			newPath = commonPath + newPath
		}

		log.Println(newPath)

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
