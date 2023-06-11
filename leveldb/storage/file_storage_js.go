// Copyright (c) 2012, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build js && wasm
// +build js,wasm

package storage

import (
	"os"
	"syscall"
)

type unixFileLock struct {
}

func (fl *unixFileLock) release() error {
	return nil
}

func newFileLock(path string, readOnly bool) (fl fileLock, err error) {
	var flag int
	if readOnly {
		flag = os.O_RDONLY
	} else {
		flag = os.O_RDWR
	}
	f, err := os.OpenFile(path, flag, 0)
	if os.IsNotExist(err) {
		f, err = os.OpenFile(path, flag|os.O_CREATE, 0644)
	}
	if err != nil {
		return
	}
	defer f.Close()
	fl = &unixFileLock{}
	return
}

func rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func isErrInvalid(err error) bool {
	if err == os.ErrInvalid {
		return true
	}
	// Go < 1.8
	if syserr, ok := err.(*os.SyscallError); ok && syserr.Err == syscall.EINVAL {
		return true
	}
	// Go >= 1.8 returns *os.PathError instead
	if patherr, ok := err.(*os.PathError); ok && patherr.Err == syscall.EINVAL {
		return true
	}
	return false
}

func syncDir(name string) error {
	// As per fsync manpage, Linux seems to expect fsync on directory, however
	// some system don't support this, so we will ignore syscall.EINVAL.
	//
	// From fsync(2):
	//   Calling fsync() does not necessarily ensure that the entry in the
	//   directory containing the file has also reached disk. For that an
	//   explicit fsync() on a file descriptor for the directory is also needed.
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.Sync(); err != nil && !isErrInvalid(err) {
		return err
	}
	return nil
}
