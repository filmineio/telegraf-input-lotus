package lotus

import (
	"fmt"

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
	Daemon      *Daemon         `toml:"-"`
	Miner       *Miner          `toml:"-"`
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
	if s.DaemonAddr != "" {
		daemon, err := NewDaemon(s.DaemonAddr, s.DaemonToken)
		if err != nil {
			return err
		}
		s.Daemon = daemon
	}

	if s.MinerAddr != "" {
		miner, err := NewMiner(s.MinerAddr, s.MinerToken)
		if err != nil {
			return err
		}
		s.Miner = miner
	}

	return nil
}

func (s *LotusInput) Gather(acc telegraf.Accumulator) error {
	var daemonMetrics DaemonMetrics
	if s.Daemon != nil {
		daemonMetrics = s.Daemon.FetchMetrics()
	}

	var minerMetrics MinerMetrics
	if s.Miner != nil {
		minerMetrics = s.Miner.FetchMetrics()
	}

	measurements := map[string]interface{}{
		"epoch":          daemonMetrics.Status.SyncStatus.Epoch,
		"behind":         daemonMetrics.Status.SyncStatus.Behind,
		"messagePeers":   daemonMetrics.Status.PeerStatus.PeersToPublishMsgs,
		"blockPeers":     daemonMetrics.Status.PeerStatus.PeersToPublishBlocks,
		"marketDeals":    len(minerMetrics.MarketDeals),
		"retrievalDeals": len(minerMetrics.RetrievalDeals),
		"balance":        daemonMetrics.Balance}

	sectorsTotal := 0
	for sectorState, count := range minerMetrics.SectorSummary {
		measurements[fmt.Sprintf("sectors%s", sectorState)] = count
		sectorsTotal += count
	}
	measurements["sectorsTotal"] = sectorsTotal

	// TODO: Extract argument to struct
	acc.AddFields(lotusMeasurement, measurements,
		nil)

	return nil
}

func init() {
	inputs.Add(pluginName, func() telegraf.Input { return &LotusInput{} })
}
