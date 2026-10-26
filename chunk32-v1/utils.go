
package main

import "fmt"

// splitBytesIntoChunks разбивает байтовый срез на блоки размером size.
// Последний блок может быть короче size.
// Возвращаемые блоки ссылаются на исходный массив data.
func splitBytesIntoChunks(data []byte, size int) [][]byte {
	if size <= 0 {
		panic(fmt.Sprintf("splitBytesIntoChunks: size must be positive, got %d", size))
	}

	if len(data) == 0 {
		return nil
	}

	chunks := make([][]byte, 0, (len(data)+size-1)/size)

	for len(data) > 0 {
		n := min(len(data), size)
		chunks = append(chunks, data[:n])
		data = data[n:]
	}

	return chunks
}

// splitStringIntoChunks разбивает строку на блоки размером size байт.
// Последний блок может быть короче size.
// Разбиение выполняется по байтам, а не по Unicode-символам.
func splitStringIntoChunks(data string, size int) []string {
	if size <= 0 {
		panic(fmt.Sprintf("splitStringIntoChunks: size must be positive, got %d", size))
	}

	if len(data) == 0 {
		return nil
	}

	chunks := make([]string, 0, (len(data)+size-1)/size)

	for len(data) > 0 {
		n := min(len(data), size)
		chunks = append(chunks, data[:n])
		data = data[n:]
	}

	return chunks
}
