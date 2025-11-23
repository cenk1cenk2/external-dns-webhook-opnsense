package services

import (
	"bytes"
	"fmt"
	"html/template"

	sprig "github.com/go-task/slim-sprig/v3"
)

func InlineTemplate[Ctx any](tmpl string, ctx Ctx, funcs ...template.FuncMap) (string, error) {
	if tmpl == "" {
		return "", nil
	}

	parser := template.New("inline").Funcs(TemplateFuncMap())

	for _, f := range funcs {
		parser.Funcs(f)
	}

	tmp, err := parser.Parse(tmpl)

	if err != nil {
		return "", fmt.Errorf("Can not create inline template: %w", err)
	}

	var w bytes.Buffer

	err = tmp.ExecuteTemplate(&w, "inline", ctx)

	if err != nil {
		return "", fmt.Errorf("Can not generate inline template: %w", err)
	}

	return w.String(), nil
}

func TemplateFuncMap() template.FuncMap {
	// functions can be found here: https://go-task.github.io/slim-sprig/
	return sprig.FuncMap()
}
