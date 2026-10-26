
package main

import (
	"fmt"
	"math/big"
)

const (
	base58Radix = 58
	invalidChar = 255
)

var (
	bigRadix = big.NewInt(base58Radix)
	bigZero  = big.NewInt(0)
	bigMod37 = big.NewInt(37)
)

// decode преобразует Base58-строку в байты и числовое значение.
func decode(s string) ([]byte, *big.Int, error) {
	value := new(big.Int)
	digit := new(big.Int)

	for i := 0; i < len(s); i++ {
		index := b58[s[i]]
		if index == invalidChar {
			return nil, nil, fmt.Errorf(
				"decode: invalid Base58 character %q at byte %d",
				s[i],
				i,
			)
		}

		value.Mul(value, bigRadix)
		digit.SetUint64(uint64(index))
		value.Add(value, digit)
	}

	decoded := value.Bytes()

	// В Base58 ведущий символ '1' кодирует нулевой байт.
	leadingZeros := 0
	for leadingZeros < len(s) && s[leadingZeros] == alphabetIdx0 {
		leadingZeros++
	}

	output := make([]byte, leadingZeros+len(decoded))
	copy(output[leadingZeros:], decoded)

	return output, new(big.Int).Set(value), nil
}

// Decode декодирует строку Base58 в исходные байты.
func Decode(s string) ([]byte, error) {
	output, _, err := decode(s)
	return output, err
}

// encode кодирует байты в Base58.
// При wantCheck возвращает остаток исходного числа по модулю 37.
func encode(b []byte, wantCheck bool) (string, uint8) {
	value := new(big.Int).SetBytes(b)
	var checksum uint8

	if wantCheck {
		checksumValue := new(big.Int).Mod(
			new(big.Int).Set(value),
			bigMod37,
		)
		checksum = uint8(checksumValue.Uint64())
	}

	output := make([]byte, 0, len(b)*138/100+1)
	remainder := new(big.Int)

	for value.Sign() > 0 {
		value.DivMod(value, bigRadix, remainder)
		output = append(output, alphabet[remainder.Int64()])
	}

	// Сохраняем каждый ведущий нулевой байт как символ '1'.
	for _, value := range b {
		if value != 0 {
			break
		}
		output = append(output, alphabetIdx0)
	}

	// Результат деления формируется справа налево.
	for left, right := 0, len(output)-1; left < right; left, right = left+1, right-1 {
		output[left], output[right] = output[right], output[left]
	}

	return string(output), checksum
}

// Encode кодирует байты в Base58.
func Encode(b []byte) string {
	output, _ := encode(b, false)
	return output
}
