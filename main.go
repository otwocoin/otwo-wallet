package main

import (
	"flag"

	"go.uber.org/zap"
)

var (
	ledger_path = "./ledger/wallet"
	logger, _   = zap.NewDevelopment()
)

func main() {
	port := flag.Uint("port", 8080, "port for oxygen wallet")
	flag.Parse()
	NewDir(ledger_path)
	InitWalletLedger(ledger_path)
	http := NewWeb(uint16(*port))
	http.Run()
}
