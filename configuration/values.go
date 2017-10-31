package configuration

type ServerConfiguration struct {
	Address           string                `json:"address"`
	ConnectionTimeout int                   `json:"connection_timeout"`
	LogFile           string                `json:"log_file"`
	Motd              MessageOfTheDayValues `json:"motd"`
	LoginAttempt      LoginAttemptValues    `json:"login_attempt"`
}

// clickEvent or hoverEvent is not needed
type ChatComponentValue struct {
	Text          string `json:"text"`
	Bold          string `json:"bold,omitempty"`
	Italic        string `json:"italic,omitempty"`
	Underlined    string `json:"underlined,omitempty"`
	Strikethrough string `json:"strikethrough,omitempty"`
	Obfuscated    string `json:"obfuscated, omitempty"`
	Color         string `json:"color, omitempty"`
	Insertion     string `json:"insertion, omitempty"`
}

// clickEvent or hoverEvent is not needed
type ChatValue struct {
	Text          string               `json:"text"`
	Bold          string               `json:"bold,omitempty"`
	Italic        string               `json:"italic,omitempty"`
	Underlined    string               `json:"underlined,omitempty"`
	Strikethrough string               `json:"strikethrough,omitempty"`
	Obfuscated    string               `json:"obfuscated, omitempty"`
	Color         string               `json:"color, omitempty"`
	Insertion     string               `json:"insertion, omitempty"`
	Extra         []ChatComponentValue `json:"extra, omitempty"`
}

// MOTD
type MessageOfTheDayValues struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    int `json:"max"`
		Online int `json:"online"`
		Sample []struct {
			Name string `json:"name"`
			Id   string `json:"id"`
		} `json:"sample,omitempty"`
	} `json:"players"`
	Description struct {
		Text string `json:"text"`
	} `json:"description"`
	FaviconPath string `json:"favicon-path,omitempty"`
}

// if a user tries to login
type LoginAttemptValues struct {
	DisconnectText ChatValue
}
