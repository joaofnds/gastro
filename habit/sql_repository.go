package habit

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func resultErr(result *gorm.DB) error {
	err := result.Error
	if err == nil && result.RowsAffected == 0 {
		return ErrNotFound
	}
	return translateError(err)
}

func translateError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	default:
		fmt.Printf("\n%v\n", err)
		return ErrRepository
	}
}

func normalizeUUIDs(ids ...*string) bool {
	for _, id := range ids {
		parsed, err := uuid.Parse(*id)
		if err != nil {
			return false
		}

		*id = parsed.String()
	}

	return true
}
