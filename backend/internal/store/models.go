package store

import "time"

// FilterGateway holds the OneBot reverse-WS gateway settings.
// Singleton row identified by ID = 1.
type FilterGateway struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	Enabled      bool      `gorm:"default:true" json:"enabled"`
	Host         string    `gorm:"default:'0.0.0.0'" json:"host"`
	Port         int       `gorm:"default:3939" json:"port"`
	Suffix       string    `gorm:"default:'/ws'" json:"suffix"`
	BotID        string    `gorm:"default:'10000'" json:"bot_id"`
	AccessToken  string    `gorm:"default:''" json:"access_token"`
	UserAgent    string    `gorm:"default:'NEXTonebotfilter'" json:"user_agent"`
	BufferSize   int       `gorm:"default:4096" json:"buffer_size"`
	SleepTime    float32   `gorm:"default:5" json:"sleep_time"`
	Debug        bool      `json:"debug"`
	DedupEnabled bool      `gorm:"default:true" json:"dedup_enabled"`
	DedupTTL     int       `gorm:"default:60" json:"dedup_ttl"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// FilterTemplate is a reusable bundle of filter rules referenced by FilterApp.
// The seeded "default" template doubles as the global ID-rule fallback.
type FilterTemplate struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	Name        string `gorm:"not null;uniqueIndex" json:"name"`
	Description string `json:"description"`
	Builtin     bool   `gorm:"default:false" json:"builtin"`

	UserIDRules         string `gorm:"type:text;default:'{}'" json:"user_id_rules"`
	GroupIDRules        string `gorm:"type:text;default:'{}'" json:"group_id_rules"`
	MessageRules        string `gorm:"type:text;default:'{}'" json:"message_rules"`
	PrivateMessageRules string `gorm:"type:text;default:'{}'" json:"private_message_rules"`
	GroupMessageRules   string `gorm:"type:text;default:'{}'" json:"group_message_rules"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FilterApp is one downstream OneBot bot the gateway forwards messages to.
// When TemplateID is set, the per-app rule fields are ignored at compile time.
type FilterApp struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	Name        string `gorm:"not null;uniqueIndex" json:"name"`
	URI         string `gorm:"not null" json:"uri"`
	AccessToken string `json:"access_token"`
	Enabled     bool   `gorm:"default:true" json:"enabled"`
	Builtin     bool   `gorm:"default:false" json:"builtin"`
	Internal    bool   `gorm:"default:false" json:"internal"`
	SortOrder   int    `gorm:"default:0" json:"sort_order"`

	TemplateID *uint `gorm:"index" json:"template_id,omitempty"`

	UserIDRules         string `gorm:"type:text;default:'{}'" json:"user_id_rules"`
	GroupIDRules        string `gorm:"type:text;default:'{}'" json:"group_id_rules"`
	MessageRules        string `gorm:"type:text;default:'{}'" json:"message_rules"`
	PrivateMessageRules string `gorm:"type:text;default:'{}'" json:"private_message_rules"`
	GroupMessageRules   string `gorm:"type:text;default:'{}'" json:"group_message_rules"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
