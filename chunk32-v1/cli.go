
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// encodeFromStdin читает данные из stdin, кодирует их и выводит результат.
func encodeFromStdin(upper, check, noNewline bool) error {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	if len(input) == 0 {
		return nil
	}

	// Удаляем только завершающие LF, сохраняя остальные байты.
	input = bytes.TrimRight(input, "\n")

	text := Chunk32Encode(input, check)
	if !upper {
		text = strings.ToLower(text)
	}

	if noNewline {
		_, err = fmt.Print(text)
	} else {
		_, err = fmt.Println(text)
	}
	if err != nil {
		return fmt.Errorf("write encoded output: %w", err)
	}

	return nil
}

// decodeFromStdin читает закодированный текст из stdin и выводит исходные данные.
func decodeFromStdin(noNewline bool) error {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	if len(input) == 0 {
		return nil
	}

	input = bytes.TrimRight(input, "\n")

	output, err := Chunk32Decode(string(input))
	if err != nil {
		return fmt.Errorf("decode input: %w", err)
	}

	if noNewline {
		_, err = os.Stdout.Write(output)
	} else {
		_, err = os.Stdout.Write(append(output, '\n'))
	}
	if err != nil {
		return fmt.Errorf("write decoded output: %w", err)
	}

	return nil
}
