
package main

import (
	"bytes"
	"fmt"
	"strings"
)

const chunkSize = 10
const displayGroupSize = 4

// addTrailingDashes дополняет строку дефисами до кратной длины.
func addTrailingDashes(line string, size int) string {
	if size <= 0 {
		return line
	}

	remainder := len(line) % size
	if remainder == 0 {
		return line
	}

	return line + strings.Repeat("-", size-remainder)
}

// addSpacesToPlainTextLine разбивает строку на группы фиксированной длины.
func addSpacesToPlainTextLine(line string, size int) string {
	if size <= 0 {
		return line
	}

	return strings.Join(splitStringIntoChunks(line, size), " ")
}

// encodeChunk кодирует блок без контрольного символа.
func encodeChunk(chunk []byte) string {
	line := Encode(chunk)
	line = addTrailingDashes(line, displayGroupSize)

	return addSpacesToPlainTextLine(line, displayGroupSize)
}

// encodeChunkWithCheck кодирует блок с контрольным символом.
// Контрольный символ отделяется пробелом и не дополняется дефисами.
func encodeChunkWithCheck(chunk []byte) string {
	line := CheckEncode(chunk)
	if len(line) == 0 {
		return ""
	}

	payload := line[:len(line)-1]
	checkChar := line[len(line)-1]

	payload = addTrailingDashes(payload, displayGroupSize)
	payload = addSpacesToPlainTextLine(payload, displayGroupSize)

	return payload + " " + string(checkChar)
}

// Chunk32Encode кодирует данные блоками и объединяет их переводами строк.
func Chunk32Encode(data []byte, check bool) string {
	chunks := splitBytesIntoChunks(data, chunkSize)
	lines := make([]string, len(chunks))

	for i, chunk := range chunks {
		if check {
			lines[i] = encodeChunkWithCheck(chunk)
		} else {
			lines[i] = encodeChunk(chunk)
		}
	}

	return strings.Join(lines, "\n")
}

// decodeChunk декодирует одну строку.
// Наличие контрольного символа определяется по отдельному последнему полю.
func decodeChunk(line string, allowCheck bool) (bool, []byte, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return false, nil, fmt.Errorf("empty encoded chunk")
	}

	hadCheck := false
	payload := line

	fields := strings.Fields(line)
	if len(fields) > 1 && len(fields[len(fields)-1]) == 1 {
		if !allowCheck {
			return false, nil, fmt.Errorf(
				"unexpected check character in unchecked format",
			)
		}

		hadCheck = true
		payload = strings.Join(fields[:len(fields)-1], "")
		payload += " " + fields[len(fields)-1]
	} else {
		payload = strings.Join(fields, "")
	}

	payload = strings.ReplaceAll(payload, "-", "")

	var (
		chunk []byte
		err   error
	)

	if hadCheck {
		chunk, err = CheckDecode(payload)
	} else {
		chunk, err = Decode(payload)
	}

	if err != nil {
		return hadCheck, nil, err
	}

	return hadCheck, chunk, nil
}

// Chunk32Decode декодирует многострочный текст в исходные байты.
// Все непустые строки должны использовать один формат контрольной суммы.
func Chunk32Decode(text string) ([]byte, error) {
	if strings.TrimSpace(text) == "" {
		return []byte{}, nil
	}

	lines := strings.Split(text, "\n")
	chunks := make([][]byte, 0, len(lines))

	var (
		formatKnown bool
		checkMode   bool
	)

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		hadCheck, chunk, err := decodeChunk(line, true)
		if err != nil {
			return nil, fmt.Errorf("decode line %d: %w", i+1, err)
		}

		if !formatKnown {
			checkMode = hadCheck
			formatKnown = true
		} else if hadCheck != checkMode {
			return nil, fmt.Errorf(
				"decode line %d: inconsistent checksum format",
				i+1,
			)
		}

		chunks = append(chunks, chunk)
	}

	return bytes.Join(chunks, nil), nil
}
