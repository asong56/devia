package core

import (
	"math/big"
	"strconv"
	"strings"
)

type RadixResult struct {
	Bin string `json:"bin"`
	Oct string `json:"oct"`
	Dec string `json:"dec"`
	Hex string `json:"hex"`
}

func detectBase(s string) (cleaned string, base int) {
	t := strings.TrimSpace(s)
	lower := strings.ToLower(t)
	switch {
	case strings.HasPrefix(lower, "0x"):
		return t[2:], 16
	case strings.HasPrefix(lower, "0o"):
		return t[2:], 8
	case strings.HasPrefix(lower, "0b"):
		return t[2:], 2
	default:
		return t, 10
	}
}

// ConvertRadix parses input in fromBase (0 = auto-detect via a
// 0x/0o/0b prefix, decimal otherwise) and returns its value in binary,
// octal, decimal, and hex. Uses math/big so arbitrarily large integers
// don't silently overflow.
func ConvertRadix(input string, fromBase int) (*RadixResult, error) {
	cleaned := strings.TrimSpace(input)
	base := fromBase
	if base == 0 {
		cleaned, base = detectBase(input)
	}
	n := new(big.Int)
	if _, ok := n.SetString(cleaned, base); !ok {
		return nil, NewInputError("invalid number for base " + strconv.Itoa(base) + ": " + input)
	}
	return &RadixResult{
		Bin: n.Text(2),
		Oct: n.Text(8),
		Dec: n.Text(10),
		Hex: n.Text(16),
	}, nil
}
