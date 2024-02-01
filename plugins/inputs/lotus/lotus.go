package lotus

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const pluginName string = "telegraf-input-lotus"
const lotusMeasurement string = "lotus"
const lotusSealingWorkers string = "lotus_sealing_workers"
const lotusSealingJobs string = "lotus_sealing_jobs"
const lotusStorageStats string = "lotus_storage_stats"

type LotusInput struct {
	DaemonAddr  string          `toml:"daemonAddr"`
	DaemonToken string          `toml:"daemonToken"`
	ApiVersion  string          `toml:"daemonApiVersion"`
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
	## Lotus daemon API Version (default: v0)
	daemonApiVersion = "v1"
  ## Lotus miner listen address
  minerAddr = 127.0.0.1:1234
	## Lotus miner API token (example)
	minerToken = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIl19.aneYo3I_Ts45E36uBcLNNK61q2aKj3p462fByqnam1s
`
}

func (s *LotusInput) Init() error {
	if s.DaemonAddr != "" {
		daemon, err := NewDaemon(s.DaemonAddr, s.DaemonToken, s.ApiVersion)
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
		"epoch":        daemonMetrics.Status.SyncStatus.Epoch,
		"behind":       daemonMetrics.Status.SyncStatus.Behind,
		"messagePeers": daemonMetrics.Status.PeerStatus.PeersToPublishMsgs,
		"blockPeers":   daemonMetrics.Status.PeerStatus.PeersToPublishBlocks,
		"balance":      daemonMetrics.Balance}

	sectorsTotal := 0
	for sectorState, count := range minerMetrics.SectorSummary {
		measurements[fmt.Sprintf("sectors%s", sectorState)] = count
		sectorsTotal += count
	}
	measurements["sectorsTotal"] = sectorsTotal

	// TODO: Extract argument to struct
	// add tags worker_host:key.value.info.Hostname
	// add tags worker_id:key.String()
	acc.AddFields(lotusMeasurement, measurements, nil)
	workerIDtoNameMap := map[string]string{}
	for key, value := range minerMetrics.WorkerStats {
		workerMeasurements := map[string]interface{}{}
		tags := map[string]string{}
		tags["worker_host"] = value.Info.Hostname
		tags["worker_id"] = key.String()
		workerIDtoNameMap[key.String()] = value.Info.Hostname
		workerMeasurements["cpu_use"] = value.CpuUse
		workerMeasurements["mem_physical"] = value.Info.Resources.MemPhysical
		workerMeasurements["mem_used"] = value.Info.Resources.MemUsed
		workerMeasurements["mem_swap_used"] = value.Info.Resources.MemSwapUsed
		workerMeasurements["gpu_used"] = value.GpuUsed
		acc.AddFields(lotusSealingWorkers, workerMeasurements, tags)
	}

	for key, value := range minerMetrics.WorkerJobs {
		for _, job := range value {
			jobMeasurements := map[string]interface{}{}
			tags := map[string]string{}
			tags["worker_host"] = workerIDtoNameMap[key.String()]
			tags["worker_id"] = key.String()
			tags["job_id"] = job.ID.ID.String()
			tags["sector"] = job.Sector.Number.String()
			tags["miner_id"] = job.Sector.Miner.String()
			jobMeasurements["run_wait"] = job.RunWait
			jobMeasurements["start"] = job.Start.String()
			jobMeasurements["task"] = job.Task.Short()
			acc.AddFields(lotusSealingJobs, jobMeasurements, tags)
		}

	}

	for key, stat := range minerMetrics.StorageStats {
		storageMeasurments := map[string]interface{}{}
		tags := map[string]string{}
		tags["storage_id"] = string(key)
		storageMeasurments["available"] = stat.Available
		storageMeasurments["capacity"] = stat.Capacity
		storageMeasurments["fs_available"] = stat.FSAvailable
		storageMeasurments["max"] = stat.Max
		storageMeasurments["reserved"] = stat.Reserved
		storageMeasurments["used"] = stat.Used
		acc.AddFields(lotusStorageStats, storageMeasurments, map[string]string{})
	}
	return nil
}

func init() {
	inputs.Add(pluginName, func() telegraf.Input { return &LotusInput{} })
}
