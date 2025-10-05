package actions

type ActionBinding struct {
	On []string `toml:"on"`
	Do Action   `toml:"do"`
}

func (ab *ActionBinding) UnmarshalTOML(data any) error {
	switch value := data.(type) {
	case map[string]interface{}:
		if on, ok := value["on"]; ok {
			switch v := on.(type) {
			case string:
				ab.On = []string{v}
			case []interface{}:
				ab.On = []string{}
				for _, item := range v {
					ab.On = append(ab.On, item.(string))
				}
			}
		}
		if doAction, ok := value["do"]; ok {
			return ab.Do.UnmarshalTOML(doAction)
		}
	}
	return nil
}

type ActionMap struct {
	Key      string
	Bindings []ActionBinding
}

func (am *ActionMap) UnmarshalTOML(data any) error {
	switch value := data.(type) {
	case []interface{}:
		am.Bindings = []ActionBinding{}
		for _, v := range value {
			binding := ActionBinding{}
			if err := binding.UnmarshalTOML(v); err != nil {
				return err
			}
			am.Bindings = append(am.Bindings, binding)
		}
	}
	return nil
}

func (am ActionMap) Get(key string) (*ActionBinding, bool) {
	for i := range am.Bindings {
		binding := &am.Bindings[i]
		for _, on := range binding.On {
			if on == key {
				return binding, true
			}
		}
	}
	return nil, false
}
