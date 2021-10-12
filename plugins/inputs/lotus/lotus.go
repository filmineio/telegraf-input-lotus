package lotus

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	jsonrpc "github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/big"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
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
	daemonMetrics := fetchDaemonMetrics(s.DaemonAddr, s.DaemonToken)
	minerMetrics := fetchMinerMetrics(s.MinerAddr, s.MinerToken)

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

// TODO: Refactor this
func fetchDaemonMetrics(host string, key string) DaemonMetrics {
	authToken := key
	headers := http.Header{"Authorization": []string{"Bearer " + authToken}}
	addr := host

	// TODO: Persist connection in DaemonClient struct, don't open it on every call
	var daemonApi lotusapi.FullNodeStruct
	closer, err := jsonrpc.NewMergeClient(context.Background(), "ws://"+addr+"/rpc/v0", "Filecoin", []interface{}{&daemonApi.Internal, &daemonApi.CommonStruct.Internal}, headers)
	if err != nil {
		log.Printf("addr: %s, token: %s", host, key)
		log.Fatalf("connecting with lotus-daemon failed: %s", err)
	}
	defer closer()

	// TODO: What is the second argument?
	status, err := daemonApi.NodeStatus(context.Background(), true)
	if err != nil {
		log.Fatalf("calling daemon status: %s", err)
	}

	addresses, err := daemonApi.WalletList(context.Background())
	if err != nil {
		log.Fatalf("calling Wallet List %s", err)
	}

	// ⚠️ Filecoin uses it's own "big" package (it's not the stdlib big) and it's quite clunky
	totalBalance := big.Zero()
	for _, addr := range addresses {
		balance, err := daemonApi.WalletBalance(context.Background(), addr)
		if err != nil {
			log.Fatalf("calling Wallet Balance: %s", err)
		}

		totalBalance = big.Add(totalBalance, balance)
	}

	// TODO: Fetch granular balance information: vesting, available etc
	// ⚠️  Influx doesn't support big.Int so we have to risk it with float64 conversions (shouldn't be a problem)
	stringBalance := types.FIL(totalBalance).Unitless()
	floatBalance, err := strconv.ParseFloat(stringBalance, 64)
	if err != nil {
		log.Fatalf("parsing balance: %s", err)
	}

	return DaemonMetrics{
		Status:  status,
		Balance: floatBalance,
	}
}

func fetchMinerMetrics(minerAddr string, minerToken string) MinerMetrics {
	headers := http.Header{"Authorization": []string{"Bearer " + minerToken}}

	// TODO: Persist connection in MinerClient struct, don't open it on every call
	var minerApi lotusapi.StorageMinerStruct
	closer, err := jsonrpc.NewMergeClient(context.Background(), "ws://"+minerAddr+"/rpc/v0", "Filecoin", []interface{}{&minerApi.Internal, &minerApi.CommonStruct.Internal}, headers)
	if err != nil {
		log.Printf("addr: %s, token: %s", minerAddr, minerToken)
		log.Fatalf("connecting with lotus-miner failed: %s", err)
	}
	defer closer()

	sectorSummary, err := minerApi.SectorsSummary(context.Background())
	if err != nil {
		log.Fatalf("calling sectors summary: %s", err)
	}

	marketDeals, err := minerApi.MarketListDeals(context.Background())
	if err != nil {
		log.Fatalf("callung MarketListDeals: %s", err)
	}

	return MinerMetrics{
		SectorSummary: sectorSummary,
		MarketDeals:   marketDeals,
	}
}

type MinerMetrics struct {
	SectorSummary  map[lotusapi.SectorState]int
	MarketDeals    []lotusapi.MarketDeal
	RetrievalDeals []retrievalmarket.ProviderDealState
}

type DaemonMetrics struct {
	Status  lotusapi.NodeStatus
	Balance float64
}

func init() {
	inputs.Add(pluginName, func() telegraf.Input { return &LotusInput{} })
}
