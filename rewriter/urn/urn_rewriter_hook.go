package urn

import (
	"bytes"
	"errors"
	"strings"
	"text/template"

	"github.com/gogap/spirit"
)

const (
	hookURNRewriterURN = "urn:spirit:rewriter:urn:hook"
)

var _ spirit.URNRewriter = new(HookURNRewriter)
var _ spirit.Actor = new(HookURNRewriter)

var (
	ErrHookURNRewriterTemplateDuplicate = errors.New("hook-urn-rewriter template duplicate")
)

type TemplateDelims struct {
	Left  string `json:"left"`
	Right string `json:"right"`
}

type HookURNRewriterConfig struct {
	WholeMatching bool              `json:"whole_matching"`
	Hooks         map[string]string `json:"hooks"`
	Delims        TemplateDelims    `json:"delims"`
}

type HookURNRewriter struct {
	name string
	conf HookURNRewriterConfig

	tmpls map[string]*template.Template
}

func init() {
	spirit.RegisterURNRewriter(hookURNRewriterURN, NewHookURNRewriter)
}

func NewHookURNRewriter(name string, options spirit.Map) (rewriter spirit.URNRewriter, err error) {
	conf := HookURNRewriterConfig{}

	if err = options.ToObject(&conf); err != nil {
		return
	}

	conf.Delims.Left = strings.TrimSpace(conf.Delims.Left)
	conf.Delims.Right = strings.TrimSpace(conf.Delims.Right)

	tmpTmpls := make(map[string]*template.Template)
	if conf.Hooks != nil {
		for hookURN, urnTemplate := range conf.Hooks {
			if _, exist := tmpTmpls[hookURN]; exist {
				err = ErrHookURNRewriterTemplateDuplicate
				return
			}

			var tmpl *template.Template

			tmpl = template.New(hookURN).Option("missingkey=error")

			if conf.Delims.Left != "" && conf.Delims.Right != "" {
				tmpl = tmpl.Delims(conf.Delims.Left, conf.Delims.Right)
			}

			if tmpl, err = tmpl.Parse(urnTemplate); err != nil {
				return
			}

			tmpTmpls[hookURN] = tmpl
		}
	}

	rewriter = &HookURNRewriter{
		name:  name,
		conf:  conf,
		tmpls: tmpTmpls,
	}

	return
}

func (p *HookURNRewriter) Name() string {
	return p.name
}

func (p *HookURNRewriter) URN() string {
	return hookURNRewriterURN
}

func (p *HookURNRewriter) wholeMatchingRewrite(delivery spirit.Delivery) (err error) {

	originalURN := delivery.URN()
	originalURN = strings.Trim(originalURN, "|")

	if p.tmpls == nil || originalURN == "" {
		return
	}

	tmpURNs := []string{}

	var tmpl *template.Template
	var exist bool

	if tmpl, exist = p.tmpls[originalURN]; !exist {
		tmpURNs = append(tmpURNs, originalURN)
	} else {
		var buf bytes.Buffer
		if err = tmpl.Execute(&buf, map[string]interface{}{"urn": originalURN}); err != nil {
			return
		}
		tmpURNs = append(tmpURNs, buf.String())
	}

	newURN := strings.Join(tmpURNs, "|")

	if newURN != originalURN {
		spirit.Logger().WithField("actor", spirit.ActorURNRewriter).
			WithField("urn", hookURNRewriterURN).
			WithField("name", p.name).
			WithField("event", "rewrite delivery urn").
			WithField("mode", "whole matching").
			WithField("orignial_urn", delivery.URN()).
			WithField("new_urn", newURN).
			Debugln("rewrite delivery urn")
	}

	delivery.SetURN(newURN)

	return
}

func (p *HookURNRewriter) splitMatchRewrite(delivery spirit.Delivery) (err error) {
	originalURN := delivery.URN()
	originalURN = strings.Trim(originalURN, "|")

	if p.tmpls == nil || originalURN == "" {
		return
	}

	urns := strings.Split(originalURN, "|")

	tmpURNs := []string{}

	for _, urn := range urns {
		if urn != "" {

			var tmpl *template.Template
			var exist bool

			if tmpl, exist = p.tmpls[urn]; !exist {
				tmpURNs = append(tmpURNs, urn)
				continue
			}

			var buf bytes.Buffer
			if err = tmpl.Execute(&buf, map[string]interface{}{"urn": urn}); err != nil {
				return
			}
			tmpURNs = append(tmpURNs, buf.String())
		}
	}

	newURN := strings.Join(tmpURNs, "|")

	if newURN != originalURN {
		spirit.Logger().WithField("actor", spirit.ActorURNRewriter).
			WithField("urn", hookURNRewriterURN).
			WithField("name", p.name).
			WithField("event", "rewrite delivery urn").
			WithField("mode", "split matching").
			WithField("orignial_urn", delivery.URN()).
			WithField("new_urn", newURN).
			Debugln("rewrite delivery urn")
	}

	return
}

func (p *HookURNRewriter) Rewrite(delivery spirit.Delivery) (err error) {
	if p.conf.WholeMatching {
		return p.wholeMatchingRewrite(delivery)
	} else {
		return p.splitMatchRewrite(delivery)
	}
}
