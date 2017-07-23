package discgo

import (
	"net/http"
	"os"
	"testing"
)

var client *Client

func init() {
	client = NewClient()
	// TODO maybe make it all a simple struct like acme's autocert manager but that's so magical :(
	client.Token = os.Getenv("DISCORD_TOKEN")
}

func TestClient_APIError(t *testing.T) {
	c := NewClient()
	_, err := c.Me().Connections().Get()
	if err == nil {
		t.Fatal("expected non nil error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatal("expected error to be of type *discgo.APIError")
	}
	if apiErr.JSON == nil {
		t.Fatal("expected non nil apiErr.JSON")
	}
	if apiErr.Response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected %v but got %v", http.StatusUnauthorized, apiErr.Response.StatusCode)
	}
	if apiErr.JSON.Code != 0 {
		t.Fatalf("expected %v but got %v", 0, apiErr.JSON.Code)
	}
	if apiErr.JSON.Message != "401: Unauthorized" {
		t.Fatalf("expected %v but got %v", 0, apiErr.JSON.Message)
	}
}
