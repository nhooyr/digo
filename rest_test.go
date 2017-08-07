package discgo

import (
	"context"
	"net/http"
	"os"
	"testing"
)

var (
	client = &RESTClient{
		Token: os.Getenv("DISCORD_TOKEN"),
	}
	ctx = context.Background()
)

func TestClient_APIError(t *testing.T) {
	c := new(RESTClient)
	_, err := c.Me().Connections().Get(ctx)
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
