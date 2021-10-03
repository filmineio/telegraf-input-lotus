package lotus

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Simple struct {
	Ok  bool            `toml:"ok"`
	Log telegraf.Logger `toml:"-"`
}

func (s *Simple) Description() string {
	return "a demo plugin"
}

func (s *Simple) SampleConfig() string {
	return `
  ## Indicate if everything is fine
  ok = true
`
}

// Init is for setup, and validating config.
func (s *Simple) Init() error {
	return nil
}

func (s *Simple) Gather(acc telegraf.Accumulator) error {
	if s.Ok {
		acc.AddFields("lotus", map[string]interface{}{"value": "pretty good"}, nil)
	} else {
		acc.AddFields("lotus", map[string]interface{}{"value": "not great"}, nil)
	}

	return nil
}

func init() {
	inputs.Add("telegraf-lotus", func() telegraf.Input { return &Simple{} })
}
