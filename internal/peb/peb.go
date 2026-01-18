package peb

import "time"

const timestampFormat = "2006-01-02T15:04:05-07:00"

type Type string

const (
	TypeBug     Type = "bug"
	TypeFeature Type = "feature"
	TypeEpic    Type = "epic"
	TypeTask    Type = "task"
)

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in-progress"
	StatusFixed      Status = "fixed"
	StatusWontFix    Status = "wont-fix"
)

type Peb struct {
	ID        string   `yaml:"id" json:"id"`
	Title     string   `yaml:"title" json:"title"`
	Type      Type     `yaml:"type" json:"type"`
	Status    Status   `yaml:"status" json:"status"`
	Created   string   `yaml:"created" json:"created"`
	Changed   string   `yaml:"changed" json:"changed"`
	BlockedBy []string `yaml:"blocked-by,omitempty" json:"blocked-by,omitempty"`
	Content   string   `yaml:"-" json:"content"`
}

func New(id, title string, pebType Type, status Status, content string) *Peb {
	now := time.Now()
	timestamp := now.Local().Format(timestampFormat)
	return &Peb{
		ID:        id,
		Title:     title,
		Type:      pebType,
		Status:    status,
		Created:   timestamp,
		Changed:   timestamp,
		BlockedBy: []string{},
		Content:   content,
	}
}

func (p *Peb) UpdateTimestamp() {
	p.Changed = time.Now().Local().Format(timestampFormat)
}
