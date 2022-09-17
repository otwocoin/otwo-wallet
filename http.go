package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"text/template"

	"github.com/avvvet/oxygen-wallet/wallet"
	"github.com/avvvet/oxygen/pkg/kv"
)

const tempDir = "templates"

type HttpServer struct {
	port uint16
}

func (h *HttpServer) GetPort() uint16 {
	return h.port
}

func NewWeb(port uint16) *HttpServer {
	return &HttpServer{port}
}

func NewDir(path string) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
}

func InitWalletLedger(path string) ([]*wallet.Wo, error) {
	var listWallet []*wallet.Wo

	ledger, err := kv.NewLedger(path)
	if err != nil {
		logger.Sugar().Fatal("unable to initialize ledger.")
	}

	iter := ledger.Db.NewIterator(nil, nil)
	if !iter.Last() {
		wallet := wallet.NewWallet()
		b, _ := json.Marshal(wallet)

		err = ledger.Upsert([]byte(wallet.BlockchainAddress), b)
		if err != nil {
			return nil, err
		}
		iter.Release()
		listWallet = append(listWallet, wallet)
		return listWallet, err
	}

	for ok := iter.First(); ok; ok = iter.Next() {

		v := iter.Value()

		w := &wallet.Wo{}
		err = json.Unmarshal(v, w)
		listWallet = append(listWallet, w)
		if err != nil {
			return nil, err
		}
	}
	iter.Release()
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
		myWallet := wallet.NewWallet()
		m, _ := json.Marshal(myWallet)
		io.WriteString(w, string(m[:]))
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("ERROR: Invalid HTTP Method")
	}
}

func (h *HttpServer) Run() {
	http.HandleFunc("/", h.Wallet)
	fmt.Printf("starting sever at port %v", strconv.Itoa(int(h.port)))
	http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(h.port)), nil)
}
