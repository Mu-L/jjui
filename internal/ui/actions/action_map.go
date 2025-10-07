package actions

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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

func (am *ActionMap) Get(key string) (*ActionBinding, bool) {
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

func (am *ActionMap) GetMatch(previous []string, key string) []*ActionBinding {
	var matches []*ActionBinding
	for i := range am.Bindings {
		binding := &am.Bindings[i]
		if len(binding.On) <= len(previous) {
			continue
		}
		for i := 0; i < len(previous); i++ {
			if binding.On[i] != previous[i] {
				goto next
			}
		}
		if binding.On[len(previous)] != key {
			goto next
		}
		matches = append(matches, binding)
	next:
	}
	return matches
}

var border = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).PaddingLeft(1).PaddingRight(1)

// RenderAvailableActions renders the available actions
// This is temporary until a better UI is implemented
func RenderAvailableActions(bindings []*ActionBinding, textStyle lipgloss.Style, shortCutStyle lipgloss.Style) string {
	lines := make([]string, len(bindings))
	for i, binding := range bindings {
		keys := strings.Join(binding.On, " ")
		lines[i] = lipgloss.JoinHorizontal(0, shortCutStyle.Render(keys), textStyle.Render(" ", binding.Do.GetDesc()))
	}
	content := lipgloss.JoinVertical(0, lines...)
	return border.Render(content)
}
