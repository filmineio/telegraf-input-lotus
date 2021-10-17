package lotus

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	jsonrpc "github.com/filecoin-project/go-jsonrpc"
	lotusapi "github.com/filecoin-project/lotus/api"
)

type Miner struct {
	api lotusapi.StorageMinerStruct
}

func (m Miner) FetchMetrics() MinerMetrics {
	sectorSummary, err := m.api.SectorsSummary(context.Background())
	if err != nil {
		log.Fatalf("calling sectors summary: %s", err)
	}

	marketDeals, err := m.api.MarketListDeals(context.Background())
	if err != nil {
		log.Fatalf("callung MarketListDeals: %s", err)
	}

	return MinerMetrics{
		SectorSummary: sectorSummary,
		MarketDeals:   marketDeals,
	}
}

func NewMiner(addr string, token string) (*Miner, error) {
	if addr == "" {
		return nil, errors.New("addr can't be an empty string")
	}
	if token == "" {
		return nil, errors.New("token can't be an empty string")
	}

	headers := http.Header{"Authorization": []string{"Bearer " + token}}

	var minerApi lotusapi.StorageMinerStruct
	_, err := jsonrpc.NewMergeClient(context.Background(), "ws://"+addr+"/rpc/v0", "Filecoin", []interface{}{&minerApi.Internal, &minerApi.CommonStruct.Internal}, headers)
	if err != nil {
		log.Printf("addr: %s, token: %s", addr, token)
		log.Fatalf("connecting with lotus-miner failed: %s", err)
	}

	return &Miner{
		api: minerApi,
	}, nil
}

type MinerMetrics struct {
	SectorSummary  map[lotusapi.SectorState]int
	MarketDeals    []lotusapi.MarketDeal
	RetrievalDeals []retrievalmarket.ProviderDealState
}
