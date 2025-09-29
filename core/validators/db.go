package validators

import (
	"database/sql"
	"errors"
	"regexp"
	"slices"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pocketbase/dbx"
)

// UniqueId checks whether a field string id already exists in the specified table.
//
// Example:
//
//	validation.Field(&form.RelId, validation.By(validators.UniqueId(form.app.DB(), "tbl_example"))
func UniqueId(db dbx.Builder, tableName string) validation.RuleFunc {
	return func(value any) error {
		v, _ := value.(string)
		if v == "" {
			return nil // nothing to check
		}

		var foundId string

		err := db.
			Select("id").
			From(tableName).
			Where(dbx.HashExp{"id": v}).
			Limit(1).
			Row(&foundId)

		if (err != nil && !errors.Is(err, sql.ErrNoRows)) || foundId != "" {
			return validation.NewError("validation_invalid_or_existing_id", "The model id is invalid or already exists.")
		}

		return nil
	}
}

// NormalizeUniqueIndexError attempts to convert a
// "unique constraint failed" error into a validation.Errors.
//
// The provided err is returned as it is without changes if:
// - err is nil
// - err is already validation.Errors
// - err is not "unique constraint failed" error
func NormalizeUniqueIndexError(err error, tableOrAlias string, fieldNames []string) error {
	if err == nil {
		return err
	}

	if _, ok := err.(validation.Errors); ok {
		return err
	}

	// Try to cast to a pgx Error
	pgxErr, isPgxError := err.(*pgconn.PgError)

	// If it's a pgx error and it's a unique violation
	if isPgxError && pgxErr.Code == "23505" { // code for unique violation
		normalizedErrs := validation.Errors{}

		if pgxErr.TableName != tableOrAlias {
			return err
		}

		uniqueKeys := strings.Split(strings.TrimSpace(regexp.MustCompile(`Key \((.*?)\)=`).FindStringSubmatch(pgxErr.Detail)[1]), ",")

		for _, key := range uniqueKeys {
			if slices.Contains(fieldNames, key) {
				normalizedErrs[key] = validation.NewError("validation_not_unique", "Value must be unique")
			}
		}

		if len(normalizedErrs) > 0 {
			return normalizedErrs
		}
	}

	return err
}
