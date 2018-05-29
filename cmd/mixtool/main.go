package main

import (
	"fmt"
	"os"
	"path"

	jsonnet "github.com/google/go-jsonnet"
	"github.com/prometheus/prometheus/pkg/rulefmt"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	yaml "gopkg.in/yaml.v2"
)

var (
	app     = kingpin.New("mixtool", "")
	lint    = app.Command("lint", "")
	lintDir = lint.Arg("mixin", "").String()
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case lint.FullCommand():
		lintFn()
	}
}

func lintFn() {
	root := *lintDir
	if fi, err := os.Stat(root); err != nil || fi.IsDir() {
		root = path.Join(root, "mixin.libsonnet")
	}

	snippet := fmt.Sprintf("(import \"%s\").prometheusAlerts", root)
	vm := jsonnet.MakeVM()
	rulesRawJSON, err := vm.EvaluateSnippet("", snippet)
	if err != nil {
		app.Fatalf("failed to parse mixin: %v", err)
		return
	}

	// convert to pretty yaml
	var rules map[string]interface{}
	if err := yaml.Unmarshal([]byte(rulesRawJSON), &rules); err != nil {
		app.Fatalf("failed to parse json: %v", err)
	}
	rulesRawYAML, err := yaml.Marshal(rules)
	if err != nil {
		app.Fatalf("failed to marshall yaml: %v", err)
	}

	_, errs := rulefmt.Parse(rulesRawYAML)
	if len(errs) > 0 {
		for _, err := range errs {
			app.Errorf("rulefmt verification: %v", err)
		}
	}
}
