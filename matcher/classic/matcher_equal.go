package classic

import (
	"github.com/gogap/spirit"
)

const (
	equalMatcherURN = "urn:spirit:matcher:label:equal"
)

type EqualLabelMatcher struct {
}

func init() {
	spirit.RegisterLabelMatcher(equalMatcherURN, NewContainsLabelMatcher)
}

func NewEqualLabelMatcher(config spirit.Map) (matcher spirit.LabelMatcher, err error) {
	return &EqualLabelMatcher{}, nil
}

func (p *EqualLabelMatcher) Match(la spirit.Labels, lb spirit.Labels) bool {
	if la == nil && lb == nil {
		return true
	}

	if len(la) != len(lb) {
		return false
	}

	equalCount := 0
	for ka, va := range la {
		if vb, exist := lb[ka]; exist {
			if va == vb {
				equalCount += 1
			}
		}
	}

	if equalCount == len(la) {
		return true
	}

	return false
}
