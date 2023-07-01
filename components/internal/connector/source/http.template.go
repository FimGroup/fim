package source

import (
	"html/template"
	"io"

	"github.com/FimGroup/fim/fimapi/pluginapi"
)

type templateRender struct {
	template *template.Template
}

func (t templateRender) Render(w io.Writer, data any) error {
	return t.template.Execute(w, data)
}

type templateFolder struct {
	fileTemplateMapping map[string]templateRender
}

func newTemplateFolder() *templateFolder {
	return &templateFolder{
		fileTemplateMapping: map[string]templateRender{},
	}
}

func (t *templateFolder) loadAndSaveTemplate(path string, fm pluginapi.FileResourceManager) (templateRender, error) {
	tr, err := loadTemplate(path, fm)
	if err != nil {
		return templateRender{}, err
	}
	t.fileTemplateMapping[path] = tr
	return tr, nil
}

func loadTemplate(path string, fm pluginapi.FileResourceManager) (templateRender, error) {
	funcMap := template.FuncMap{}

	data, err := fm.LoadFile(path)
	if err != nil {
		return templateRender{}, err
	}

	tmpl, err := template.New("").Funcs(funcMap).Parse(string(data))
	if err != nil {
		return templateRender{}, err
	}

	tr := templateRender{template: tmpl}

	return tr, nil
}
