package types

import (
	"database/sql/driver"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
)

// Big is a wrapper around hexutil.Big that implements the sql.Scanner and driver.Valuer interfaces.
type Big struct {
	hexutil.Big
}

// Scan implements the Scanner interface. It is used by the sql package to convert a database value into Big.
func (b *Big) Scan(value interface{}) error {
	if value == nil {
		b.Big = hexutil.Big{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			b.Big = hexutil.Big{}
			return nil
		}
		d, err := hex.DecodeString(string(v))
		if err != nil {
			return err
		}
		b.Big = hexutil.Big(*new(big.Int).SetBytes(d))
	case string:
		if len(v) == 0 {
			b.Big = hexutil.Big{}
			return nil
		}
		if len(v) > 2 && v[:2] == "0x" {
			v = v[2:]
		}
		d, err := hex.DecodeString(v)
		if err != nil {
			return err
		}
		b.Big = hexutil.Big(*new(big.Int).SetBytes(d))
	default:
		return fmt.Errorf("unsupported bigint type %T", value)
	}
	return nil
}

// Value implements the driver.Valuer interface. This method is used by the database/sql
// package to convert Big into a value that can be stored in a database.
func (b *Big) Value() (driver.Value, error) {
	return b.Big.String()[2:], nil
}
