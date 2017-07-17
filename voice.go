package discgo

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

type VoiceRegionsEndpoint struct {
	*endpoint
}

func (c *Client) VoiceRegions() VoiceRegionsEndpoint {
	e2 := c.e.appendMajor("voice").appendMajor("regions")
	return VoiceRegionsEndpoint{e2}
}

func (e VoiceRegionsEndpoint) Get() (voiceRegions []*VoiceRegion, err error) {
	return voiceRegions, e.doMethod("GET", nil, &voiceRegions)
}
