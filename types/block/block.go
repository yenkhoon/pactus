package block

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/pactus-project/pactus/crypto"
	"github.com/pactus-project/pactus/crypto/hash"
	"github.com/pactus-project/pactus/sortition"
	"github.com/pactus-project/pactus/types/tx"
	"github.com/pactus-project/pactus/util"
	"github.com/pactus-project/pactus/util/encoding"
	"github.com/pactus-project/pactus/util/errors"
)

type Block struct {
	memorizedHash *hash.Hash
	memorizedData []byte
	data          blockData
}

type blockData struct {
	Header   Header
	PrevCert *Certificate
	Txs      Txs
}

func NewBlock(header Header, prevCert *Certificate, txs Txs) *Block {
	return &Block{
		data: blockData{
			Header:   header,
			PrevCert: prevCert,
			Txs:      txs,
		},
	}
}

// FromBytes constructs a new block from byte array.
func FromBytes(data []byte) (*Block, error) {
	b := new(Block)
	r := bytes.NewReader(data)
	if err := b.Decode(r); err != nil {
		return nil, err
	}

	return b, nil
}

func MakeBlock(version uint8, timestamp time.Time, txs Txs,
	prevBlockHash, stateRoot hash.Hash,
	prevCert *Certificate, sortitionSeed sortition.VerifiableSeed, proposer crypto.Address) *Block {
	header := NewHeader(version, timestamp,
		stateRoot, prevBlockHash, sortitionSeed, proposer)

	b := NewBlock(header, prevCert, txs)
	if err := b.SanityCheck(); err != nil {
		panic(err)
	}
	return b
}

func (b *Block) Header() *Header               { return &b.data.Header }
func (b *Block) PrevCertificate() *Certificate { return b.data.PrevCert }
func (b *Block) Transactions() Txs             { return b.data.Txs }

func (b *Block) SanityCheck() error {
	if err := b.Header().SanityCheck(); err != nil {
		return err
	}
	if b.Transactions().Len() == 0 {
		return errors.Errorf(errors.ErrInvalidBlock, "block at least should have one transaction")
	}
	if b.Transactions().Len() > 1000 {
		return errors.Errorf(errors.ErrInvalidBlock, "block is full")
	}
	if b.PrevCertificate() != nil {
		if err := b.PrevCertificate().SanityCheck(); err != nil {
			return err
		}
		if err := b.Header().PrevBlockHash().SanityCheck(); err != nil {
			return err
		}
	} else {
		// Genesis block checks
		if !b.Header().PrevBlockHash().IsUndef() {
			return errors.Errorf(errors.ErrInvalidBlock, "invalid previous block hash")
		}
	}

	for _, trx := range b.Transactions() {
		if err := trx.SanityCheck(); err != nil {
			return errors.Errorf(errors.ErrInvalidBlock, err.Error())
		}
	}

	return nil
}

func (b *Block) Hash() hash.Hash {
	if b.memorizedHash != nil {
		return *b.memorizedHash
	}

	w := &bytes.Buffer{}
	if err := b.data.Header.Encode(w); err != nil {
		return hash.UndefHash
	}
	// Genesis block has no certificate
	if b.data.PrevCert != nil {
		w.Write(b.data.PrevCert.Hash().Bytes())
	}
	w.Write(b.data.Txs.Root().Bytes())
	w.Write(util.Int32ToSlice(int32(b.data.Txs.Len())))

	h := hash.CalcHash(w.Bytes())
	b.memorizedHash = &h
	return h
}

func (b *Block) Stamp() hash.Stamp {
	return b.Hash().Stamp()
}

func (b *Block) Fingerprint() string {
	return fmt.Sprintf("{⌘ %v 👤 %v 💻 %v 📨 %d}",
		b.Hash().Fingerprint(),
		b.data.Header.ProposerAddress().Fingerprint(),
		b.data.Header.StateRoot().Fingerprint(),
		b.data.Txs.Len(),
	)
}

func (b *Block) MarshalCBOR() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, b.SerializeSize()))
	if err := b.Encode(buf); err != nil {
		return nil, err
	}
	return cbor.Marshal(buf.Bytes())
}

func (b *Block) UnmarshalCBOR(bs []byte) error {
	data := make([]byte, 0, b.SerializeSize())
	err := cbor.Unmarshal(bs, &data)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(data)
	return b.Decode(buf)
}

// Encode encodes the receiver to w.
func (b *Block) Encode(w io.Writer) error {
	if err := b.data.Header.Encode(w); err != nil {
		return err
	}
	if b.data.PrevCert != nil {
		if err := b.data.PrevCert.Encode(w); err != nil {
			return err
		}
	}
	if err := encoding.WriteVarInt(w, uint64(b.data.Txs.Len())); err != nil {
		return err
	}
	for _, tx := range b.Transactions() {
		if err := tx.Encode(w); err != nil {
			return err
		}
	}
	return nil
}

func (b *Block) Decode(r io.Reader) error {
	if err := b.data.Header.Decode(r); err != nil {
		return err
	}
	if !b.data.Header.PrevBlockHash().IsUndef() {
		b.data.PrevCert = new(Certificate)
		if err := b.data.PrevCert.Decode(r); err != nil {
			return err
		}
	}
	len, err := encoding.ReadVarInt(r)
	if err != nil {
		return err
	}
	b.data.Txs = make([]*tx.Tx, len)
	for i := 0; i < int(len); i++ {
		tx := new(tx.Tx)
		if err := tx.Decode(r); err != nil {
			return err
		}
		b.data.Txs[i] = tx
	}
	return nil
}

// SerializeSize returns the number of bytes it would take to serialize the block.
func (b *Block) SerializeSize() int {
	n := b.Header().SerializeSize()

	if b.PrevCertificate() != nil {
		n += b.PrevCertificate().SerializeSize()
	}

	n += encoding.VarIntSerializeSize(uint64(b.Transactions().Len()))
	for _, tx := range b.Transactions() {
		n += tx.SerializeSize()
	}
	return n
}

// Bytes returns the serialized bytes for the Block. It caches the
// result so subsequent calls are more efficient.
func (b *Block) Bytes() ([]byte, error) {
	// Return the cached serialized bytes if it has already been generated.
	if len(b.memorizedData) != 0 {
		return b.memorizedData, nil
	}

	w := bytes.NewBuffer(make([]byte, 0, b.SerializeSize()))
	err := b.Encode(w)
	if err != nil {
		return nil, err
	}

	// Cache the serialized bytes and return them.
	b.memorizedData = w.Bytes()
	return b.memorizedData, nil
}

// GenerateTestBlock generates a block vote for testing.
func GenerateTestBlock(proposer *crypto.Address, prevBlockHash *hash.Hash) *Block {
	if proposer == nil {
		addr := crypto.GenerateTestAddress()
		proposer = &addr
	}
	txs := NewTxs()
	tx1, _ := tx.GenerateTestSendTx()
	tx2, _ := tx.GenerateTestSendTx()
	tx3, _ := tx.GenerateTestSendTx()
	tx4, _ := tx.GenerateTestSendTx()

	txs.Append(tx1)
	txs.Append(tx2)
	txs.Append(tx3)
	txs.Append(tx4)

	if prevBlockHash == nil {
		h := hash.GenerateTestHash()
		prevBlockHash = &h
	}
	cert := GenerateTestCertificate(*prevBlockHash)
	if prevBlockHash.IsUndef() {
		cert = nil
	}
	sortitionSeed := sortition.GenerateRandomSeed()
	header := NewHeader(1, util.Now(),
		hash.GenerateTestHash(),
		*prevBlockHash,
		sortitionSeed,
		*proposer)

	return NewBlock(header, cert, txs)
}
