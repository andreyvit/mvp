package deploymentutil

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"unicode/utf8"
)

type User struct {
	Username string
	Uid      int
	Gid      int
}

var Root = &User{Uid: 0, Gid: 0, Username: "root"}

func NeedUser(username string) *User {
	u, err := user.Lookup(username)
	if err != nil || u == nil {
		log.Fatalf("user %s not found: %v", username, err)
	}
	return &User{Username: u.Username, Uid: atoi(u.Uid), Gid: atoi(u.Gid)}
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		panic(err)
	}
}

func Install(path string, content []byte, perm os.FileMode, user *User) {
	log.Printf("▸ %s", path)
	if utf8.Valid(content) {
		indent := strings.Repeat(" ", 12)
		log.Println(indent + strings.TrimSpace(strings.ReplaceAll(string(content), "\n", "\n"+indent)))
		log.Println()
	}
	temp := must(os.CreateTemp(filepath.Dir(path), ".~"+filepath.Base(path)+".*"))
	ensure(temp.Chmod(perm))
	ensure(temp.Chown(user.Uid, user.Gid))
	must(temp.Write(content))
	ensure(temp.Close())
	ensure(os.Rename(temp.Name(), path))
}

func InstallDir(path string, perm os.FileMode, user *User) string {
	log.Printf("▸ %s/", path)
	ensureSkippingOSExists(os.Mkdir(path, perm))
	ensure(os.Chown(path, user.Uid, user.Gid))
	return path
}

func InstallSubdir(parent, name string, perm os.FileMode, user *User) string {
	return InstallDir(filepath.Join(parent, name), perm, user)
}

func InstallIfNotExistsFromStdin(path string, message string, perm os.FileMode, user *User) {
	if Exists(path) {
		return
	}
	message = strings.ReplaceAll(message, "{{.path}}", path)
	log.Print(message)

	var keyringText bytes.Buffer
	for reader := bufio.NewReader(os.Stdin); ; {
		fmt.Fprintf(os.Stderr, "> ")
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			keyringText.Write(line)
		}
		if err == io.EOF || len(bytes.TrimSpace(line)) == 0 {
			break
		}
		ensure(err)
	}
	Install(path, keyringText.Bytes(), 0600, user)
}

func InstallIfNotExistsFromString(path string, data string, perm os.FileMode, user *User) {
	if Exists(path) {
		return
	}
	Install(path, []byte(data), perm, user)
}

func Templ(templ string, data map[string]any) []byte {
	t := template.New("")
	_ = must(t.Parse(templ))
	var buf bytes.Buffer
	ensure(t.Execute(&buf, data))
	return buf.Bytes()
}

func Run(name string, args ...string) {
	log.Printf("▸ %s", shellQuoteCmdline(name, args...))
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("** %s failed: %v\n%s", name, err, output)
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func ensure(err error) {
	if err != nil {
		panic(err)
	}
}

func ensureSkippingOSExists(err error) {
	if os.IsExist(err) {
		return
	}
	if err != nil {
		panic(err)
	}
}

func atoi(s string) int {
	return must(strconv.Atoi(s))
}

func shellQuoteCmdline(name string, args ...string) string {
	var buf strings.Builder
	buf.WriteString(shellQuote(name))
	for _, arg := range args {
		buf.WriteByte(' ')
		buf.WriteString(shellQuote(arg))
	}
	return buf.String()
}

func shellQuote(source string) string {
	const specialChars = "\\'\"`${[|&;<>()*?! \t\n~"
	const specialInDouble = "$\\\"!"

	var buf strings.Builder
	buf.Grow(len(source) + 10)

	// pick quotation style, preferring single quotes
	if !strings.ContainsAny(source, specialChars) && len(source) > 0 {
		buf.WriteString(source)
	} else if !strings.ContainsRune(source, '\'') {
		buf.WriteByte('\'')
		buf.WriteString(source)
		buf.WriteByte('\'')
	} else {
		buf.WriteByte('"')
		for {
			i := strings.IndexAny(source, specialInDouble)
			if i < 0 {
				break
			}
			buf.WriteString(source[:i])
			buf.WriteByte('\\')
			buf.WriteByte(source[i])
			source = source[i+1:]
		}
		buf.WriteString(source)
		buf.WriteByte('"')
	}
	return buf.String()
}
