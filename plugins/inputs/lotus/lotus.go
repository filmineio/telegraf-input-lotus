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
	DaemonAddr  string          `toml:"daemonAddr"`
	DaemonToken string          `toml:"daemonToken"`
	MinerAddr   string          `toml:"minerAddr"`
	MinerToken  string          `toml:"minerToken"`
	Log         telegraf.Logger `toml:"-"`
}

func (s *LotusInput) Description() string {
	return "Stream lotus-daemon and lotus-miner metrics"
}

func (s *LotusInput) SampleConfig() string {

	return `
  ## Lotus daemon listen address
  daemonAddr = 127.0.0.1:1234
	## Lotus daemon API token (example)
	daemonToken = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIl19.aneYo3I_Ts45E36uBcLNNK61q2aKj3p462fByqnam1s
  ## Lotus miner listen address
  minerAddr = 127.0.0.1:1234
	## Lotus miner API token (example)
	minerToken = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIl19.aneYo3I_Ts45E36uBcLNNK61q2aKj3p462fByqnam1s
`
}

func (s *LotusInput) Init() error {
	if s.DaemonAddr == "" {
		s.DaemonAddr = "127.0.0.1:1234"
	}

	if s.MinerAddr == "" {
		s.MinerAddr = "127.0.0.1:2345"
	}

	// TODO: Initialize the Daemon and Miner clients, check connection

	return nil
}

func (s *LotusInput) Gather(acc telegraf.Accumulator) error {
	daemonStatus := fetchLotusInfo(s.DaemonAddr, s.DaemonToken)

	// TODO: Extract argument to struct
	acc.AddFields(lotusMeasurement, map[string]interface{}{
		"epoch":        daemonStatus.SyncStatus.Epoch,
		"behind":       daemonStatus.SyncStatus.Behind,
		"messagePeers": daemonStatus.PeerStatus.PeersToPublishMsgs,
		"blockPeers":   daemonStatus.PeerStatus.PeersToPublishBlocks},
		nil)

	return nil
}

func fetchLotusInfo(host string, key string) lotusapi.NodeStatus {
	authToken := key
	headers := http.Header{"Authorization": []string{"Bearer " + authToken}}
	addr := host

	// TODO: Persist connection in DaemonClient struct, don't open it on every call
	var daemonApi lotusapi.FullNodeStruct
	closer, err := jsonrpc.NewMergeClient(context.Background(), "ws://"+addr+"/rpc/v0", "Filecoin", []interface{}{&daemonApi.Internal, &daemonApi.CommonStruct.Internal}, headers)
	if err != nil {
		log.Printf("addr: %s, token: %s", host, key)
		log.Fatalf("connecting with lotus daemon failed: %s", err)
	}
	defer closer()

	// TODO: What is the second argument?
	status, err := daemonApi.NodeStatus(context.Background(), true)
	if err != nil {
		log.Fatalf("calling daemon status: %s", err)
	}

	return status
}

// Miner info
// Miner Balance:    34762.999 FIL
//       PreCommit:  0
//       Pledge:     2 aFIL
//       Vesting:    26072.249 FIL
//       Available:  8690.75 FIL
// Market Balance:   0
//        Locked:    0
//        Available: 0
// Worker Balance:   50000000 FIL
// Total Spendable:  50008690.75 FIL

// Sectors:
// 	Total: 2
// 	Proving: 2

// Storage Deals: 0, 0 B

// Retrieval Deals (complete): 0, 0 B

func init() {
	inputs.Add(pluginName, func() telegraf.Input { return &LotusInput{} })
}
