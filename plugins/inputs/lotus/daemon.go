package lotus

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	jsonrpc "github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/big"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
)

type Daemon struct {
	api lotusapi.FullNodeStruct
}

func (d Daemon) FetchMetrics() DaemonMetrics {
	// TODO: What is the second argument?
	status, err := d.api.NodeStatus(context.Background(), true)
	if err != nil {
		log.Fatalf("calling daemon status: %s", err)
	}

	addresses, err := d.api.WalletList(context.Background())
	if err != nil {
		log.Fatalf("calling Wallet List %s", err)
	}

	// ⚠️ Filecoin uses it's own "big" package (it's not the stdlib big)
	totalBalance := big.Zero()
	for _, addr := range addresses {
		balance, err := d.api.WalletBalance(context.Background(), addr)
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

func NewDaemon(addr string, token string) (*Daemon, error) {
	if addr == "" {
		return nil, errors.New("addr can't be an empty string")
	}
	if token == "" {
		return nil, errors.New("token can't be an empty string")
	}

	headers := http.Header{"Authorization": []string{"Bearer " + token}}

	var daemonApi lotusapi.FullNodeStruct
	_, err := jsonrpc.NewMergeClient(context.Background(), "ws://"+addr+"/rpc/v0", "Filecoin", []interface{}{&daemonApi.Internal, &daemonApi.CommonStruct.Internal}, headers)
	if err != nil {
		log.Printf("addr: %s", addr)
		log.Fatalf("connecting with lotus-daemon failed: %s", err)
	}

	return &Daemon{
		api: daemonApi,
	}, nil
}

type DaemonMetrics struct {
	Status  lotusapi.NodeStatus
	Balance float64
}
