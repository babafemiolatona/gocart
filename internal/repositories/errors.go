package repositories

import "gorm.io/gorm"

var (
	// ErrRecordNotFound and ErrDuplicate are repository-level error sentinels.
	// They alias the underlying ORM's errors so callers such as services never
	// need to import gorm for control flow.
	ErrRecordNotFound = gorm.ErrRecordNotFound
	ErrDuplicate      = gorm.ErrDuplicatedKey
)
