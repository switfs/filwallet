package messagesigner

import (
	"fmt"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/wallet/key"
	_ "github.com/filecoin-project/lotus/lib/sigs/bls"
	_ "github.com/filecoin-project/lotus/lib/sigs/secp"
	logging "github.com/ipfs/go-log/v2"
	"github.com/switfs/filwallet/lib/sigs"
	_ "github.com/switfs/filwallet/lib/sigs/bls"
	_ "github.com/switfs/filwallet/lib/sigs/secp"
	"github.com/switfs/filwallet/modules/buildmessage"
	"golang.org/x/xerrors"
	"sync"
)

var log = logging.Logger("buildmessage")

type Signer interface {
	RegisterSigner(...key.Key) error
	SignMsg(msg *types.Message) (*types.SignedMessage, error)
	Sign(from string, data []byte) (*crypto.Signature, error)
	HasSigner(addr string) bool
}

type SignerHouse struct {
	signers map[string]key.Key // key is address
	lk      sync.Mutex
}

func NewSigner() Signer {
	return &SignerHouse{
		signers: map[string]key.Key{},
	}
}

func (s *SignerHouse) RegisterSigner(keys ...key.Key) error {
	s.lk.Lock()
	defer s.lk.Unlock()

	for _, key := range keys {
		if _, ok := s.signers[key.Address.String()]; ok {
			return fmt.Errorf("wallet: %s already exist", key.Address.String())
		}

		s.signers[key.Address.String()] = key
		log.Infow("RegisterSigner", "address", key.Address.String())
	}

	return nil
}

func (s *SignerHouse) SignMsg(msg *types.Message) (*types.SignedMessage, error) {
	s.lk.Lock()
	defer s.lk.Unlock()

	signer, ok := s.signers[msg.From.String()]
	if !ok {
		return nil, fmt.Errorf("wallet: %s does not exist", msg.From.String())
	}

	mb, err := msg.ToStorageBlock()
	if err != nil {
		return nil, xerrors.Errorf("serializing message: %w", err)
	}

	sig, err := sigs.Sign(key.ActSigType(signer.Type), signer.PrivateKey, mb.Cid().Bytes())
	if err != nil {
		return nil, xerrors.Errorf("failed to sign message: %w", err)
	}

	log.Infow("SignMsg", "message", buildmessage.LotusMessageToString(msg))
	return &types.SignedMessage{
		Message:   *msg,
		Signature: *sig,
	}, nil
}

func (s *SignerHouse) Sign(from string, data []byte) (*crypto.Signature, error) {
	s.lk.Lock()
	defer s.lk.Unlock()

	signer, ok := s.signers[from]
	if !ok {
		return nil, fmt.Errorf("wallet: %s does not exist", from)
	}

	sig, err := sigs.Sign(key.ActSigType(signer.Type), signer.PrivateKey, data)
	if err != nil {
		return nil, xerrors.Errorf("failed to sign message: %w", err)
	}

	log.Infow("Sign", "data", string(data))
	return sig, nil
}

func (s *SignerHouse) HasSigner(addr string) bool {
	s.lk.Lock()
	defer s.lk.Unlock()

	_, ok := s.signers[addr]

	return ok
}