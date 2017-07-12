package discgo

import "path"

type VoiceState struct {
	GuildID   *string `json:"guild_id"`
	ChannelID string  `json:"channel_id"`
	UserID    string  `json:"user_id"`
	SessionID string  `json:"session_id"`
	Deaf      bool    `json:"deaf"`
	Mute      bool    `json:"mute"`
	SelfDeaf  bool    `json:"self_deaf"`
	SelfMute  bool    `json:"self_mute"`
	Suppress  bool    `json:"suppress"`
}

type VoiceRegion struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	SampleHostname string `json:"sample_hostname"`
	SamplePort     int    `json:"sample_port"`
	VIP            bool   `json:"vip"`
	Optimal        bool   `json:"optimal"`
	Deprecated     bool   `json:"deprecated"`
	Custom         bool   `json:"custom"`
}

func (c *Client) GetVoiceRegions() (regions []*VoiceRegion, err error) {
	endpoint := path.Join("voice", "regions")
	req := c.newRequest("GET", endpoint, nil)
	return regions, c.doUnmarshal(req, endpoint, &regions)
}
