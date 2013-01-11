/*
The MIT License (MIT)

Copyright (c) 2012 Cory Martin

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
'Software'), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	options := processArgs()
	parts := getFilenameParts(options.file)
	newfilename := createNewFilename(options, parts)

	if _, err := copyFile(newfilename, parts.abs); err != nil {
		log.Fatalf(`Error writing file: "%s"`, newfilename)
	} else {
		log.Printf(`File created: "%s"`, filepath.Base(newfilename))
	}
}

type configOptions struct {
	file     string
	outdir   string
	prefix   string
	noprefix bool
}

type fileParts struct {
	name string
	abs  string
	dir  string
	base string
	ext  string
}

func processArgs() (options configOptions) {
	flag.StringVar(&options.file, "file", "", "File to bust.")
	flag.StringVar(&options.outdir, "outdir", "", "Directory to write file. Default is same as file.")
	flag.StringVar(&options.prefix, "prefix", "", "Prefix of fingerprinted file. Default is the basename of file.")
	flag.BoolVar(&options.noprefix, "noprefix", false, "Do not prefix filename with original file base name.")
	flag.Parse()

	if options.file == "" {
		log.Fatalf("Error: file not specified")
	}

	return options
}

func getFilenameParts(file string) (parts fileParts) {
	if _, err := os.Stat(file); err != nil {
		log.Fatalf(`File does not exist: "%s"`, file)
	}
	parts.abs, _ = filepath.Abs(file)
	parts.dir = filepath.Dir(parts.abs)
	parts.ext = filepath.Ext(parts.abs)
	parts.name = filepath.Base(parts.abs)
	parts.base = strings.TrimRight(parts.name, parts.ext)
	return parts
}

func md5hash(contents string) string {
	h := md5.New()
	io.WriteString(h, contents)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func createNewFilename(options configOptions, parts fileParts) string {
	// Read file and MD5 hash its contents
	contents, err := ioutil.ReadFile(parts.abs)
	if err != nil {
		log.Fatalf(`Error reading file: "%s"`, options.file)
	}
	hash := md5hash(string(contents))

	newfilename := hash + parts.ext
	if !options.noprefix {
		if options.prefix == "" {
			options.prefix = parts.base
		}
		newfilename = options.prefix + "-" + newfilename
	}
	if options.outdir != "" {
		newfilename = path.Join(options.outdir, newfilename)
	} else {
		newfilename = path.Join(parts.dir, newfilename)
	}

	return newfilename
}

func copyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := os.Create(dstName)
	if err != nil {
		return
	}
	defer dst.Close()

	return io.Copy(dst, src)
}
