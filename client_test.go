package discgo_test

import (
	"github.com/nhooyr/discgo"
	"net/http"
	"os"
	"testing"
)

var c *discgo.Client

func init() {
	c = discgo.NewClient()
	// TODO maybe make it all a simple struct like acme's autocert manager but that's so magical :(
	c.Token = "Bot " + os.Getenv("DISCORD_TOKEN")
}

func TestClient_APIError(t *testing.T) {
	c := discgo.NewClient()
	_, err := c.GetMyConnections()
	if err == nil {
		t.Fatal("expected non nil error")
	}
	apiErr, ok := err.(*discgo.APIError)
	if !ok {
		t.Fatal("expected error to be of type *discgo.APIError")
	}
	if apiErr.Response.StatusCode != http.StatusUnauthorized {
		t.Fatal("expected %v but got %v", http.StatusUnauthorized, apiErr.Response.StatusCode)
	}
	if apiErr.JSON.Code != 0 {
		t.Fatal("expected %v but got %v", 0, apiErr.JSON.Code)
	}
	if apiErr.JSON.Message != "401: Unauthorized" {
		t.Fatal("expected %v but got %v", 0, apiErr.JSON.Message)
	}
}
