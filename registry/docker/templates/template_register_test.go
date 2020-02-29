package dockerTemplates_test

import (
	"bytes"
	"fmt"
	"github.com/ihaiker/aginx/logs"
	"github.com/ihaiker/aginx/registry/docker"
	dockerTemplates "github.com/ihaiker/aginx/registry/docker/templates"
	"github.com/ihaiker/aginx/util"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"testing"
	"text/template"
)

func init() {
	logs.SetLevel(logrus.DebugLevel)
}

func TestTemplateRegister(t *testing.T) {

	templateContent, err := ioutil.ReadFile("../_test/docker.tpl")
	if err != nil {
		t.Fatal(err)
	}

	reg, err := dockerTemplates.TemplateRegister("10.24.0.1", []string{"^portainer_portainer$"}, []string{".*"})
	if err != nil {
		t.Fatal(err)
	}
	if err := reg.Start(); err != nil {
		t.Fatal(err)
	}

	for event := range reg.Listener() {
		if tem, err := template.New("").Funcs(docker.Plugin.TemplateFuns()).Parse(string(templateContent)); err != nil {
			t.Fatal(err)
		} else {
			out := bytes.NewBufferString("")
			if err := tem.Execute(out, event); err != nil {
				t.Fatal(err)
			}
			fmt.Println(string(util.CleanEmptyLine(out.Bytes())))
		}
		break
	}
}
