package main

import (
	"go/parser"
	"github.com/golang/glog"
	"go/token"
	"flag"
	"strings"
	"io/ioutil"
	"fmt"
	"path"
	"github.com/kopeio/gladish/pkg/sets"
)

func main() {
	flag.Parse()

	dir := "v/k8s.io/kubernetes"

	package(default_visibility = ["//visibility:public"])

	load("@io_bazel_rules_go//go:def.bzl", "go_prefix", "go_library")

	go_prefix("github.com/jmespath/go-jmespath")

	go_library(
		name = "go_default_library",
	srcs = glob(
		include = ["*.go"],
	exclude = ["*_test.go"],
),
)

	err := visitDir("k8s.io/kubernetes/", "", dir)
	if err != nil {
		glog.Fatalf("unexpected error: %v", err)
	}
}

func visitDir(projectPath, relativePath, dir string) error {
	stdlib := sets.NewString(
		"bufio",
		"bytes",
		"compress/gzip",
		"container/heap",
		"container/list",
		"container/ring",
		"crypto/md5",
		"crypto/rand",
		"crypto/rsa",
		"crypto/sha256",
		"crypto/tls",
		"crypto/x509",
		"crypto/x509/pkix",
		"encoding/base32",
		"encoding/base64",
		"encoding/binary",
		"encoding/csv",
		"encoding/hex",
		"encoding/json",
		"encoding/pem",
		"encoding/xml",
		"errors",
		"flag",
		"fmt",
		"go/ast",
		"go/doc",
		"go/format",
		"go/parser",
		"go/token",
		"hash",
		"hash/adler32",
		"hash/crc64",
		"hash/fnv",
		"io",
		"io/ioutil",
		"log",
		"math",
		"math/big",
		"math/rand",
		"mime",
		"mime/multipart",
		"net",
		"net/http",
		"net/http/httptest",
		"net/http/httputil",
		"net/http/pprof",
		"net/url",
		"os",
		"os/exec",
		"os/signal",
		"os/user",
		"path",
		"path/filepath",
		"reflect",
		"regexp",
		"runtime",
		"runtime/debug",
		"runtime/pprof",
		"sort",
		"strconv",
		"strings",
		"sync",
		"sync/atomic",
		"syscall",
		"testing",
		"text/tabwriter",
		"text/template",
		"time",
		"unicode",
		"unicode/utf8",
	)

	hasGoCode, subdirs, imports, testimports, err := getImports(dir)
	if err != nil {
		return fmt.Errorf("error parsing dir %q: %v", dir, err)
	}
	if hasGoCode {
		fmt.Printf("go_library(\n")
		fmt.Printf("\tname = %q,\n", relativePath)
		fmt.Printf("\tsrcs = glob(\n")
		fmt.Printf("\t\tinclude = [%q],\n", relativePath + "/*.go")
		fmt.Printf("\t\texclude = [%q]),\n", relativePath + "/*_test.go")
		fmt.Printf("\tdeps = [\n")
		for _, i := range imports.List() {
			if stdlib.Has(i) {
				continue
			}
			if strings.HasPrefix(i, projectPath) {
				// Relative path when in same project
				fmt.Printf("\t\t%q,\n", ":" + i[len(projectPath):])
			} else {
				fmt.Printf("\t\t%q,\n", "//vendor/" + i)
			}
		}
		for _, i := range testimports.List() {
			if stdlib.Has(i) {
				continue
			}
			//glog.Infof("testimport %s", i)
		}
		fmt.Printf("\t],\n")
		fmt.Printf(")\n\n")
	}

	for _, subdir := range subdirs {
		err := visitDir(projectPath, path.Join(relativePath, subdir), path.Join(dir, subdir))
		if err != nil {
			return err
		}
	}

	return nil
}

func getImports(dir string) (bool, []string, sets.String, sets.String, error) {
	var subdirs []string
	imports := sets.NewString()
	testimports := sets.NewString()

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return false, nil, nil, nil, fmt.Errorf("error reading directory %q: %v", dir, err)
	}

	hasGoCode := false
	for _, f := range files {
		name := f.Name()
		if f.IsDir() {
			subdirs = append(subdirs, name)
			continue
		}
		if !strings.HasSuffix(name, ".go") {
			glog.V(4).Infof("skipping file %s", name)
			continue
		}

		isTest := false
		if strings.HasSuffix(name, "_test.go") {
			isTest = true
		}
		glog.V(4).Infof("visiting file %s", name)

		mode := parser.ImportsOnly
		fset := token.NewFileSet()
		p := path.Join(dir, f.Name())
		f, err := parser.ParseFile(fset, p, nil, mode)
		if err != nil {
			return false, nil, nil, nil, fmt.Errorf("error parsing %q: %v", p, err)
		}

		hasGoCode = true
		// Print the imports from the file's AST.
		for _, s := range f.Imports {
			i := s.Path.Value
			if len(i) > 2 && i[0] == '"' && i[len(i) - 1] == '"' {
				i = i[1:len(i) - 1]
			}
			if isTest {
				testimports.Insert(i)
			} else {
				imports.Insert(i)
			}
		}
	}

	return hasGoCode, subdirs, imports, testimports, nil
}