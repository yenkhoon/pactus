package message

import (
	"fmt"

	"github.com/pactus-project/pactus/util/errors"
)

type BlocksRequestMessage struct {
	SessionID int    `cbor:"1,keyasint"`
	From      uint32 `cbor:"2,keyasint"`
	To        uint32 `cbor:"3,keyasint"`
}

func NewBlocksRequestMessage(sid int, from, to uint32) *BlocksRequestMessage {
	return &BlocksRequestMessage{
		SessionID: sid,
		From:      from,
		To:        to,
	}
}

func (m *BlocksRequestMessage) SanityCheck() error {
	if m.From > m.To {
		return errors.Errorf(errors.ErrInvalidHeight, "invalid range")
	}
	return nil
}

func (m *BlocksRequestMessage) Type() Type {
	return MessageTypeBlocksRequest
}

func (m *BlocksRequestMessage) Fingerprint() string {
	return fmt.Sprintf("{⚓ %d %v:%v}", m.SessionID, m.From, m.To)
}
