package main

import (
	"flag"
	"fmt"

	"github.com/common-nighthawk/go-figure"
	"go.uber.org/zap"
)

var (
	ledger_path = "./ledger/wallet"
	logger, _   = zap.NewDevelopment()
)

func main() {
	art()
	port := flag.Uint("port", 0, "port for otwo wallet")
	flag.Parse()
	NewDir(ledger_path)
	InitWalletLedger(ledger_path)
	http := NewWeb(uint16(*port))
	http.Run()
}

func art() {
	fmt.Println()
	fmt.Println("secure wallet for otwo blockchain")
	myFigure := figure.NewColorFigure("otwo WALLET", "", "yellow", true)
	myFigure.Print()
	fmt.Println()
}
