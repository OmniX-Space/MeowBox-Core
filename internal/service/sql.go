package service

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// ConnectDatabase connects to MySQL and returns a *sql.DB handle.
// The caller is responsible for calling db.Close() when done.
func ConnectDatabase(host string, port int, user string, pass string, dbName string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, pass, host, port, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("[Error] Failed to connect to MySQL: %v", err)
	}

	// Optional: verify the connection is actually working
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("[Error] Failed to ping MySQL: %v", err)
	}

	return db, nil
}

// CloseDatabase closes the database connection
func CloseDatabase(db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil {
			log.Printf("[Warning] Failed to close database connection: %v", err)
		}
	}
}

// DatabaseExists checks if a database exists
func DatabaseExists(db *sql.DB, dbName string) bool {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = ?", dbName).Scan(&count)
	if err != nil {
		log.Printf("[Warning] Failed to check database existence: %v", err)
		return false
	}
	return count > 0
}

// TableExists checks if a table exists in the current database
func TableExists(db *sql.DB, tableName string) bool {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", tableName).Scan(&count)
	if err != nil {
		log.Printf("[Warning] Failed to check table existence: %v", err)
		return false
	}
	return count > 0
}

// CreateDatabase creates a new database
func CreateDatabase(db *sql.DB, dbName string) error {
	if !isValidName(dbName) {
		return fmt.Errorf("invalid database name: %s", dbName)
	}
	_, err := db.Exec(fmt.Sprintf("CREATE DATABASE `%s`", dbName))
	if err != nil {
		log.Printf("[Warning] Failed to create database: %v", err)
		return err
	}
	return nil
}

// DropDatabase drops a database
func DropDatabase(db *sql.DB, dbName string) error {
	if !isValidName(dbName) {
		return fmt.Errorf("invalid database name: %s", dbName)
	}
	_, err := db.Exec(fmt.Sprintf("DROP DATABASE `%s`", dbName))
	if err != nil {
		log.Printf("[Warning] Failed to drop database: %v", err)
		return err
	}
	return nil
}

// CreateTable creates a new table in the current database
func CreateTable(db *sql.DB, tableName string, tableDef string) error {
	if !isValidName(tableName) {
		return fmt.Errorf("invalid table name: %s", tableName)
	}
	_, err := db.Exec(fmt.Sprintf("CREATE TABLE `%s` (%s)", tableName, tableDef))
	if err != nil {
		log.Printf("[Warning] Failed to create table: %v", err)
		return err
	}
	return nil
}

// DropTable drops a table from the current database
func DropTable(db *sql.DB, tableName string) error {
	if !isValidName(tableName) {
		return fmt.Errorf("invalid table name: %s", tableName)
	}
	_, err := db.Exec(fmt.Sprintf("DROP TABLE `%s`", tableName))
	if err != nil {
		log.Printf("[Warning] Failed to drop table: %v", err)
		return err
	}
	return nil
}

// InsertRow inserts a new row into a table
func InsertRow(db *sql.DB, tableName string, rowData map[string]interface{}) error {
	if !isValidName(tableName) {
		return fmt.Errorf("invalid table name: %s", tableName)
	}

	cols := make([]string, 0, len(rowData))
	params := make([]interface{}, 0, len(rowData))
	for col, val := range rowData {
		if !isValidName(col) {
			return fmt.Errorf("invalid column name: %s", col)
		}
		cols = append(cols, "`"+col+"`")
		params = append(params, val)
	}

	placeholders := strings.Repeat("?, ", len(params)-1) + "?"
	query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", tableName, strings.Join(cols, ", "), placeholders)

	_, err := db.Exec(query, params...)
	if err != nil {
		log.Printf("[Warning] Failed to insert row: %v", err)
		return err
	}
	return nil
}

// UpdateRow updates a row in a table
func UpdateRow(db *sql.DB, tableName string, rowData map[string]interface{}, where string, whereArgs ...interface{}) error {
	if !isValidName(tableName) {
		return fmt.Errorf("invalid table name: %s", tableName)
	}
	setCols, setVals, err := buildUpdateQuery(rowData)
	if err != nil {
		log.Printf("[Warning] Failed to build update query: %v", err)
		return err
	}
	query := fmt.Sprintf("UPDATE `%s` SET %s WHERE %s", tableName, setCols, where)
	_, err = db.Exec(query, append(setVals, whereArgs...)...)
	if err != nil {
		log.Printf("[Warning] Failed to update row: %v", err)
		return err
	}
	return nil
}

// buildUpdateQuery builds an update query with placeholders for the values
func buildUpdateQuery(rowData map[string]interface{}) (string, []interface{}, error) {
	setCols := ""
	setVals := make([]interface{}, 0)
	for col, val := range rowData {
		if !isValidName(col) {
			return "", nil, fmt.Errorf("invalid column name: %s", col)
		}
		if setCols != "" {
			setCols += ", "
		}
		setCols += "`" + col + "` = ?"
		setVals = append(setVals, val)
	}
	if setCols == "" {
		return "", nil, fmt.Errorf("no columns specified")
	}
	return setCols, setVals, nil
}

// DeleteRow deletes a row from a table
func DeleteRow(db *sql.DB, tableName string, where string, whereArgs ...interface{}) error {
	if !isValidName(tableName) {
		return fmt.Errorf("invalid table name: %s", tableName)
	}
	query := fmt.Sprintf("DELETE FROM `%s` WHERE %s", tableName, where)
	_, err := db.Exec(query, whereArgs...)
	if err != nil {
		log.Printf("[Warning] Failed to delete row: %v", err)
		return err
	}
	return nil
}

func isValidName(s string) bool {
	if s == "" {
		return false
	}
	return !strings.ContainsAny(s, "`\x00")
}

// QueryRows queries rows from a table with optional WHERE clause.
// Example: rows, err := QueryRows(db, "users", "*", "age > ? AND status = ?", 18, "active")
func QueryRows(db *sql.DB, tableName string, columns string, where string, whereArgs ...interface{}) (*sql.Rows, error) {
	if !isValidName(tableName) {
		return nil, fmt.Errorf("invalid table name: %s", tableName)
	}

	// Basic validation for columns (very simple; for production consider stricter rules or whitelist)
	if columns == "" {
		columns = "*"
	}

	var query string
	if where != "" {
		query = fmt.Sprintf("SELECT %s FROM `%s` WHERE %s", columns, tableName, where)
	} else {
		query = fmt.Sprintf("SELECT %s FROM `%s`", columns, tableName)
	}

	rows, err := db.Query(query, whereArgs...)
	if err != nil {
		log.Printf("[Error] Failed to query rows: %v", err)
		return nil, err
	}
	return rows, nil
}

// ScanRowsToMap converts sql.Rows to a slice of maps.
// Note: This is a basic implementation and assumes all columns are nullable or compatible with interface{}.
func ScanRowsToMap(rows *sql.Rows) ([]map[string]interface{}, error) {
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := []map[string]interface{}{}

	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the column name as the key.
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Handle nil values properly
			if val == nil {
				rowMap[col] = nil
			} else if b, ok := val.([]byte); ok {
				// Convert []byte to string for easier handling (e.g., TEXT, VARCHAR)
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}
		results = append(results, rowMap)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
