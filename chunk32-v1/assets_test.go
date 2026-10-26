
package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	defaultDirMode  os.FileMode = 0o755
	defaultFileMode os.FileMode = 0o664
)

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi *bindataFileInfo) Name() string       { return fi.name }
func (fi *bindataFileInfo) Size() int64        { return fi.size }
func (fi *bindataFileInfo) Mode() os.FileMode  { return fi.mode }
func (fi *bindataFileInfo) ModTime() time.Time { return fi.modTime }
func (fi *bindataFileInfo) IsDir() bool        { return false }
func (fi *bindataFileInfo) Sys() any           { return nil }

// bindataRead распаковывает GZIP-ресурс и сохраняет исходную причину ошибки.
func bindataRead(data []byte, name string) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("open gzip asset %q: %w", name, err)
	}

	var buf bytes.Buffer

	_, readErr := io.Copy(&buf, reader)
	closeErr := reader.Close()

	if readErr != nil {
		return nil, fmt.Errorf("decompress asset %q: %w", name, readErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close gzip asset %q: %w", name, closeErr)
	}

	return buf.Bytes(), nil
}

// canonicalAssetName приводит разделители к единому виду.
func canonicalAssetName(name string) string {
	return strings.ReplaceAll(name, "\\", "/")
}

// lookupAsset находит ресурс по каноническому имени.
func lookupAsset(name string) (*asset, error) {
	name = canonicalAssetName(name)

	f, ok := _bindata[name]
	if !ok {
		return nil, fmt.Errorf("asset %q not found", name)
	}

	a, err := f()
	if err != nil {
		return nil, fmt.Errorf("load asset %q: %w", name, err)
	}
	return a, nil
}

// Asset возвращает содержимое встроенного файла.
func Asset(name string) ([]byte, error) {
	a, err := lookupAsset(name)
	if err != nil {
		return nil, err
	}
	return a.bytes, nil
}

// MustAsset возвращает содержимое или вызывает panic при ошибке.
func MustAsset(name string) []byte {
	data, err := Asset(name)
	if err != nil {
		panic(fmt.Sprintf("asset %q: %v", name, err))
	}
	return data
}

// AssetInfo возвращает метаданные встроенного файла.
func AssetInfo(name string) (os.FileInfo, error) {
	a, err := lookupAsset(name)
	if err != nil {
		return nil, fmt.Errorf("get asset metadata: %w", err)
	}
	return a.info, nil
}

// AssetNames возвращает стабильный отсортированный список ресурсов.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// AssetDir возвращает непосредственных потомков каталога.
func AssetDir(name string) ([]string, error) {
	name = canonicalAssetName(name)
	node := _bintree

	if name != "" {
		for _, part := range strings.Split(name, "/") {
			if part == "" || part == "." || part == ".." {
				return nil, fmt.Errorf("invalid asset directory %q", name)
			}

			if node == nil {
				return nil, fmt.Errorf("asset directory %q not found", name)
			}

			node = node.Children[part]
			if node == nil {
				return nil, fmt.Errorf("asset directory %q not found", name)
			}
		}
	}

	if node.Func != nil {
		return nil, fmt.Errorf("asset %q is a file, not a directory", name)
	}

	children := make([]string, 0, len(node.Children))
	for child := range node.Children {
		children = append(children, child)
	}
	sort.Strings(children)

	return children, nil
}

// safeAssetPath запрещает абсолютные пути и выход за корень восстановления.
func safeAssetPath(root, name string) (string, error) {
	name = canonicalAssetName(name)

	if name == "" || strings.HasPrefix(name, "/") {
		return "", fmt.Errorf("invalid asset path %q", name)
	}

	cleanName := filepath.FromSlash(name)
	if !filepath.IsLocal(cleanName) {
		return "", fmt.Errorf("asset path escapes destination: %q", name)
	}

	rootPath, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve destination root: %w", err)
	}

	target := filepath.Join(rootPath, cleanName)
	relative, err := filepath.Rel(rootPath, target)
	if err != nil {
		return "", fmt.Errorf("validate asset destination: %w", err)
	}
	if !filepath.IsLocal(relative) {
		return "", fmt.Errorf("asset path escapes destination: %q", name)
	}

	return target, nil
}

// RestoreAsset восстанавливает встроенный файл на диске.
func RestoreAsset(dir, name string) error {
	target, err := safeAssetPath(dir, name)
	if err != nil {
		return err
	}

	a, err := lookupAsset(name)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(target), defaultDirMode); err != nil {
		return fmt.Errorf("create parent directory for %q: %w", name, err)
	}

	if err := os.WriteFile(target, a.bytes, a.info.Mode().Perm()); err != nil {
		return fmt.Errorf("write restored asset %q: %w", name, err)
	}

	modTime := a.info.ModTime()
	if err := os.Chtimes(target, modTime, modTime); err != nil {
		return fmt.Errorf("restore timestamps for %q: %w", name, err)
	}

	return nil
}

// RestoreAssets рекурсивно восстанавливает каталог или отдельный файл.
func RestoreAssets(dir, name string) error {
	name = canonicalAssetName(name)

	children, err := AssetDir(name)
	if err != nil {
		// Если путь не является каталогом, пробуем восстановить файл.
		return RestoreAsset(dir, name)
	}

	for _, child := range children {
		childName := child
		if name != "" {
			childName = name + "/" + child
		}

		if err := RestoreAssets(dir, childName); err != nil {
			return fmt.Errorf("restore asset %q: %w", childName, err)
		}
	}

	return nil
}

// _filePath сохраняет совместимость с исходным внутренним API.
func _filePath(dir, name string) string {
	return filepath.Join(dir, filepath.FromSlash(canonicalAssetName(name)))
}
