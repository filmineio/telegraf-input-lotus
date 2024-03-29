package lotus

import (
	"context"
	"errors"
	"log"
	"net/http"

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

func (m Miner) FetchMetrics() MinerMetrics {
	sectorSummary, err := m.api.SectorsSummary(context.Background())
	if err != nil {
		log.Printf("calling sectors summary: %s", err)
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
	for id := range storageList {
		stat, err := m.api.StorageStat(context.Background(), id)
		if err != nil {
			log.Printf("calling StorageStat: %s", err)
		}
		storageStats[id] = stat
	}
	return MinerMetrics{
		SectorSummary: sectorSummary,
		WorkerStats:   workerStats,
		WorkerJobs:    workerJobs,
		StorageStats:  storageStats,
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
	WorkerJobs    map[uuid.UUID][]storiface.WorkerJob
	WorkerStats   map[uuid.UUID]storiface.WorkerStats
	StorageStats  map[stores.ID]fsutil.FsStat
	SectorSummary map[lotusapi.SectorState]int
}
