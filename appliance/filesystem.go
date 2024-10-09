/**
 * This software and associated documentation files (the “Software”),
 * including GFI AppManager, is the property of GFI USA, LLC and its affiliates.
 * No part of the Software may be copied, modified, distributed, sold, or otherwise
 * used except as expressly permitted by the terms of the software license agreement.
 */

package appliance

import (
	"io"
	"os"
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

func (F *FileSystem) RemoveFile(path string) error {
	return os.Remove(path)
}
