package spirit

type ActorConfig struct {
	Name    string  `json:"name"`
	URN     string  `json:"urn"`
	Options Options `json:"options"`
}

type ComposeReceiverConfig struct {
	Name       string `json:"name"`
	Translator string `json:"translator"`
	Reader     string `json:"reader"`
}

type ComposeSenderConfig struct {
	Name       string `json:"name"`
	Translator string `json:"translator"`
	Writer     string `json:"writer"`
}

type ComposeInboxConfig struct {
	Name      string                  `json:"name"`
	Receivers []ComposeReceiverConfig `json:"receivers"`
}

type ComposeOutboxConfig struct {
	Name    string                `json:"name"`
	Senders []ComposeSenderConfig `json:"senders"`
}

type ComposeLabelMatchConfig struct {
	Component string `json:"component"`
	Outbox    string `json:"outbox"`
}

type ComposeRouterConfig struct {
	Name          string                  `json:"name"`
	LabelMatchers ComposeLabelMatchConfig `json:"label_matchers"`
	Components    []string                `json:"components"`
	Inboxes       []ComposeInboxConfig    `json:"inboxes"`
	Outboxes      []ComposeOutboxConfig   `json:"outboxes"`
}

type SpiritConfig struct {
	Readers           []ActorConfig `json:"readers"`
	Writers           []ActorConfig `json:"writers"`
	InputTranslators  []ActorConfig `json:"input_translators"`
	OutputTranslators []ActorConfig `json:"output_translators"`
	Inboxes           []ActorConfig `json:"inboxes"`
	Outboxes          []ActorConfig `json:"outboxes"`
	Receivers         []ActorConfig `json:"receivers"`
	Senders           []ActorConfig `json:"senders"`
	Routers           []ActorConfig `json:"routers"`
	Components        []ActorConfig `json:"components"`
	LabelMatchers     []ActorConfig `json:"label_matchers"`

	Compose []ComposeRouterConfig `json:"compose"`
}

func (p *SpiritConfig) Validate() (err error) {
	actorNames := map[string]bool{}

	if p.Readers != nil {
		dupCheck := map[string]bool{}
		for _, actor := range p.Readers {
			actorNames["reader:"+actor.Name] = true
			if _, exist := newReaderFuncs[actor.URN]; !exist {
				err = ErrReaderURNNotExist
				return
			}

			if _, exist := dupCheck[actor.Name]; exist {
				err = ErrReaderNameDuplicate
				return
			} else {
				dupCheck[actor.Name] = true
			}
		}
	}

	if p.Writers != nil {
		dupCheck := map[string]bool{}
		for _, actor := range p.Writers {
			actorNames["writer:"+actor.Name] = true
			if _, exist := newWriterFuncs[actor.URN]; !exist {
				err = ErrWriterURNNotExist
				return
			}

			if _, exist := dupCheck[actor.Name]; exist {
				err = ErrWriterNameDuplicate
				return
			} else {
				dupCheck[actor.Name] = true
			}
		}
	}

	if p.InputTranslators != nil {
		dupCheck := map[string]bool{}
		for _, actor := range p.InputTranslators {
			actorNames["input_translator:"+actor.Name] = true
			if _, exist := newInputTranslatorFuncs[actor.URN]; !exist {
				err = ErrInputTranslatorURNNotExist
				return
			}

			if _, exist := dupCheck[actor.Name]; exist {
				err = ErrInputTranslatorNameDuplicate
				return
			} else {
				dupCheck[actor.Name] = true
			}
		}
	}

	if p.OutputTranslators != nil {
		dupCheck := map[string]bool{}
		for _, actor := range p.OutputTranslators {
			actorNames["output_translator:"+actor.Name] = true
			if _, exist := newOutputTranslatorFuncs[actor.URN]; !exist {
				err = ErrOutputTranslatorURNNotExist
				return
			}

			if _, exist := dupCheck[actor.Name]; exist {
				err = ErrOutputTranslatorNameDuplicate
				return
			} else {
				dupCheck[actor.Name] = true
			}
		}
	}

	if p.Inboxes != nil {
		dupCheck := map[string]bool{}
		for _, actor := range p.Inboxes {
			actorNames["inbox:"+actor.Name] = true
			if _, exist := newInboxFuncs[actor.URN]; !exist {
				err = ErrInboxURNNotExist
				return
			}

			if _, exist := dupCheck[actor.Name]; exist {
				err = ErrInboxNameDuplicate
				return
			} else {
				dupCheck[actor.Name] = true
			}
		}
	}

	if p.Outboxes != nil {
		dupCheck := map[string]bool{}
		for _, actor := range p.Outboxes {
			actorNames["outbox:"+actor.Name] = true
			if _, exist := newOutboxFuncs[actor.URN]; !exist {
				err = ErrOutboxURNNotExist
				return
			}

			if _, exist := dupCheck[actor.Name]; exist {
				err = ErrOutboxNameDuplicate
				return
			} else {
				dupCheck[actor.Name] = true
			}
		}
	}

	if p.Routers != nil {
		dupCheck := map[string]bool{}
		for _, actor := range p.Routers {
			actorNames["router:"+actor.Name] = true
			if _, exist := newRouterFuncs[actor.URN]; !exist {
				err = ErrRouterURNNotExist
				return
			}

			if _, exist := dupCheck[actor.Name]; exist {
				err = ErrRouterNameDuplicate
				return
			} else {
				dupCheck[actor.Name] = true
			}
		}
	}

	if p.Components != nil {
		for _, actor := range p.Components {
			if _, exist := newComponentFuncs[actor.URN]; !exist {
				err = ErrComponentURNNotExist
				return
			}
		}
	}

	if p.Receivers != nil {
		dupCheck := map[string]bool{}
		for _, actor := range p.Receivers {
			actorNames["receiver:"+actor.Name] = true
			if _, exist := newReceiverFuncs[actor.URN]; !exist {
				err = ErrReceiverURNNotExist
				return
			}

			if _, exist := dupCheck[actor.Name]; exist {
				err = ErrReceiverNameDuplicate
				return
			} else {
				dupCheck[actor.Name] = true
			}
		}
	}

	if p.Senders != nil {
		dupCheck := map[string]bool{}
		for _, actor := range p.Senders {
			actorNames["sender:"+actor.Name] = true
			if _, exist := newSenderFuncs[actor.URN]; !exist {
				err = ErrSenderURNNotExist
				return
			}

			if _, exist := dupCheck[actor.Name]; exist {
				err = ErrSenderNameDuplicate
				return
			} else {
				dupCheck[actor.Name] = true
			}
		}
	}

	if p.LabelMatchers != nil {
		dupCheck := map[string]bool{}
		for _, actor := range p.LabelMatchers {
			actorNames["label_matchers:"+actor.Name] = true
			if _, exist := newLabelMatcherFuncs[actor.URN]; !exist {
				err = ErrLabelMatcherURNNotExist
				return
			}

			if _, exist := dupCheck[actor.Name]; exist {
				err = ErrLabelMatcherNameDuplicate
				return
			} else {
				dupCheck[actor.Name] = true
			}
		}
	}

	for _, router := range p.Compose {
		for _, compName := range router.Components {
			if _, exist := actorNames["component:"+compName]; !exist {
				err = ErrActorNotExist
				return
			}
		}

		for _, inbox := range router.Inboxes {
			if _, exist := actorNames["inbox:"+inbox.Name]; !exist {
				err = ErrActorNotExist
				return
			}

			for _, receiver := range inbox.Receivers {
				if _, exist := actorNames["receiver:"+receiver.Name]; !exist {
					err = ErrActorNotExist
					return
				}

				if _, exist := actorNames["input_translator:"+receiver.Translator]; !exist {
					err = ErrActorNotExist
					return
				}

				if _, exist := actorNames["reader:"+receiver.Reader]; !exist {
					err = ErrActorNotExist
					return
				}
			}
		}

		if _, exist := actorNames["label_matchers:"+router.LabelMatchers.Component]; !exist {
			err = ErrActorNotExist
			return
		}

		if _, exist := actorNames["label_matchers:"+router.LabelMatchers.Outbox]; !exist {
			err = ErrActorNotExist
			return
		}

		for _, outbox := range router.Outboxes {
			if _, exist := actorNames["outbox:"+outbox.Name]; !exist {
				err = ErrActorNotExist
				return
			}

			for _, sender := range outbox.Senders {
				if _, exist := actorNames["sender:"+sender.Name]; !exist {
					err = ErrActorNotExist
					return
				}

				if _, exist := actorNames["output_translator:"+sender.Translator]; !exist {
					err = ErrActorNotExist
					return
				}

				if _, exist := actorNames["reader:"+sender.Writer]; !exist {
					err = ErrActorNotExist
					return
				}
			}
		}
	}

	return
}
