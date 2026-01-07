package template

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
)

var baseDir = "internal/template/email"

type Parser struct {
	templates map[string]*template.Template
}

func NewParser() (*Parser, error) {
	parser := &Parser{
		templates: make(map[string]*template.Template),
	}

	if err := parser.loadTemplates(); err != nil {
		return nil, err
	}

	return parser, nil
}

func (p *Parser) loadTemplates() error {
	pattern := filepath.Join(baseDir, "*.html")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to find template files: %w", err)
	}

	for _, file := range files {
		name := filepath.Base(file)
		name = name[:len(name)-len(filepath.Ext(name))]

		tmpl, err := template.ParseFiles(file)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}

		p.templates[name] = tmpl
	}

	return nil
}

func (p *Parser) Render(templateName string, data map[string]any) (subject string, body string, err error) {
	tmpl, exists := p.templates[templateName]
	if !exists {
		return "", "", fmt.Errorf("template %s not found", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", "", fmt.Errorf("failed to execute template: %w", err)
	}

	if subj, ok := data["Subject"].(string); ok {
		subject = subj
	} else {
		subject = "Notification from E-Commerce Platform"
	}

	body = buf.String()
	return subject, body, nil
}
