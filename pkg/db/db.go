package db

import (
	"database/sql"
	"database/sql/driver"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/ViBiOh/httputils/pkg/tools"

	// Not referenced but needed for database/sql
	_ "github.com/lib/pq"
)

// WhereInUint wrapper for assigning `IN ($n)` in WHERE clause
type WhereInUint []uint

// Value implements the driver.Valuer interface.
func (a WhereInUint) Value() (driver.Value, error) {
	ints := make([]string, len(a))
	for i, v := range a {
		ints[i] = strconv.FormatUint(uint64(v), 10)
	}

	return strings.Join(ints, ","), nil
}

// PrepareFullTextSearch replace $INDEX param in query and expand words
func PrepareFullTextSearch(query, search string, index uint) (string, string) {
	if search == `` {
		return ``, ``
	}

	words := strings.Split(search, ` `)
	transformedWords := make([]string, 0, len(words))

	for _, word := range words {
		transformedWords = append(transformedWords, word+`:*`)
	}

	return strings.Replace(query, `$INDEX`, fmt.Sprintf(`$%d`, index), -1), strings.Join(transformedWords, ` | `)
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`host`: flag.String(tools.ToCamel(fmt.Sprintf(`%sHost`, prefix)), ``, `[database] Host`),
		`port`: flag.String(tools.ToCamel(fmt.Sprintf(`%sPort`, prefix)), `5432`, `[database] Port`),
		`user`: flag.String(tools.ToCamel(fmt.Sprintf(`%sUser`, prefix)), ``, `[database] User`),
		`pass`: flag.String(tools.ToCamel(fmt.Sprintf(`%sPass`, prefix)), ``, `[database] Pass`),
		`name`: flag.String(tools.ToCamel(fmt.Sprintf(`%sName`, prefix)), ``, `[database] Name`),
	}
}

// GetDB start DB connection
func GetDB(config map[string]*string) (*sql.DB, error) {
	host := strings.TrimSpace(*config[`host`])
	if host == `` {
		return nil, nil
	}

	port := strings.TrimSpace(*config[`port`])
	user := strings.TrimSpace(*config[`user`])
	pass := strings.TrimSpace(*config[`pass`])
	name := strings.TrimSpace(*config[`name`])

	db, err := sql.Open(`postgres`, fmt.Sprintf(`host=%s port=%s user=%s password=%s dbname=%s sslmode=disable`, host, port, user, pass, name))
	if err != nil {
		return nil, fmt.Errorf(`Error while opening database connection: %v`, err)
	}

	if err = db.Ping(); err != nil {
		return db, fmt.Errorf(`Error while connecting to database: %v`, err)
	}

	return db, nil
}

// Ping indicate if database is ready or not
func Ping(db *sql.DB) bool {
	return db != nil && db.Ping() == nil
}

// GetTx return given transaction if not nil or create a new one
func GetTx(db *sql.DB, tx *sql.Tx) (*sql.Tx, error) {
	if db != nil && tx == nil {
		usedTx, err := db.Begin()

		if err != nil {
			return nil, fmt.Errorf(`Error while starting transaction: %v`, err)
		}
		return usedTx, nil
	}

	return tx, nil
}

// EndTx ends transaction according error without shadowing given error
func EndTx(tx *sql.Tx, err error) error {
	if err != nil {
		if endErr := tx.Rollback(); endErr != nil {
			return fmt.Errorf(`%v, and also, Error while rolling back transaction: %v`, err, endErr)
		}
	} else if endErr := tx.Commit(); endErr != nil {
		return fmt.Errorf(`Error while committing transaction: %v`, endErr)
	}

	return err
}

// RowsClose closes rows without shadowing error
func RowsClose(rows *sql.Rows, err error) error {
	if endErr := rows.Close(); endErr != nil {
		endErr = fmt.Errorf(`Error while closing rows: %v`, endErr)

		if err == nil {
			return endErr
		}
		return fmt.Errorf(`%v, and also, %v`, err, endErr)
	}

	return err
}
