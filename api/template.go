package api

import (
	"io"
	"text/template"

	"github.com/GeertJohan/go.rice"
	"github.com/labstack/echo/v4"
)

type TemplateRenderer struct {
	templates *template.Template
}

func NewTemplateRenderer() (*TemplateRenderer, error) {
	templateBox, err := rice.FindBox("templates")
	if err != nil {
		return nil, err
	}

	ipxeString, err := templateBox.String("ipxe.tmpl")
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("ipxe.tmpl").Parse(ipxeString)
	if err != nil {
		return nil, err
	}

	kickstartString, err := templateBox.String("kickstart.tmpl")
	if err != nil {
		return nil, err
	}

	tmpl, err = tmpl.New("kickstart.tmpl").Parse(kickstartString)
	if err != nil {
		return nil, err
	}

	t := &TemplateRenderer{
		templates: tmpl,
	}

	return t, nil
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)

	return t.templates.ExecuteTemplate(w, name, data)
}
