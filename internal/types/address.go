package types

import (
	"database/sql/driver"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
)

// Address is a wrapper around common.Address that implements the sql.Scanner and driver.Valuer interfaces.
type Address struct {
	common.Address
}

// Scan implements the Scanner interface. It is used by the sql package to convert a database value into Address.
func (a *Address) Scan(value interface{}) error {
	if value == nil {
		a.Address = common.Address{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			a.Address = common.Address{}
			return nil
		}
		if len(v) != common.AddressLength {
			return fmt.Errorf("invalid address length %v", len(v))
		}
		copy(a.Address[:], v)
	case string:
		if len(v) == 0 {
			a.Address = common.Address{}
			return nil
		}
		if len(v) == common.AddressLength {
			copy(a.Address[:], v)
		} else if len(v) == common.AddressLength*2+2 && v[:2] == "0x" {
			b, err := hex.DecodeString(v[2:])
			if err != nil {
				return err
			}
			copy(a.Address[:], b)
		} else {
			return fmt.Errorf("invalid address string %v", v)
		}
	default:
		return fmt.Errorf("unsupported address type %T", value)
	}
	return nil
}

// Value implements the driver.Valuer interface. This method is used by the database/sql
// package to convert Address into a value that can be stored in a database. The converted value is a string
// without the 0x prefix.
func (a *Address) Value() (driver.Value, error) {
	if a == nil || a.Address == (common.Address{}) {
		return nil, nil
	}
	return a.Address.Hex()[2:], nil
}
