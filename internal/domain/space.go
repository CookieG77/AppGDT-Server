// Contains the type of space

package domain

import "time"

type Space struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"-"` // Hidden from serialization since useless when requested by the same user (and no other user can request it)
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
