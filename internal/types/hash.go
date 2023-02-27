package types

import (
	"database/sql/driver"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
)

// Hash is a wrapper around common.Hash that implements the sql.Scanner and driver.Valuer interfaces.
type Hash struct {
	common.Hash
}

// Scan implements the Scanner interface. It is used by the sql package to convert a database value into Hash.
func (h *Hash) Scan(value interface{}) error {
	if value == nil {
		h.Hash = common.Hash{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			h.Hash = common.Hash{}
			return nil
		}
		if len(v) != common.HashLength {
			return fmt.Errorf("invalid hash length %v", len(v))
		}
		copy(h.Hash[:], v)
	case string:
		if len(v) == 0 {
			h.Hash = common.Hash{}
			return nil
		}
		if len(v) == common.HashLength {
			copy(h.Hash[:], v)
		} else if len(v) == common.HashLength*2+2 && v[:2] == "0x" {
			b, err := hex.DecodeString(v[2:])
			if err != nil {
				return err
			}
			copy(h.Hash[:], b)
		} else {
			return fmt.Errorf("invalid hash string %v", v)
		}
	default:
		return fmt.Errorf("unsupported hash type %T", value)
	}
	return nil
}

// Value implements the driver.Valuer interface. This method is used by the database/sql
// package to convert Hash into a value that can be stored in a database. The converted value is a string
// without the 0x prefix.
func (h *Hash) Value() (driver.Value, error) {
	if h == nil || h.Hash == (common.Hash{}) {
		return nil, nil
	}
	return h.Hash.Hex()[2:], nil
}
