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

var StatusOpen = []Status{StatusNew, StatusInProgress}
var StatusClosed = []Status{StatusFixed, StatusWontFix}

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
		ID:      id,
		Title:   title,
		Type:    pebType,
		Status:  status,
		Created: timestamp,
		Changed: timestamp,
		Content: content,
	}
}

func (p *Peb) UpdateTimestamp() {
	p.Changed = time.Now().Local().Format(timestampFormat)
}

func IsClosed(status Status) bool {
	for _, s := range StatusClosed {
		if status == s {
			return true
		}
	}
	return false
}

func IsOpen(status Status) bool {
	for _, s := range StatusOpen {
		if status == s {
			return true
		}
	}
	return false
}
