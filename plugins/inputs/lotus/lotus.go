package lotus

import (
	"context"
	"log"
	"net/http"

	jsonrpc "github.com/filecoin-project/go-jsonrpc"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const pluginName string = "telegraf-input-lotus"
const lotusMeasurement string = "lotus"

type LotusInput struct {
	Ok   bool            `toml:"ok"`
	Host string          `toml:"host"`
	Key  string          `toml:"key"`
	Log  telegraf.Logger `toml:"-"`
}

func (s *LotusInput) Description() string {
	return "a demo plugin"
}

func (s *LotusInput) SampleConfig() string {
	return `
  ## Lotus API connection string
  host = 127.0.0.1:1234
	## Lotus API key
	key = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIl19.aneYo3I_Ts45E36uBcLNNK61q2aKj3p462fByqnam1s
`
}

func (s *LotusInput) Init() error {
	if s.Host == "" {
		s.Host = "127.0.0.1:1234"
	}

	if s.Key == "" {
		s.Key = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIl19.aneYo3I_Ts45E36uBcLNNK61q2aKj3p462fByqnam1s"
	}

	return nil
}

func (s *LotusInput) Gather(acc telegraf.Accumulator) error {
	tipset := fetchLotusInfo(s.Host, s.Key)

	acc.AddFields(lotusMeasurement, map[string]interface{}{"host": s.Host, "key": s.Key, "tipset": tipset}, nil)

	return nil
}

func fetchLotusInfo(host string, key string) string {
	authToken := key
	headers := http.Header{"Authorization": []string{"Bearer " + authToken}}
	addr := host

	var api lotusapi.FullNodeStruct
	closer, err := jsonrpc.NewMergeClient(context.Background(), "ws://"+addr+"/rpc/v0", "Filecoin", []interface{}{&api.Internal, &api.CommonStruct.Internal}, headers)
	if err != nil {
		log.Fatalf("connecting with lotus failed: %s", err)
	}
	defer closer()

	// Now you can call any API you're interested in.
	tipset, err := api.ChainHead(context.Background())
	if err != nil {
		log.Fatalf("calling chain head: %s", err)
	}
	// fmt.Printf("Current chain head is: %s", tipset.String())
	return tipset.String()
}

func init() {
	inputs.Add(pluginName, func() telegraf.Input { return &LotusInput{} })
}
