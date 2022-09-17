package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/avvvet/oxygen-wallet/internal/app"
	"github.com/avvvet/oxygen-wallet/internal/pkg/net"
	"github.com/common-nighthawk/go-figure"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

var (
	ledger_path = "./ledger/wallet"
	logger, _   = zap.NewDevelopment()
)

type addrList []multiaddr.Multiaddr

type Config struct {
	HttpPort       uint
	PeerPort       uint
	ProtocolID     string
	Rendezvous     string
	Seed           int64
	DiscoveryPeers addrList
}

func main() {
	art()
	config := Config{}

	flag.StringVar(&config.Rendezvous, "meet", "otwo", "peer joining place")
	flag.Int64Var(&config.Seed, "seed", 0, "0 is for random PeerID")
	flag.Var(&config.DiscoveryPeers, "peer", "Perr address for peer discovery")
	flag.StringVar(&config.ProtocolID, "protocolid", "/p2p/otwo", "")
	flag.UintVar(&config.HttpPort, "httpPort", 0, "http port for otwo wallet")
	flag.UintVar(&config.PeerPort, "peerPort", 0, "port for otwo wallet peer that connects to the network")
	flag.Parse()

	app.NewDir(ledger_path)

	_, err := app.InitWalletLedger(ledger_path)
	if err != nil {
		logger.Sugar().Warn("critical error in wallet address")
	}

	ctx, cancel := context.WithCancel(context.Background())

	h, err := net.NewHost(ctx, config.Seed, int(config.PeerPort))
	if err != nil {
		log.Fatal(err)
	}

	logger.Sugar().Info("peer network address")
	for _, addr := range h.Addrs() {
		log.Printf("  %s/p2p/%s", addr, h.ID().Pretty())
	}
	fmt.Println("")

	/*create gossipSub */
	_, err = net.InitPubSub(ctx, h, "otwo-wallet")
	if err != err {
		logger.Sugar().Fatal("Error: creating pubsub", err)
	}

	dht, err := net.NewDHT(ctx, h, config.DiscoveryPeers)
	if err != nil {
		log.Fatal(err)
	}

	go net.Discover(ctx, h, dht, config.Rendezvous)

	/*http server */
	http := app.NewWeb(config.HttpPort)
	go http.Run()

	run(h, cancel)
}

func run(h host.Host, cancel func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	<-c
	fmt.Printf("\r👋️ stopped...\n")

	cancel()

	if err := h.Close(); err != nil {
		panic(err)
	}
	os.Exit(0)
}

func art() {
	fmt.Println()
	fmt.Println("secure wallet for otwo blockchain")
	myFigure := figure.NewColorFigure("otwo WALLET", "", "yellow", true)
	myFigure.Print()
	fmt.Println()
}

func (al *addrList) String() string {
	strs := make([]string, len(*al))
	for i, addr := range *al {
		strs[i] = addr.String()
	}
	return strings.Join(strs, ",")
}

func (al *addrList) Set(value string) error {
	addr, err := multiaddr.NewMultiaddr(value)
	if err != nil {
		return err
	}
	*al = append(*al, addr)
	return nil
}
