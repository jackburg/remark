package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	flags "github.com/jessevdk/go-flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackup_Execute(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.Path, "/api/v1/admin/export")
		assert.Equal(t, "GET", r.Method)
		fmt.Fprint(w, "blah\nblah2\n12345678\n")
	}))
	defer ts.Close()

	cmd := BackupCommand{}
	p := flags.NewParser(&cmd, flags.Default)
	_, err := p.ParseArgs([]string{"--secret=123456", "--site=remark", "--path=/tmp",
		"--file={{.SITE}}-test.export", "--url=" + ts.URL})
	require.Nil(t, err)
	err = cmd.Execute(nil)
	assert.NoError(t, err)
	defer os.Remove("/tmp/remark-test.export")

	data, err := ioutil.ReadFile("/tmp/remark-test.export")
	require.Nil(t, err)
	assert.Equal(t, "blah\nblah2\n12345678\n", string(data))
}

func TestBackup_ExecuteFailedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.Path, "/api/v1/admin/export")
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(400)
		fmt.Fprint(w, "some error")
	}))
	defer ts.Close()

	cmd := BackupCommand{}
	p := flags.NewParser(&cmd, flags.Default)
	_, err := p.ParseArgs([]string{"--secret=123456", "--site=remark", "--path=/tmp",
		"--file={{.SITE}}-test.export", "--url=" + ts.URL})
	require.Nil(t, err)
	err = cmd.Execute(nil)
	assert.EqualError(t, err, `error response "400 Bad Request", some error`)
}

func TestBackup_ExecuteFailedWrite(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.Path, "/api/v1/admin/export")
		assert.Equal(t, "GET", r.Method)
		fmt.Fprint(w, "blah\nblah2\n12345678\n")
	}))
	defer ts.Close()

	cmd := BackupCommand{}
	p := flags.NewParser(&cmd, flags.Default)
	_, err := p.ParseArgs([]string{"--secret=123456", "--site=remark", "--path=/tmp",
		"--file=/tmp/no-such-dir/{{.SITE}}-test.export", "--url=" + ts.URL})
	require.Nil(t, err)
	err = cmd.Execute(nil)
	assert.EqualError(t, err, `can't create backup file /tmp/no-such-dir/remark-test.export: open /tmp/no-such-dir/remark-test.export: no such file or directory`)
}
