package fs_test

import (
	"io"
	"io/fs"
	"os"
	"testing"

	wanixfs "tractor.dev/wanix/fs"
)

// stickyFS mimics FSA OpenFile behavior before the O_TRUNC fix:
// Create does not clear bytes; Open+Write overwrites from offset 0 without
// shrinking; Truncate actually changes length.
type stickyFS struct {
	data          map[string][]byte
	truncateCalls int
}

func (s *stickyFS) Open(name string) (fs.File, error) {
	if _, ok := s.data[name]; !ok {
		return nil, fs.ErrNotExist
	}
	return &stickyFile{fsys: s, name: name}, nil
}

func (s *stickyFS) Create(name string) (fs.File, error) {
	if _, ok := s.data[name]; !ok {
		s.data[name] = nil
	}
	return &stickyFile{fsys: s, name: name}, nil
}

func (s *stickyFS) Chmod(name string, mode fs.FileMode) error {
	return nil
}

func (s *stickyFS) Truncate(name string, size int64) error {
	s.truncateCalls++
	b, ok := s.data[name]
	if !ok {
		return fs.ErrNotExist
	}
	if size < 0 {
		return fs.ErrInvalid
	}
	if size > int64(len(b)) {
		nb := make([]byte, size)
		copy(nb, b)
		s.data[name] = nb
		return nil
	}
	s.data[name] = append([]byte(nil), b[:size]...)
	return nil
}

type stickyFile struct {
	fsys   *stickyFS
	name   string
	offset int64
}

func (f *stickyFile) Stat() (fs.FileInfo, error) { return nil, fs.ErrInvalid }
func (f *stickyFile) Read([]byte) (int, error)   { return 0, io.EOF }
func (f *stickyFile) Close() error               { return nil }

func (f *stickyFile) Write(p []byte) (int, error) {
	data := f.fsys.data[f.name]
	end := int(f.offset) + len(p)
	if end > len(data) {
		nb := make([]byte, end)
		copy(nb, data)
		data = nb
	}
	copy(data[f.offset:], p)
	f.fsys.data[f.name] = data
	f.offset += int64(len(p))
	return len(p), nil
}

func TestOpenFileOTruncAfterChmodShrinks(t *testing.T) {
	long := []byte("{\n  \"lastRemoteRev\": 172,\n  \"lastZedcafeRev\": 169\n}")
	short := []byte("{\n  \"lastRemoteRev\": 158,\n  \"lastZedcafeRev\": 53\n}")
	if len(long) <= len(short) {
		t.Fatalf("fixture sizes: long=%d short=%d (need long > short)", len(long), len(short))
	}

	s := &stickyFS{data: map[string][]byte{"state.json": append([]byte(nil), long...)}}

	f, err := wanixfs.OpenFile(s, "state.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	if _, err := wanixfs.Write(f, short); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	got := s.data["state.json"]
	if string(got) != string(short) {
		t.Fatalf("after shrink write:\n got %q (%d bytes)\nwant %q (%d bytes)", got, len(got), short, len(short))
	}
	if s.truncateCalls < 1 {
		t.Fatalf("Truncate calls=%d want >= 1 (O_TRUNC after chmod reopen)", s.truncateCalls)
	}
}
