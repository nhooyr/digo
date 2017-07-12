package discgo_test

import "testing"

func TestClient_GetVoiceRegions(t *testing.T) {
	regions, err := c.GetVoiceRegions()
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range regions {
		t.Log(r.Name)
	}
}