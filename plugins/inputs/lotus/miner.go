package lotus

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	jsonrpc "github.com/filecoin-project/go-jsonrpc"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/extern/sector-storage/storiface"
	"github.com/google/uuid"
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
		log.Fatalf("calling MarketListDeals: %s", err)
	}

	workerStats, err := m.api.WorkerStats(context.Background())
	if err != nil {
		log.Fatalf("calling WorkerStats: %s", err)
	}

	workerJobs, err := m.api.WorkerJobs(context.Background())
	if err != nil {
		log.Fatalf("calling WorkerJobs: %s", err)
	}

	return MinerMetrics{
		SectorSummary: sectorSummary,
		MarketDeals:   marketDeals,
		WorkerStats:   workerStats,
		WorkerJobs:    workerJobs,
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
	WorkerJobs     map[uuid.UUID][]storiface.WorkerJob
	WorkerStats    map[uuid.UUID]storiface.WorkerStats
	SectorSummary  map[lotusapi.SectorState]int
	MarketDeals    []lotusapi.MarketDeal
	RetrievalDeals []retrievalmarket.ProviderDealState
}