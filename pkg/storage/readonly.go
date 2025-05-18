package storage

import (
	"io/fs"
)

type lockedReadOnlyFs struct {
	inner ReadOnlyFS
}

var _ ReadOnlyFS = (*lockedReadOnlyFs)(nil)

func LockedReadOnlyFs(fs ReadOnlyFS) *lockedReadOnlyFs {
	return &lockedReadOnlyFs{inner: fs}
}

func (fs *lockedReadOnlyFs) Open(name string) (fs.File, error) {
	return fs.inner.Open(name)
}

func (fs *lockedReadOnlyFs) Stat(name string) (FileMetaData, error) {
	return fs.inner.Stat(name)
}
func (fs *lockedReadOnlyFs) ReadDir(name string) ([]FileMetaData, error) {
	return fs.inner.ReadDir(name)
}
func (fs *lockedReadOnlyFs) Glob(pattern string) ([]string, error) {
	return fs.inner.Glob(pattern)
}
func (fs *lockedReadOnlyFs) Sub(name string) (ReadOnlyFS, error) {
	return fs.inner.Sub(name)
}
