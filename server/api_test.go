package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPI(t *testing.T) {
	r, err := New()
	require.NoError(t, err)
	server := httptest.NewServer(r)
	defer server.Close()

	// create httpexpect instance
	e := httpexpect.New(t, server.URL)

	t.Run("basic client", func(t *testing.T) {
		state := e.GET("/state").
			Expect().
			Status(http.StatusOK).Text().Raw()
		assert.JSONEq(t, `{}`, state)

	})
}
