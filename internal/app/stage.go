package app

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"

	net "github.com/avvvet/otwo-wallet/internal/pkg/net"
	"github.com/avvvet/oxygen/pkg/blockchain"

	"github.com/avvvet/oxygen/pkg/util"

	"github.com/avvvet/oxygen/pkg/wallet"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

func BroadcastTransaction(ctx context.Context, token int,
	receiverAddress string, walletIndex int, t *pubsub.Topic, wl []*wallet.WalletAddressByte) error {

	w := wl[walletIndex]

	randomString, err := util.GenerateRandomString(32)
	if err != nil {
		// Serve an appropriately vague error to the
		// user, but log the details internally.
		logger.Sugar().Fatal(err)
		return errors.New("os level error")
	}

	/*create the raw transaction*/
	rawTx1 := &wallet.RawTx{
		SenderPublicKey:       w.PublicKey,
		SenderWalletAddress:   w.WalletAddress,
		SenderRandomHash:      sha256.Sum256([]byte(randomString)),
		Token:                 token,
		ReceiverPublicKey:     nil,
		ReceiverWalletAddress: receiverAddress,
	}

	/*sign the raw transaction output*/
	txoutput := &blockchain.TxOutput{
		RawTx:     rawTx1,
		Signature: rawTx1.Sign(util.DecodePrivateKey(w.PrivateKey)),
	}

	txByte, err := json.Marshal(txoutput)
	if err != nil {
		logger.Sugar().Fatal(err)
		return errors.New("transaction encoding error")
	}

	data, err := json.Marshal(&net.BroadcastData{Type: "NEWTXOUTPUT", Data: txByte})
	if err != nil {
		logger.Sugar().Fatal(err)
		return errors.New("broadcast data encoding error")
	}

	/* broadcast the signed transaction*/
	err = t.Publish(ctx, data)
	if err != nil {
		logger.Sugar().Warn(err)
		return errors.New("broadcast error")
	}
	logger.Sugar().Info("💥️ transaction broadcasted")
	return nil
}
