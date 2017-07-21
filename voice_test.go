package discgo

import "testing"

func TestClient_GetVoiceRegions(t *testing.T) {
	regions, err := c.VoiceRegions().Get()
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range regions {
		t.Log(r.Name)
	}
}
