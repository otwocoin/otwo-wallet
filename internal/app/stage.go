package app

import (
	"context"
	"crypto/sha256"
	"encoding/json"

	//"github.com/avvvet/otwo-wallet/internal/pkg/util"
	net "github.com/avvvet/otwo-wallet/internal/pkg/net"
	"github.com/avvvet/oxygen/pkg/blockchain"

	//net "github.com/avvvet/oxygen/pkg/net"
	"github.com/avvvet/oxygen/pkg/util"

	"github.com/avvvet/oxygen/pkg/wallet"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

func SendTransaction(ctx context.Context, token int, walletIndex int, t *pubsub.Topic, wl []*wallet.WalletAddressByte) {
	w := wl[walletIndex]

	randomString, err := util.GenerateRandomString(32)
	if err != nil {
		// Serve an appropriately vague error to the
		// user, but log the details internally.
		panic(err)
	}

	/*create the raw transaction*/
	rawTx1 := &wallet.RawTx{
		SenderPublicKey:       w.PublicKey,
		SenderWalletAddress:   w.WalletAddress,
		SenderRandomHash:      sha256.Sum256([]byte(randomString)),
		Token:                 400,
		ReceiverPublicKey:     nil,
		ReceiverWalletAddress: w.WalletAddress,
	}

	/*sign the raw transaction output*/
	txoutput := &blockchain.TxOutput{
		RawTx:     rawTx1,
		Signature: rawTx1.Sign(util.DecodePrivateKey(w.PrivateKey)),
	}

	txByte, err := json.Marshal(txoutput)
	if err != nil {
		logger.Sugar().Fatal(err)
	}

	data, err := json.Marshal(&net.BroadcastData{Type: "NEWBLOCK", Data: txByte})
	if err != nil {
		logger.Sugar().Fatal(err)
	}

	/* broadcast the signed transaction*/
	err = t.Publish(ctx, data)
	if err != nil {
		logger.Sugar().Warn(err)
		return
	}
	logger.Sugar().Info("💥️ transaction broadcasted")
}
