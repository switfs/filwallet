package wallet

import (
	logging "github.com/ipfs/go-log/v2"
	"github.com/switfs/filwallet/account"
	"github.com/switfs/filwallet/crypto"
	"github.com/switfs/filwallet/datastore"
	"github.com/switfs/filwallet/modules/messagesigner"
	"sync"
)

var log = logging.Logger("wallet-server")

type Wallet struct {
	*login
	*node
	*txTracker

	offline bool

	signer messagesigner.Signer

	masterPassword string

	db datastore.WalletDB
	lk sync.Mutex
}

func NewWallet(offline bool, masterPassword string, db datastore.WalletDB, close <-chan struct{}) (*Wallet, error) {
	login := newLogin(close)

	w := &Wallet{
		offline:        offline,
		login:          login,
		signer:         messagesigner.NewSigner(),
		masterPassword: masterPassword,
		db:             db,
	}

	nodeInfo, err := w.getBestNode()
	if err != nil {
		return nil, err
	}

	n, err := newNode(nodeInfo.Name, nodeInfo.Endpoint, nodeInfo.Token)
	if err == nil {
		w.node = n
	} else {
		log.Warn("no nodes available")
	}

	txTracker := newTxTracker(n, db, close)
	w.txTracker = txTracker

	keys, err := account.LoadPrivateKeys(db, crypto.GenerateEncryptKey([]byte(masterPassword)))
	if err != nil {
		log.Warnw("NewWallet: LoadPrivateKeys", "err", err)
		return nil, err
	}

	err = w.signer.RegisterSigner(keys...)
	if err != nil {
		return nil, err
	}

	return w, nil
}
