package storage

type cache struct {
	writeable WritableFS
	readonly  ReadOnlyFS
}

var _ WritableFS = (*cache)(nil)

func NewWritableCache(inner WritableFS) WritableFS {
	return &cache{readonly: inner, writeable: inner}
}

func NewReadOnlyCache(inner WritableFS) ReadOnlyFS {
	c := &cache{readonly: inner}
	return LockedReadOnlyFs(c)
}

//--------------------------------------------------------------------
