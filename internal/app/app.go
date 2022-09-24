package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"text/template"

	//"github.com/avvvet/oxygen-wallet/internal/pkg/wallet"
	"github.com/avvvet/oxygen/pkg/kv"

	"github.com/avvvet/oxygen/pkg/wallet"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/savioxavier/termlink"
	"go.uber.org/zap"
)

const tempDir = "templates"

var (
	logger, _ = zap.NewDevelopment()
)

type HttpServer struct {
	port        uint
	topic       *pubsub.Topic
	walletStore []*wallet.WalletAddressByte
	context     context.Context
}

func (h *HttpServer) GetPort() uint {
	return h.port
}

func NewApp(ctx context.Context, port uint, t *pubsub.Topic, wl []*wallet.WalletAddressByte) *HttpServer {
	return &HttpServer{port, t, wl, ctx}
}

func NewDir(path string) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
		logger.Sugar().Info("local store for otwo wallet created \n")
	} else {
		logger.Sugar().Info("local wallet store found \n")
	}
}

func InitWalletLedger(path string) ([]*wallet.WalletAddressByte, error) {
	var listWallet []*wallet.WalletAddressByte

	ledger, err := kv.NewLedger(path)
	if err != nil {
		logger.Sugar().Fatal("unable to initialize wallet ledger.")
	}

	iter := ledger.Db.NewIterator(nil, nil)
	if !iter.Last() {
		wallet := wallet.NewWallet()
		/* encode wallet before storing and decod before usage*/
		walletByte := wallet.EncodeWallet()
		b, _ := json.Marshal(walletByte)

		err = ledger.Upsert([]byte(wallet.WalletAddress), b)
		if err != nil {
			return nil, err
		}
		iter.Release()
		listWallet = append(listWallet, walletByte)
		logger.Sugar().Info("new wallet address created and stored in local ledger \n")
		return listWallet, err
	}

	for ok := iter.First(); ok; ok = iter.Next() {

		v := iter.Value()

		w := &wallet.WalletAddressByte{}
		err = json.Unmarshal(v, w)
		listWallet = append(listWallet, w)
		if err != nil {
			return nil, err
		}
	}
	iter.Release()
	logger.Sugar().Info("existing wallets fetched \n")
	return listWallet, nil
}

func (h *HttpServer) Status(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Add("Content-Type", "application/json")
		//lets create struct on th fly
		data, _ := json.Marshal(struct {
			Status string
		}{
			Status: "server started...",
		})
		io.WriteString(w, string(data[:]))
	}
}

func (h *HttpServer) Index(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		t, _ := template.ParseFiles(path.Join(tempDir, "index.html"))
		t.Execute(w, "")
	default:
		log.Printf("ERROR: Invalid HTTP Method")
	}
}

func (h *HttpServer) Wallet(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Add("Content-Type", "application/json")
		myWallet := h.walletStore
		m, _ := json.Marshal(myWallet)
		io.WriteString(w, string(m[:]))
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("ERROR: Invalid HTTP Method")
	}
}

type TransactionReq struct {
	Token                 int
	WalletIndex           int
	ReceiverWalletAddress string
}

func (h *HttpServer) Transaction(w http.ResponseWriter, req *http.Request) {
	resp := make(map[string]string)
	switch req.Method {
	case http.MethodGet:
		t, _ := template.ParseFiles(path.Join(tempDir, "index_cpu.html"))
		t.Execute(w, "")
	case http.MethodPost:
		tReq := &TransactionReq{}
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(tReq)
		if err != nil {
			io.WriteString(w, "error to decode")
			return
		}

		err = BroadcastTransaction(h.context, tReq.Token, tReq.ReceiverWalletAddress, tReq.WalletIndex, h.topic, h.walletStore)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp["message"] = err.Error()
			jsonResp, _ := json.Marshal(resp)
			io.WriteString(w, string(jsonResp))
		}

		w.WriteHeader(http.StatusAccepted)
		resp["message"] = "Transaction Broadcasted"
		jsonResp, _ := json.Marshal(resp)
		io.WriteString(w, string(jsonResp))
	default:
		log.Printf("ERROR: Invalid HTTP Method")
	}
}

func (h *HttpServer) Run() {
	http.HandleFunc("/", h.Wallet)
	http.HandleFunc("/transaction", h.Transaction)
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(int(h.port)))
	if err != nil {
		panic(err)
	}
	logger.Sugar().Infof("listening web requests at port 😎️ %v", listener.Addr().(*net.TCPAddr).Port)
	fmt.Println("")
	fmt.Println(termlink.ColorLink("access your otwo wallet app at this link 👉️ ", "http://127.0.0.1:"+strconv.Itoa(listener.Addr().(*net.TCPAddr).Port), "italic yellow"))
	if err := http.Serve(listener, nil); err != nil {
		logger.Sugar().Fatal(err)
	}
}
