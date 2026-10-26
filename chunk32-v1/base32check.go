
package main

import (
	"fmt"
	"math/big"
)

const (
	checkBase      = 37
	checkAlphabet  = "*~$=U"
	invalidCheck   = 255
	checkSuffixLen = 1
)

var checkModulus = big.NewInt(checkBase)

// getCheckByte преобразует значение контрольной суммы в символ.
func getCheckByte(mod uint8) byte {
	if mod < 32 {
		return alphabet[int(mod)]
	}

	if int(mod-32) >= len(checkAlphabet) {
		panic(fmt.Sprintf("invalid check value: %d", mod))
	}

	return checkAlphabet[int(mod-32)]
}

// getCheckValue преобразует контрольный символ в числовое значение.
func getCheckValue(b byte) uint8 {
	switch b {
	case '*':
		return 32
	case '~':
		return 33
	case '$':
		return 34
	case '=':
		return 35
	case 'u', 'U':
		return 36
	default:
		return b32[b]
	}
}

// CheckEncode кодирует данные в Base58 и добавляет контрольный символ.
func CheckEncode(b []byte) string {
	encoded, mod := encode(b, true)

	if mod >= checkBase {
		panic(fmt.Sprintf("invalid check value: %d", mod))
	}

	return encoded + string(getCheckByte(mod))
}

// CheckDecode декодирует данные и проверяет контрольный символ.
func CheckDecode(s string) ([]byte, error) {
	if len(s) < checkSuffixLen {
		return nil, fmt.Errorf("check decode: input is empty")
	}

	payload := s[:len(s)-checkSuffixLen]
	checkChar := s[len(s)-checkSuffixLen]

	output, value, err := decode(payload)
	if err != nil {
		return nil, fmt.Errorf("decode checked payload: %w", err)
	}

	expected := getCheckValue(checkChar)
	if expected == invalidCheck {
		return nil, fmt.Errorf("invalid check character %q", checkChar)
	}

	actual := new(big.Int).Mod(
		new(big.Int).Set(value),
		checkModulus,
	).Uint64()

	if actual != uint64(expected) {
		return nil, fmt.Errorf(
			"checksum mismatch: expected %d, got %d (character %q)",
			actual,
			expected,
			checkChar,
		)
	}

	return output, nil
}
