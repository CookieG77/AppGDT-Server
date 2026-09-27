package domain

import "time"

type NoteStatus string

const (
	Todo       NoteStatus = "todo"
	InProgress NoteStatus = "in_progress"
	Done       NoteStatus = "done"
)

func (n NoteStatus) IsValid() bool {
	switch n {
	case Todo, InProgress, Done:
		return true
	}
	return false
}

const (
	NoteTitleMaxLength   = 255
	NoteContentMaxLength = 50000
)

type Note struct {
	ID        int64      `json:"id"`
	SpaceID   int64      `json:"spaceId"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Status    NoteStatus `json:"status"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}
