package main

import (
	"context"
	gocore2 "github.com/core-coin/go-core"
	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/core/types"
	"github.com/core-coin/go-core/xcbclient"
	"github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb-client-go/api"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	delay      int
	dbWriteAPI api.WriteAPI
)

type GocoreInfo struct {
	GocoreServer     string
	ContractsCreated int64
	ContractCalls    int64
	XcbTransfers     int64
	BlockSize        float64
	LoadTime         float64
	TotalXcb         *big.Int
	CurrentBlock     *types.Block
	Sync             *gocore2.SyncProgress
	LastBlockUpdate  time.Time
	SugEnergyPrice   *big.Int
	PendingTx        uint
	NetworkId        *big.Int
}

func main() {
	id, err := strconv.Atoi(os.Getenv("NETWORK_ID"))
	if err != nil {
		panic(err)
	}
	common.DefaultNetworkID = common.NetworkID(id)

	token := os.Getenv("INFLUXDB_TOKEN")
	bucket := os.Getenv("INFLUXDB_BUCKET")
	org := os.Getenv("INFLUXDB_ORG")
	url := os.Getenv("INFLUXDB_URL")

	client := influxdb2.NewClient(url, token)
	defer client.Close()
	dbWriteAPI = client.WriteAPI(org, bucket)

	hosts := os.Getenv("GOCORE_HOSTS")
	delay, _ = strconv.Atoi(os.Getenv("DELAY"))
	if delay == 0 {
		delay = 1000
	}

	for _, host := range strings.Split(hosts, ",") {
		gocore := new(GocoreInfo)
		gocore.TotalXcb = big.NewInt(0)
		gocore.GocoreServer = host

		log.Printf("Connecting to Go-dial node: %v\n", gocore.GocoreServer)

		dial, err := xcbclient.Dial(gocore.GocoreServer)
		if err != nil {
			log.Println("FATAL ERR: cannot connect to", host)
			continue
		}
		defer dial.Close()

		gocore.CurrentBlock, err = dial.BlockByNumber(context.TODO(), nil)
		if err != nil {
			log.Println("FATAL ERR: cannot get latest block from", host)
			continue
		}

		go Routine(gocore, dial)
	}

	log.Printf("Gocore Exporter running with hosts: %s\n", hosts)
	var wait chan bool
	<-wait
}

func CalculateTotals(gocore *GocoreInfo) {
	gocore.TotalXcb = big.NewInt(0)
	gocore.ContractsCreated = 0
	gocore.XcbTransfers = 0
	for _, b := range gocore.CurrentBlock.Transactions() {

		if b.To() == nil {
			gocore.ContractsCreated++
		}

		if b.Value().Sign() == 1 {
			gocore.XcbTransfers++
		}
		gocore.TotalXcb.Add(gocore.TotalXcb, b.Value())
	}
	size := strings.Split(gocore.CurrentBlock.Size().String(), " ")
	gocore.BlockSize = stringToFloat(size[0]) * 1000
}

func Routine(gocore *GocoreInfo, dial *xcbclient.Client) {
	var lastBlock *types.Block
	ctx := context.Background()
	for {
		t1 := time.Now()
		var err error
		gocore.CurrentBlock, err = dial.BlockByNumber(ctx, nil)
		if err != nil {
			log.Printf("issue with reponse from gocore server: %v\n", gocore.CurrentBlock)
			time.Sleep(time.Duration(delay) * time.Millisecond)
			continue
		}
		gocore.SugEnergyPrice, err = dial.SuggestEnergyPrice(ctx)
		if err != nil {
			panic(err)
		}
		gocore.PendingTx, err = dial.PendingTransactionCount(ctx)
		if err != nil {
			panic(err)
		}
		gocore.NetworkId, err = dial.NetworkID(ctx)
		if err != nil {
			panic(err)
		}
		gocore.Sync, err = dial.SyncProgress(ctx)
		if err != nil {
			panic(err)
		}

		if lastBlock == nil || gocore.CurrentBlock.NumberU64() > lastBlock.NumberU64() {
			log.Printf("Received block #%v with %v transactions (%v)\n", gocore.CurrentBlock.NumberU64(), len(gocore.CurrentBlock.Transactions()), gocore.CurrentBlock.Hash().String())
			gocore.LastBlockUpdate = time.Now()
			gocore.LoadTime = time.Now().Sub(t1).Seconds()
		}

		CalculateTotals(gocore)

		writeToDB(gocore)

		lastBlock = gocore.CurrentBlock
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
}

func writeToDB(gocore *GocoreInfo) {
	block := gocore.CurrentBlock

	var tags = map[string]interface{}{}
	tags["gocore_block"] = block.NumberU64()

	tags["gocore_seconds_last_block"] = time.Now().Sub(gocore.LastBlockUpdate).Seconds()
	tags["gocore_block_transactions"] = len(block.Transactions())

	n, _ := strconv.ParseFloat(ToXcb(gocore.TotalXcb).Text('f', 2), 64)
	tags["gocore_block_value"] = n

	tags["gocore_block_energy_used"] = block.EnergyUsed()
	tags["gocore_block_energy_limit"] = block.EnergyLimit()
	tags["gocore_block_nonce"] = block.Nonce()
	tags["gocore_block_difficulty"] = block.Difficulty().Int64()
	tags["gocore_block_uncles"] = len(block.Uncles())
	tags["gocore_block_size_bytes"] = gocore.BlockSize
	tags["gocore_energy_price"] = gocore.SugEnergyPrice.Int64()
	tags["gocore_pending_transactions"] = gocore.PendingTx
	tags["gocore_network_id"] = gocore.NetworkId.Int64()
	tags["gocore_contracts_created"] = gocore.ContractsCreated
	tags["gocore_core_transfers"] = gocore.XcbTransfers
	tags["gocore_load_time"] = gocore.LoadTime

	if gocore.Sync != nil {
		tags["gocore_known_states"] = int(gocore.Sync.KnownStates)
		tags["gocore_highest_block"] = int(gocore.Sync.HighestBlock)
		tags["gocore_pulled_states"] = int(gocore.Sync.PulledStates)
	}

	dbWriteAPI.WritePoint(influxdb2.NewPoint(
		"gocore_node",
		map[string]string{"host": gocore.GocoreServer},
		tags,
		time.Now(),
	))
	dbWriteAPI.Flush()
}

// stringToFloat will simply convert a string to a float
func stringToFloat(s string) float64 {
	amount, _ := strconv.ParseFloat(s, 10)
	return amount
}

//
// CONVERTS ORE TO XCB
func ToXcb(o *big.Int) *big.Float {
	pul, val := big.NewFloat(0), big.NewFloat(0)
	val.SetInt(o)
	pul.Mul(big.NewFloat(0.000000000000000001), val)
	return pul
}
