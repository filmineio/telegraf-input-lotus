package lotus

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	jsonrpc "github.com/filecoin-project/go-jsonrpc"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/extern/sector-storage/fsutil"
	"github.com/filecoin-project/lotus/extern/sector-storage/stores"
	"github.com/filecoin-project/lotus/extern/sector-storage/storiface"
	"github.com/google/uuid"
)

type Miner struct {
	api lotusapi.StorageMinerStruct
}

func arrayToString(a []int, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
	//return strings.Trim(strings.Join(strings.Split(fmt.Sprint(a), " "), delim), "[]")
	//return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(a)), delim), "[]")
}

func (m Miner) FetchMetrics() MinerMetrics {
	sectorSummary, err := m.api.SectorsSummary(context.Background())
	if err != nil {
		log.Printf("calling sectors summary: %s", err)
	}

	marketDeals, err := m.api.MarketListDeals(context.Background())
	if err != nil {
		log.Printf("calling MarketListDeals: %s", err)
	}

	workerStats, err := m.api.WorkerStats(context.Background())
	if err != nil {
		log.Printf("calling WorkerStats: %s", err)
	}

	workerJobs, err := m.api.WorkerJobs(context.Background())
	if err != nil {
		log.Printf("calling WorkerJobs: %s", err)
	}

	storageList, err := m.api.StorageList(context.Background())
	if err != nil {
		log.Printf("calling StorageList: %s", err)
	}
	storageStats := map[stores.ID]fsutil.FsStat{}
	storageInfos := map[stores.ID]stores.StorageInfo{}
	storageSectors := map[stores.ID]string{}
	for id, declArray := range storageList {
		sectorsArray := []int{}
		for _, decl := range declArray {
			sectorsArray = append(sectorsArray, int(decl.SectorID.Number))
		}
		storageSectors[id] = arrayToString(sectorsArray, ",")
		stat, err := m.api.StorageStat(context.Background(), id)
		if err != nil {
			log.Printf("calling StorageStat: %s", err)
		}
		storageStats[id] = stat
		info, err := m.api.StorageInfo(context.Background(), id)
		if err != nil {
			log.Printf("calling StorageInfo: %s", err)
		}
		storageInfos[id] = info
	}
	return MinerMetrics{
		SectorSummary:  sectorSummary,
		MarketDeals:    marketDeals,
		WorkerStats:    workerStats,
		WorkerJobs:     workerJobs,
		StorageStats:   storageStats,
		StorageInfos:   storageInfos,
		StorageSectors: storageSectors,
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
	StorageStats   map[stores.ID]fsutil.FsStat
	StorageInfos   map[stores.ID]stores.StorageInfo
	SectorSummary  map[lotusapi.SectorState]int
	StorageSectors map[stores.ID]string
	MarketDeals    []lotusapi.MarketDeal
	RetrievalDeals []retrievalmarket.ProviderDealState
}
