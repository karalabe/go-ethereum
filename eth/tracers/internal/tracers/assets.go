// Code generated by go-bindata.
// sources:
// call_tracer.js
// noop_tracer.js
// opcount_tracer.js
// DO NOT EDIT!

package tracers

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

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

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _call_tracerJs = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x56\x4f\x6f\xdb\x3a\x12\x3f\x5b\x9f\x62\x7a\x8a\x8d\x3a\x72\xd2\x76\x73\x70\xa0\x2e\xb2\x8d\xb1\x5b\x20\xbb\x29\xba\x79\x7d\x87\x22\x07\x5a\x1a\xc9\x6c\x28\x52\x8f\x1c\xf9\x0f\xda\x7c\xf7\x87\x21\x29\x45\x56\xd2\x3e\xe0\xf9\x24\x0f\x67\x7e\x9c\xbf\xbf\xe1\x62\x01\xb9\x50\xea\xce\x8a\x1c\x2d\x48\x07\x02\xca\x56\x29\x58\x2b\xb3\xd3\x40\x56\x68\x27\x72\x92\xc6\x7f\xb3\x0a\x6d\x04\x01\xee\xf9\x1f\x39\x10\xba\x00\x8b\x8d\xb1\xfc\xad\x54\xb2\x58\x00\x6d\x10\xa4\x26\xb4\x5a\x28\x8f\xed\xa0\x16\x05\xc2\xfa\x00\x62\x08\x38\x07\xa1\x8c\xae\x60\x27\x69\x03\x42\x1f\xa0\x75\x58\xb6\x0a\xa4\x2e\x8d\xad\x05\xab\xa4\xc9\xf7\x64\xb2\x58\x80\xd4\x5b\x93\x7b\x89\x0b\x2e\x6a\x74\x84\x05\x14\x82\x04\x38\xb2\x6d\x4e\xad\x45\xc8\x8d\x26\x21\xb5\xd4\x15\xfb\x32\x72\xc4\x68\xef\xb2\xc7\x0b\x5e\xe1\x1e\xf3\x96\x61\xc6\x9e\xa5\xc9\x64\x70\xe3\x12\xbe\xde\xcf\x93\x27\x3b\x12\xf9\x03\x7b\xc1\xf0\x79\x6b\x2d\x6a\x02\x8b\x79\x6b\x9d\xdc\xa2\x57\x81\xa0\x63\x4a\xaf\xb3\xfa\xf2\xdf\x78\x55\x80\xee\x41\x06\xc0\x05\xba\x1c\x75\x81\x85\xcf\xf2\x83\x83\xdd\x06\x69\x83\x16\x76\x78\xb2\x45\xf8\xd6\x3a\x1a\xe8\x94\xd6\xd4\x20\x34\x98\x96\xb8\x20\x83\x1a\x49\x4d\xc6\x03\x0a\xfe\xd6\x68\xbd\x3f\x69\x32\xe9\x8d\x97\x50\x0a\xe5\x30\xde\xeb\x08\x1b\x8e\x85\xc3\x7d\x60\x64\x63\x01\xb7\x68\x0f\x60\x9a\xdc\x14\x18\xaa\xcd\x51\xf4\x41\xa0\x4b\x93\x09\xdb\x2d\xa1\x6c\xb5\xbf\x76\xaa\x4c\x35\x87\x62\x3d\x83\xef\xc9\x84\x61\x3f\x96\xbe\x46\x3b\xa8\x91\x36\xa6\x18\xd4\x8f\x2f\x5b\x23\x57\xa8\x30\x1a\xe7\x20\x8a\x02\xc8\x84\x5c\xf6\x99\x4b\x26\x13\x59\x02\xa3\xa6\xa6\x81\x2c\x83\x93\x0f\x57\x37\x37\x27\xf0\xe3\x07\x8c\x64\x1f\x6e\xaf\x57\x63\xf9\xf5\xea\x66\xf5\xef\xab\xbb\x95\xb7\x09\x2e\xb1\x4f\xff\x7f\x90\x8d\x6f\xb3\xc6\xe2\x69\x6e\xea\x46\x2a\x1c\x36\xd6\x1c\x68\x63\x1c\x82\xb0\x31\xdf\xa5\xd0\x79\x97\x07\xc7\x20\x5b\x61\xd9\xd7\x0c\xc8\x5c\x15\x85\x45\xe7\xbc\x8b\xde\xe5\xb4\x41\x7c\x98\x9e\xcf\xd2\x7f\x1d\x08\xdd\x74\x36\xbb\x64\x0b\x8e\x42\xba\x4f\x16\xe3\x7d\xc5\x94\xcc\x2c\xba\x34\xb1\x48\xad\xd5\xfc\xf9\x18\x5d\xbc\xe5\x92\xef\xa4\x43\xa8\x84\xaf\xbe\x33\xf5\x68\x96\xa0\x40\x12\x52\xf5\xfe\x98\xb2\x84\xec\x28\x55\x47\xe1\xc3\x3f\xe1\x0c\x96\x70\x3e\xbb\x4c\x3a\x0b\xa9\x6f\xbd\xcd\xc8\xf5\x37\xf0\x9a\xc1\x66\xe9\x47\x4d\x17\xef\xa6\xc1\xff\xa0\xbf\xd2\x05\x64\xd1\xee\xf5\xd8\xee\xed\x33\xbb\x18\xcc\x95\x73\x58\xaf\x15\x3e\xe7\x83\x48\x18\x9e\x3b\x1c\x19\x8b\xbe\xed\x38\x45\x0a\xb9\x14\xdd\xcd\x5e\x37\x8b\xd9\xa2\x43\x83\x4b\x00\x88\x95\x9e\x7b\x21\x0f\x42\x2f\x14\x79\x6e\x5a\x4d\xe1\x84\x8c\x97\x03\x90\x09\x02\xa9\x9b\x96\x96\x5e\xf0\x1f\xdc\xfb\x84\xd5\x58\x1b\x7b\x48\x9d\x92\x39\x4e\x7d\x78\xf3\x10\xed\x6c\x16\x6c\x2a\xe1\xae\x65\x59\x2e\x3d\x7c\x25\x1c\x9c\xfa\xaf\xdc\xb8\x78\x8b\x69\xe9\x96\xcf\xc7\x49\x79\x37\x4a\x4a\xaf\x7d\x83\xfa\xb9\xf6\x3f\x46\xda\xbe\x25\xfa\xfe\x89\xa5\x7d\xf5\x93\xce\xf6\x74\x92\x6e\x85\x6a\x11\x32\x38\x39\xdb\x9f\x3c\xaf\xd1\x9b\x59\x7a\x87\x7b\x9a\x9e\x5f\x84\xb2\xfa\x7e\xa3\x8d\x74\x69\xcf\x45\x69\xd3\xba\xcd\x94\xff\x06\x15\x7f\xfa\xc4\x38\x19\x90\x6d\x31\xe9\xbb\x96\x75\x1e\xfb\x59\x7f\x91\xa4\x98\x8a\x8e\x69\x68\x0e\x16\xc9\x4a\xdc\x22\x48\x3a\x71\x1e\x92\x89\xda\xec\x84\xce\x31\x85\xdf\x31\x20\x6a\x44\x4f\x0a\x71\xc9\x80\x2c\x03\xe3\xf1\xa2\x90\xfa\x89\x2b\x84\xe7\x60\x8b\x50\x8b\x03\xac\x91\xe9\xe8\xe1\x00\x5c\xa7\xe2\xa0\x45\x2d\x73\x17\xf0\xfc\x82\xb1\x58\x09\xeb\x61\x2d\xfe\xd1\x86\xf5\xc1\x0d\x28\x72\x6a\x85\x52\x07\xa8\xe4\x16\xb5\xb7\x9e\xbe\x79\x7b\x76\x06\x8e\x64\x83\xba\x98\xc3\xc5\xdb\xc5\xc5\x3b\xb0\xad\xc2\x59\x1a\x99\xe9\x38\x3b\xb1\x12\x5d\xb1\x0a\x6c\x68\x03\xef\x61\x94\x60\x85\xba\xa2\x4d\x57\xb5\xe3\xc3\xaf\x2f\xea\xc2\x29\x9c\xdf\xfb\xbe\xcb\xba\x0e\x0c\xe5\x03\x54\x0e\xff\x26\x10\xc7\x76\x09\x8b\x05\xdc\xdd\x5e\xdf\xc2\xf4\x41\x58\xa1\xc4\x1a\x67\x4b\x58\xd9\x3a\x4d\xd3\xe3\xfe\x18\x76\x80\xdf\x1b\xc7\x85\x17\x1a\x70\x2f\x1d\x31\x9f\xfb\x92\x48\x07\xa1\x41\xa4\xae\xe6\xd0\x98\xc6\x13\xd4\x5f\xb1\xfb\xe7\xd5\x97\xd5\xe7\xbb\xae\xa3\x7d\x43\x5b\xde\x42\x14\xfb\xee\xf2\x79\xe3\x1d\x27\x3b\xcb\x7e\x99\xed\xc5\x02\x3e\x0d\x5c\x51\xc2\x51\xec\x20\x5d\x40\x85\x61\xc1\xf5\x2b\x1a\x2c\xba\x56\x91\x1b\xd1\xd0\x78\x5e\x4c\xd3\x91\x9d\x77\xb8\x12\xee\x37\xe7\xf3\x14\x67\x70\x2d\xab\xf4\x7f\xb8\xfb\xa8\x69\xda\x29\x44\xfe\xe0\xaf\xd7\xd0\x09\x99\x61\x46\xf3\xd9\xeb\xf3\xef\x57\x80\x23\xbb\x02\x15\x12\x1e\x21\x0f\x3c\x34\x2d\x35\xad\x4f\xe9\xcb\x0c\xd8\x29\x79\x1e\x1c\xfc\xe9\x7c\x0d\xf4\x35\x7b\x7e\x55\x50\x7b\x49\x7c\x83\xfa\x69\x1f\x34\x3c\x4e\xdd\xb2\x1f\xbc\x08\x94\x74\x7e\xc6\xc9\x34\xb5\x71\xd4\xa5\x5d\x61\x49\xcf\xd3\x1e\xea\x1a\x40\x7d\x0f\x78\xad\x0c\xce\x8e\x06\x6b\xb0\xd6\xc7\xc4\xf6\xc2\xda\xbd\x6a\x1a\xc1\xaf\x38\x75\x00\x49\xb0\x13\x83\x07\x66\xa0\x2d\xa9\xbf\x21\xf3\x90\x8e\xce\x37\x16\xb7\xd2\xb4\x0e\x8c\xc6\xce\x91\xd1\x1c\xb2\x5f\xa7\xe7\xf7\x41\x02\x59\x96\x41\xab\x0b\x2c\xa5\xee\xe9\x62\x3c\xb9\x23\x0b\xf8\x7a\xff\x33\xae\x3e\x56\x1d\x05\xf8\x98\x4c\x1e\xe3\xe3\x2e\xf4\xf1\xf0\x79\xb7\xdb\xa0\xee\x9f\xc6\xf1\x65\x03\x1b\xb1\x45\x58\x23\x6a\x90\x84\x56\x70\xd8\x66\x8b\x36\x3e\xed\x39\x59\xce\xc3\xb1\x4d\x29\x79\x89\x47\xe0\xf8\xbe\x65\x92\x96\xba\x4a\x93\x49\x90\x0f\xde\x85\x39\xed\x43\xb4\x5c\xd0\x68\x15\x17\x7a\xbf\xcf\x73\xda\xa7\xfc\xc7\xef\xc8\x7e\xa1\x3f\x3d\xb1\xf8\x9c\xc5\x61\x89\x0e\xf6\xfa\x50\x81\x4c\x38\xf6\x7b\x90\x35\xe2\xcc\xf0\x99\x97\xf5\x83\xe2\xd5\x2a\xe1\x02\xcc\x0b\xa3\x45\xfb\xe3\xc9\xea\x0c\x78\xba\x97\x3f\x37\xe0\xe3\x91\xd1\xe8\xc9\xc1\x8a\x5e\x14\x4e\xc3\x38\x2e\x87\xa7\x41\x14\x03\x95\xf5\x20\x3f\xb2\xf6\xf9\xf1\xaf\x82\xbe\xdb\x86\x5d\x1e\x89\xfe\x7d\x3f\x08\x21\xdb\x7d\x37\x8d\x0d\x86\x3c\xca\x37\xa0\xb5\xc6\xc2\xab\x17\xfa\x34\x02\x05\x85\x0c\x7a\xe5\x0e\x21\xb4\x48\xac\xee\x65\x32\x79\x4c\x1e\x93\x3f\x03\x00\x00\xff\xff\xba\x96\x49\x56\x55\x0e\x00\x00")

func call_tracerJsBytes() ([]byte, error) {
	return bindataRead(
		_call_tracerJs,
		"call_tracer.js",
	)
}

func call_tracerJs() (*asset, error) {
	bytes, err := call_tracerJsBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "call_tracer.js", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _noop_tracerJs = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x4c\xcf\xc1\x6a\x23\x31\x10\x04\xd0\xb3\xf5\x15\x75\xdc\x05\xe3\xb9\xef\x27\x2c\xec\x69\x43\xee\x2d\x4d\x8d\x47\x8e\xa2\x9e\xb4\x7a\x26\x0e\xc6\xff\x1e\x46\x26\xe0\x9b\x68\xa8\x57\xa5\x61\x40\x55\x5d\x5e\x4c\x12\x0d\xb9\xe1\xb2\x36\x87\xcf\x44\x14\x63\xd4\x4a\x44\xcd\x85\xb6\x14\x71\x22\xe9\x48\x18\x3f\xd6\x6c\x1c\x31\x99\xbe\x43\xf0\x57\x36\xf9\x9f\x2c\x2f\x1e\x86\x01\x1a\x2f\x4c\x0e\x57\x44\x62\x6d\x12\x0b\x21\x0d\x02\x37\xa9\x4d\x92\x67\xad\xfb\x3b\xd1\x4e\xe1\x16\x0e\xc3\x80\xe6\x5c\xf6\xee\x5c\x37\x7d\xdb\x5d\x35\x70\xa3\x7d\x41\x97\xde\xe8\xb3\x3c\x46\xbd\xfe\x03\xaf\x4c\xab\xb3\x9d\xc2\x61\xcf\xfd\xc1\xb4\xd6\x8e\xfe\x2a\x7a\x3e\x62\x8c\xbf\x71\xc3\xfd\x18\xba\x6c\x6c\x6b\xf1\x67\xfb\x73\x66\x85\x94\xd2\xb9\x07\xdf\x30\xcb\x46\x44\xb2\x22\x3b\x4d\x9c\x23\x74\xa3\x41\xea\x08\xa3\xaf\x56\x5b\xe7\xf6\xcc\x94\xab\x94\x1f\x58\xa7\x7e\xdb\xbf\x93\xeb\xf9\x14\x0e\x8f\xfb\xd3\xa8\xe4\xd7\x3e\x28\xdc\xc3\x77\x00\x00\x00\xff\xff\x8f\x9c\x5f\x55\x6c\x01\x00\x00")

func noop_tracerJsBytes() ([]byte, error) {
	return bindataRead(
		_noop_tracerJs,
		"noop_tracer.js",
	)
}

func noop_tracerJs() (*asset, error) {
	bytes, err := noop_tracerJsBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "noop_tracer.js", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _opcount_tracerJs = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x5c\x90\xb1\x6e\xeb\x30\x0c\x45\x67\xeb\x2b\xee\xf8\x1e\x12\xd8\x9d\xb3\x77\xcc\x56\x64\x97\x6d\x3a\x56\xe3\x50\x01\x49\xb9\x09\x82\xfc\x7b\x21\xb9\x2e\x8c\x8e\x22\xc8\x73\xee\x55\xd3\x20\xde\xba\x98\xd8\x3e\xc4\x77\x24\x08\x0a\x0f\xf5\xd7\xdb\x44\xb0\x65\x64\xa3\x37\x7c\x26\x35\x94\x45\x85\x8d\x04\x4e\xd7\x96\x04\x71\x40\x60\x35\x49\x9d\x85\xc8\xea\x9a\x06\x74\xa7\x2e\x19\xf5\x68\x1f\x65\xf3\xfd\x74\x44\x4b\x43\x14\x2a\x4f\x13\xcf\xea\xcb\x3a\x8c\xe4\x1a\xd8\x1b\xf5\xb5\x7b\xba\xaa\x69\x16\x43\x11\x5f\xfe\x7a\x32\x67\xeb\xfa\x15\xd5\xae\x2a\x67\x07\xbc\xed\x5d\xa1\xa8\xd1\x2d\x37\x09\x3c\xc7\x0b\xf5\x18\xa2\x80\x66\x92\x47\x29\xdb\xd3\x52\x29\xe3\x4f\xc7\x15\xa3\xb5\xab\xf2\xdd\x01\x43\xe2\x62\xf8\x37\xc5\xf3\x1e\x7d\xfb\x1f\x4f\xd8\x18\xb4\x2e\x96\xdd\x0e\xaf\x1f\x8d\x90\xa6\xc9\xb6\xa2\xaf\x91\x18\x7e\x9a\x0a\x7b\x71\x29\x46\x3f\x13\x5a\x22\x46\x30\x92\xdc\x16\x71\x26\x81\xe7\x1e\x42\x96\x84\xb5\xe0\xf2\xcd\x10\xd8\x4f\x2b\x38\x0e\xeb\x8f\x75\x81\xcf\xb5\xab\x96\xf9\x26\x61\x67\xf7\x9c\x6e\xa1\x6c\x42\xe2\xe5\x5e\xee\x3b\x00\x00\xff\xff\x6e\xdf\xbf\xab\xdc\x01\x00\x00")

func opcount_tracerJsBytes() ([]byte, error) {
	return bindataRead(
		_opcount_tracerJs,
		"opcount_tracer.js",
	)
}

func opcount_tracerJs() (*asset, error) {
	bytes, err := opcount_tracerJsBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "opcount_tracer.js", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"call_tracer.js":    call_tracerJs,
	"noop_tracer.js":    noop_tracerJs,
	"opcount_tracer.js": opcount_tracerJs,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"call_tracer.js":    {call_tracerJs, map[string]*bintree{}},
	"noop_tracer.js":    {noop_tracerJs, map[string]*bintree{}},
	"opcount_tracer.js": {opcount_tracerJs, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
