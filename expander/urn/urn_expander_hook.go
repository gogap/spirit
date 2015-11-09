package urn

import (
	"bytes"
	"errors"
	"strings"
	"text/template"

	"github.com/gogap/spirit"
)

const (
	hookURNExpanderURN = "urn:spirit:expander:urn:hook"
)

var (
	ErrHookURNExpanderTemplateDuplicate = errors.New("hook-urn-expander template dumplicate")
)

type HookURNExpanderConfig struct {
	Hooks map[string]string `hooks`
}

type HookURNExpander struct {
	conf HookURNExpanderConfig

	tmpls map[string]*template.Template
}

func init() {
	spirit.RegisterURNExpander(hookURNExpanderURN, NewHookURNExpander)
}

func NewHookURNExpander(options spirit.Options) (expander spirit.URNExpander, err error) {
	conf := HookURNExpanderConfig{}

	if err = options.ToObject(&conf); err != nil {
		return
	}

	tmpTmpls := make(map[string]*template.Template)
	if conf.Hooks != nil {
		for hookURN, urnTemplate := range conf.Hooks {
			if _, exist := tmpTmpls[hookURN]; exist {
				err = ErrHookURNExpanderTemplateDuplicate
				return
			}
			var tmpl *template.Template
			if tmpl, err = template.New(hookURN).Parse(urnTemplate); err != nil {
				return
			}
			tmpl = tmpl.Option("missingkey=error").Delims("<?", "?>")
			tmpTmpls[hookURN] = tmpl
		}
	}

	expander = &HookURNExpander{
		conf:  conf,
		tmpls: tmpTmpls,
	}

	return
}

func (p *HookURNExpander) Expand(delivery spirit.Delivery) (newURN string, err error) {
	originalURN := delivery.URN()
	originalURN = strings.Trim(originalURN, "|")

	if p.tmpls == nil || originalURN == "" {
		return delivery.URN(), nil
	}

	var tmpl *template.Template
	var exist bool

	if tmpl, exist = p.tmpls[originalURN]; !exist {
		newURN = originalURN
		return
	}

	urns := strings.Split(originalURN, "|")

	tmpURNs := []string{}

	for _, urn := range urns {
		if urn != "" {
			var buf bytes.Buffer
			if err = tmpl.Execute(&buf, map[string]interface{}{"urn": urn}); err != nil {
				return
			}
			tmpURNs = append(tmpURNs, buf.String())
		}
	}

	newURN = strings.Join(tmpURNs, "|")

	return
}
