package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ParticipantType string
type MatchMode string
type ParticipantStatus string

const (
	TypeTeam   ParticipantType = "team"
	TypePerson ParticipantType = "person"

	ModeExact MatchMode = "exact"
	ModeFuzzy MatchMode = "fuzzy"

	StatusPending ParticipantStatus = "pending"
	StatusMatched ParticipantStatus = "matched"
)

type Participant struct {
	UUID          string            `gorm:"column:uuid;primaryKey;size:19" json:"uuid"`
	Type          ParticipantType   `gorm:"column:type;type:enum('team','person');not null" json:"type"`
	Score         int               `gorm:"column:score;not null" json:"score"`
	MatchMode     MatchMode         `gorm:"column:match_mode;type:enum('exact','fuzzy');default:exact" json:"match_mode"`
	QRCodeContent string            `gorm:"column:qrcode_content;size:200;uniqueIndex" json:"qrcode_content,omitempty"`
	QRCodePath    string            `gorm:"column:qrcode_path;size:255" json:"qrcode_path,omitempty"`
	ClientIP      string            `gorm:"column:client_ip;size:45;index" json:"-"`
	Status        ParticipantStatus `gorm:"column:status;type:enum('pending','matched');default:pending" json:"status"`
	MatchedUUID   *string           `gorm:"column:matched_uuid;size:19" json:"matched_uuid,omitempty"`
	CreatedAt     time.Time         `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
	UpdatedAt     time.Time         `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Participant) TableName() string {
	return "participants"
}

func GenerateUUID() string {
	u := uuid.New()
	hex := strings.ReplaceAll(u.String(), "-", "")
	return fmt.Sprintf("%s-%s-%s-%s", hex[0:4], hex[4:8], hex[8:12], hex[12:16])
}

type MatchInfo struct {
	Score      int    `json:"score"`
	QRCodePath string `json:"qrcode_path"`
}
