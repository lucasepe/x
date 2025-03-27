package memfs

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

var fsDef = map[string][]byte{
	"/root/home/foo/1.txt": []byte("@/root/home/foo/1.txt"),
	"/root/home/bar/1.txt": []byte("@/root/home/bar/1.txt"),
	"/root/home/1.txt":     []byte("@/root/home/1.txt"),
	"/root/home/xyz/":      nil,
	"/etc/1.txt":           []byte("@/etc/1.txt"),
	"/1.txt":               []byte("@/1.txt"),
}

func names(files []*File) []string {
	names := []string{}
	for _, f := range files {
		names = append(names, f.name)
	}
	sort.Strings(names)
	return names
}

func TestFs(t *testing.T) {
	fs, err := New(fsDef)

	assert.NoError(t, err)
	assert.NotNil(t, fs)

	// Open
	of, err := fs.Open("/1.txt")
	assert.NoError(t, err)
	f := of.(*File)
	assert.Equal(t, "1.txt", f.name)
	assert.Equal(t, []byte("@/1.txt"), f.bytes)
	assert.Nil(t, f.children)

	of, err = fs.Open("root/home/1.txt")
	assert.NoError(t, err)
	f = of.(*File)
	assert.Equal(t, "1.txt", f.name)
	assert.Equal(t, []byte("@/root/home/1.txt"), f.bytes)
	assert.Nil(t, f.children)

	of, err = fs.Open("/root/home/xyz")
	assert.NoError(t, err)
	f = of.(*File)
	assert.Equal(t, "xyz", f.name)
	assert.Nil(t, f.bytes)
	assert.Nil(t, f.children)

	of, err = fs.Open("/root/home/2.txt")
	assert.Error(t, err)
	assert.Nil(t, of)

	of, err = fs.Open("/")
	assert.NoError(t, err)
	d := of.(*File)
	assert.Equal(t, "/", d.name)
	assert.Nil(t, d.bytes)
	assert.Equal(t, []string{"1.txt", "etc", "root"}, names(d.children))

	of, err = fs.Open("root/home")
	assert.NoError(t, err)
	d = of.(*File)
	assert.Equal(t, "home", d.name)
	assert.Nil(t, d.bytes, nil)
	assert.Equal(t, []string{"1.txt", "bar", "foo", "xyz"}, names(d.children))

	fi, err := d.Stat()
	assert.NoError(t, err)
	assert.True(t, fi.IsDir())
	assert.EqualValues(t, 0, fi.Size())

	of, err = fs.Open("/root/home/")
	assert.NoError(t, err)
	d = of.(*File)
	assert.Equal(t, "home", d.name)
	assert.Nil(t, d.bytes, nil)
	assert.Equal(t, []string{"1.txt", "bar", "foo", "xyz"}, names(d.children))

	fi, err = d.Stat()
	assert.NoError(t, err)
	assert.True(t, fi.IsDir())
	assert.EqualValues(t, 0, fi.Size())
}
