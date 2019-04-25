// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dav

import (
	"context"
	"io"
	"net/http"
	"os"
	"path"
)

// SlashClean is equivalent to but slightly more efficient than
// path.Clean("/" + name).
func SlashClean(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}

// A FileSystem implements access to a collection of named files. The elements
// in a file path are separated by slash ('/', U+002F) characters, regardless
// of host operating system convention.
//
// Each method has the same semantics as the os package's function of the same
// name.
//
// Note that the os.Rename documentation says that "OS-specific restrictions
// might apply". In particular, whether or not renaming a file or directory
// overwriting another existing file or directory is an error is OS-dependent.
type FileSystem interface {
	Mkdir(ctx context.Context, name string, perm os.FileMode) error
	OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (File, error)
	RemoveAll(ctx context.Context, name string) error
	Rename(ctx context.Context, oldName, newName string) error
	Stat(ctx context.Context, name string) (os.FileInfo, error)
}

// A File is returned by a FileSystem's OpenFile method and can be served by a
// Handler.
//
// A File may optionally implement the DeadPropsHolder interface, if it can
// load and save dead properties.
type File interface {
	http.File
	io.Writer
}
