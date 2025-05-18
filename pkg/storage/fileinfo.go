package storage

import (
	"io/fs"
	"time"
)

type readOnlyFileMetaData struct {
	inner FileMetaData
}

func (f *readOnlyFileMetaData) Name() string       { return f.inner.Name() }
func (f *readOnlyFileMetaData) Size() int64        { return f.inner.Size() }
func (f *readOnlyFileMetaData) Mode() fs.FileMode  { return f.inner.Mode() }
func (f *readOnlyFileMetaData) ModTime() time.Time { return f.inner.ModTime() }
func (f *readOnlyFileMetaData) IsDir() bool        { return f.inner.IsDir() }
func (f *readOnlyFileMetaData) Sys() any           { return nil }
func (f *readOnlyFileMetaData) Metadata(k string) string {
	return f.inner.Metadata(k)
}

type fileMetaData struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	meta    map[string]string
}

var _ FileMetaData = (*fileMetaData)(nil)

func (f *fileMetaData) Name() string       { return f.name }
func (f *fileMetaData) Size() int64        { return f.size }
func (f *fileMetaData) Mode() fs.FileMode  { return f.mode }
func (f *fileMetaData) ModTime() time.Time { return f.modTime }
func (f *fileMetaData) IsDir() bool        { return f.mode.IsDir() }
func (f *fileMetaData) Sys() any           { return nil }
func (f *fileMetaData) Metadata(k string) string {
	return f.meta[k]
}

func (f *fileMetaData) SetMetadata(k, v string) {
	if f.meta == nil {
		f.meta = make(map[string]string)
	}
	f.meta[k] = v
}
func (f *fileMetaData) Clone() FileMetaData {
	return &fileMetaData{
		name:    f.name,
		size:    f.size,
		mode:    f.mode,
		modTime: f.modTime,
		meta:    f.meta,
	}
}
func (f *fileMetaData) GetReadOnlyWrapper() FileMetaData {
	return &readOnlyFileMetaData{inner: f}
}

func NewFileMetaData(name string, size int64, mode fs.FileMode, modTime time.Time) *fileMetaData {
	return &fileMetaData{
		name:    name,
		size:    size,
		mode:    mode,
		modTime: modTime,
	}
}
