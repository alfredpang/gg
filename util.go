package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// other useful things

// go list -e -json output
type goListJson struct {
	Dir         string
	ImportPath  string
	Name        string
	Doc         string
	Target      string
	Root        string
	Gofiles     []string
	Imports     []string
	Deps        []string
	TestGoFiles []string
	TestImports []string
}

func getNowStr() string {
	const layout = "2006-01-02T15:04:05"
	t := time.Now()
	return t.Format(layout)
}

// github.com/jprobinson/go-imap/imap
// actually lives in the github.com/jprobinson/go-imap repo
// only works on github.com and bitbucket.org for now
//
// For other things, you will need to specify the dependencies individually
//
func trimPackageToRepo(p string) string {
	if !strings.HasPrefix(p, "github.com") && !strings.HasPrefix(p, "bitbucket.org") {
		return p
	}

	pparts := strings.Split(p, "/")
	if len(pparts) <= 3 {
		return p
	}

	return strings.Join(pparts[:3], "/")
}

func getCurrentGopath() (string, error) {
	currentenv := os.Environ()
	for _, envval := range currentenv {
		if strings.HasPrefix(envval, "GOPATH=") {
			return string(envval[7:]), nil
		}
	}
	return "", errors.New("GOPATH not found")
}

func getEnvWithNewGopath(newGopath string) []string {
	currentenv := os.Environ()
	// same env, but change GOPATH
	subenv := make([]string, len(currentenv))
	for i, envval := range currentenv {
		if strings.HasPrefix(envval, "GOPATH=") {
			subenv[i] = fmt.Sprintf("GOPATH=%s", newGopath)
		} else {
			subenv[i] = envval
		}
	}

	return subenv
}

// package/repo dependencies
func (cmd *ggcmd) rdepHelper(rpkg string, includeTestDeps bool) []string {
	tempdir, err := ioutil.TempDir("", "gg")
	if err != nil {
		ggFatal("Unable to create temp directory %s", err)
	}
	defer os.RemoveAll(tempdir)
	os.Mkdir(tempdir+"/src", os.ModePerm)

	subcmd := exec.Command("go", "get", rpkg)
	subcmd.Env = getEnvWithNewGopath(tempdir)
	_, err = subcmd.Output()
	if err != nil {
		ggFatal("Unable to get package specified err=%s", err)
	}

	subcmd = exec.Command("go", "list", "-e", "-json", rpkg)
	subcmd.Env = getEnvWithNewGopath(tempdir)
	goListJsonRaw, err := subcmd.Output()
	if err != nil {
		ggFatal("Unable call go list on package err=%s", err)
	}

	var pkgGoList goListJson
	err = json.Unmarshal(goListJsonRaw, &pkgGoList)
	if err != nil {
		ggFatal("Unable to parse json of go list err=%s", err)
	}

	hasSeen := map[string]string{}
	deps := make([]string, 0, len(pkgGoList.Deps)+len(pkgGoList.TestImports))
	for _, pkg := range pkgGoList.Deps {
		if cmd.corePkgsMap[pkg] == nil && !strings.HasPrefix(pkg, pkgGoList.ImportPath) {
			if hasSeen[pkg] != "" {
				continue
			}
			deps = append(deps, pkg)
			hasSeen[pkg] = pkg
		}
	}

	if !includeTestDeps {
		sort.Strings(deps)
		return deps
	}

	for _, pkg := range pkgGoList.TestImports {
		if cmd.corePkgsMap[pkg] == nil && !strings.HasPrefix(pkg, pkgGoList.ImportPath) {
			if hasSeen[pkg] != "" {
				continue
			}
			deps = append(deps, pkg)
			hasSeen[pkg] = pkg
		}
	}
	sort.Strings(deps)
	return deps
}

func (cmd *ggcmd) ldepHelper(goListArg string, includeTestDeps bool) []string {
	subcmd := exec.Command("go", "list", "-e", "-json", goListArg)
	goListJsonRaw, err := subcmd.Output()
	if err != nil {
		ggFatal("Unable call go list on package err=%s", err)
	}

	var pkgGoList goListJson
	err = json.Unmarshal(goListJsonRaw, &pkgGoList)
	if err != nil {
		ggFatal("Unable to parse json of go list err=%s", err)
	}

	hasSeen := map[string]string{}
	deps := make([]string, 0, len(pkgGoList.Deps)+len(pkgGoList.TestImports))
	for _, pkg := range pkgGoList.Deps {
		if cmd.corePkgsMap[pkg] == nil && !strings.HasPrefix(pkg, pkgGoList.ImportPath) {
			deps = append(deps, pkg)
			hasSeen[pkg] = pkg
		}
	}

	if !includeTestDeps {
		sort.Strings(deps)
		return deps
	}

	for _, pkg := range pkgGoList.TestImports {
		if cmd.corePkgsMap[pkg] == nil && !strings.HasPrefix(pkg, pkgGoList.ImportPath) {
			if hasSeen[pkg] != "" {
				continue
			}
			deps = append(deps, pkg)
			hasSeen[pkg] = pkg
		}
	}

	sort.Strings(deps)
	return deps
}

func (cmd *ggcmd) initCorePkgs() {
	cmd.corePkgs = []string{
		"archive/tar", "archive/zip", "bufio", "builtin",
		"bytes", "compress/bzip2", "compress/flate",
		"compress/gzip", "compress/lzw", "compress/zlib",
		"container/heap", "container/list", "container/ring",
		"crypto", "crypto/aes", "crypto/cipher", "crypto/des",
		"crypto/dsa", "crypto/ecdsa", "crypto/elliptic",
		"crypto/hmac", "crypto/md5", "crypto/rand",
		"crypto/rc4", "crypto/rsa", "crypto/sha1",
		"crypto/sha256", "crypto/sha512", "crypto/subtle",
		"crypto/tls", "crypto/x509", "crypto/x509/pkix",
		"database/sql", "database/sql/driver", "debug/dwarf",
		"debug/elf", "debug/gosym", "debug/macho", "debug/pe",
		"debug/plan9obj", "encoding", "encoding/ascii85",
		"encoding/asn1", "encoding/base32", "encoding/base64",
		"encoding/binary", "encoding/csv", "encoding/gob",
		"encoding/hex", "encoding/json", "encoding/pem",
		"encoding/xml", "errors", "expvar", "flag", "fmt",
		"go/ast", "go/build", "go/constant", "go/doc",
		"go/format", "go/importer", "go/internal/gcimporter",
		"go/parser", "go/printer", "go/scanner", "go/token",
		"go/types", "hash", "hash/adler32", "hash/crc32",
		"hash/crc64", "hash/fnv", "html", "html/template",
		"image", "image/color", "image/color/palette",
		"image/draw", "image/gif", "image/internal/imageutil",
		"image/jpeg", "image/png", "index/suffixarray", "io",
		"io/ioutil", "log", "log/syslog", "math", "math/big",
		"math/cmplx", "math/rand", "mime", "mime/multipart",
		"mime/quotedprintable", "net", "net/http",
		"net/http/cgi", "net/http/cookiejar", "net/http/fcgi",
		"net/http/httptest", "net/http/httputil",
		"net/http/internal", "net/http/pprof",
		"net/internal/socktest", "net/mail", "net/rpc",
		"net/rpc/jsonrpc", "net/smtp", "net/textproto",
		"net/url", "os", "os/exec", "os/signal", "os/user",
		"path", "path/filepath", "reflect", "regexp",
		"regexp/syntax", "runtime", "runtime/cgo",
		"runtime/debug", "runtime/pprof", "runtime/race",
		"sort", "strconv", "strings", "sync", "sync/atomic",
		"syscall", "testing", "testing/iotest",
		"testing/quick", "text/scanner", "text/tabwriter",
		"text/template", "text/template/parse", "time",
		"unicode", "unicode/utf16", "unicode/utf8", "unsafe"}

	cmd.corePkgsMap = map[string]*string{}

	for _, pkg := range cmd.corePkgs {
		cmd.corePkgsMap[pkg] = &pkg
	}
}
