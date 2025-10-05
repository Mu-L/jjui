package context

import (
	"strings"

	"github.com/idursun/jjui/internal/config"
	"github.com/idursun/jjui/internal/jj"
	"github.com/idursun/jjui/internal/ui/view"
)

type MainContext struct {
	CommandRunner
	Location      string
	JJConfig      *config.JJConfig
	DefaultRevset string
	CurrentRevset string
	Histories     *config.Histories
	Router        *view.Router
	variables     map[string]string
}

func (ctx *MainContext) Set(key string, value string) {
	ctx.variables[key] = value
}

func (ctx *MainContext) GetVariables() map[string]string {
	return ctx.variables
}

func (ctx *MainContext) ReplaceWithVariables(input string) string {
	for k, v := range ctx.variables {
		input = strings.ReplaceAll(input, k, v)
	}

	return input
}

func NewAppContext(location string) *MainContext {
	m := &MainContext{
		CommandRunner: &MainCommandRunner{
			Location: location,
		},
		Location:  location,
		Histories: config.NewHistories(),
		variables: make(map[string]string),
		Router:    view.NewRouter(view.ScopeRevisions),
	}

	m.JJConfig = &config.JJConfig{}
	if output, err := m.RunCommandImmediate(jj.ConfigListAll()); err == nil {
		m.JJConfig, _ = config.DefaultConfig(output)
	}
	return m
}
