package joinserver

import (
	"crypto/aes"

	keywrap "github.com/NickBall/go-aes-key-wrap"
	"github.com/pkg/errors"

	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
)

func getKeyEnvelope(kekLabel string, kek []byte, key lorawan.AES128Key) (*backend.KeyEnvelope, error) {
	if kekLabel == "" || len(kek) == 0 {
		return &backend.KeyEnvelope{
			AESKey: backend.HEXBytes(key[:]),
		}, nil
	}

	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, errors.Wrap(err, "new cipher error")
	}

	b, err := keywrap.Wrap(block, key[:])
	if err != nil {
		return nil, errors.Wrap(err, "key wrap error")
	}

	return &backend.KeyEnvelope{
		KEKLabel: kekLabel,
		AESKey:   backend.HEXBytes(b),
	}, nil
}
