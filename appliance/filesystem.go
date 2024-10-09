/**
 * This software and associated documentation files (the “Software”),
 * including GFI AppManager, is the property of GFI USA, LLC and its affiliates.
 * No part of the Software may be copied, modified, distributed, sold, or otherwise
 * used except as expressly permitted by the terms of the software license agreement.
 */

package appliance

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/trilogy-group/gfi-agent-sdk/logger"

	cp "github.com/otiai10/copy"
)

var FS *FileSystem

type FileSystem struct{}

func init() { FS = &FileSystem{} }

func (F *FileSystem) Copy(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func (F *FileSystem) CreateDir(path string) bool {
	return os.MkdirAll(path, 0755) == nil
}

func (F *FileSystem) RemoveDir(path string) bool {
	return os.RemoveAll(path) == nil
}

func (F *FileSystem) RemoveFile(path string) error {
	return os.Remove(path)
}

func (F *FileSystem) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func (F *FileSystem) Find(root, ext string) []string {
	var a []string
	if err := filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ext {
			a = append(a, s)
		}
		return nil
	}); err != nil {
		return []string{}
	}
	return a
}

func (F *FileSystem) CreateTempFile(dir string, pattern string) (*os.File, error) {
	return os.CreateTemp(dir, pattern)
}

func (F *FileSystem) CreateFile(path string) (*os.File, error) {
	return os.Create(path)
}

func (F *FileSystem) CreateTempDir(prefix string) (string, error) {
	return os.MkdirTemp("", prefix)
}

func (F *FileSystem) List(dir string) ([]fs.FileInfo, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	return f.Readdir(0)
}

func (F *FileSystem) ListDir(dir string) ([]string, error) {
	files, err := F.List(dir)
	if err != nil {
		return nil, err
	}
	dirs := []string{}
	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, file.Name())
		}
	}
	return dirs, nil
}

func (F *FileSystem) ListFiles(dir string, ext string) ([]string, error) {
	files, err := F.List(dir)
	if err != nil {
		return nil, err
	}
	dirs := []string{}
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ext {
			dirs = append(dirs, file.Name())
		}
	}
	return dirs, nil
}

func (F *FileSystem) CopyDir(src, dest string) error {
	return cp.Copy(src, dest)
}

func (F *FileSystem) UpdateFilePermissions(parent string, files []string, perm os.FileMode) error {
	for _, file := range files {
		if err := os.Chmod(filepath.Join(parent, file), perm); err != nil {
			return err
		}
	}
	return nil
}

func (F *FileSystem) GetFileSize(filePath string) (int64, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	size := fi.Size()
	return size, nil
}

func (F *FileSystem) GetFileHash(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	hash := h.Sum(nil)
	stringHash := hex.EncodeToString(hash[:])
	return stringHash, nil
}

func (F *FileSystem) WriteFile(path string, data []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}

func (F *FileSystem) IsFileLocked(path string) bool {
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		logger.Logger.Errorf("%s\n", err)
		return true
	}
	file.Close()
	return false
}

func RemountFS(mode string, appType string) error {
	if appType == "exinda" {
		_, err := exec.Command("remountrw").Output()
		return err
	}
	_, err := exec.Command("/bin/mount", "-o", "remount,"+mode, "/").Output()
	return err
}
