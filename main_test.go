package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_main(t *testing.T) {
	mux := configureEndpoints()
	srv := httptest.NewServer(mux)
	t.Run("hello", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/hello")
		assert.NoError(t, err)
		bytes, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "Hello World!", string(bytes))
	})
	t.Run("ping", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/ping")
		assert.NoError(t, err)
		bytes, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "{\"message\":\"pong\"}", string(bytes))
	})
	t.Run("serie", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/serie")
		assert.NoError(t, err)
		bytes, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "Serie", string(bytes))
	})
}
