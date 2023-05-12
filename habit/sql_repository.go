package habit

import (
	"errors"
	"fmt"

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
