package third

import "gorm.io/gorm"

//This is a nil ptr bug in gorm.io/gorm@v1.23.2/migrator/migrator.go:369
func MysqlMigratorHasColumn(db *gorm.DB, schemaName string, tableName string, columnName string) bool {

	var count int64
	err := db.Raw(
		"SELECT count(*) FROM INFORMATION_SCHEMA.columns WHERE table_schema = ? AND table_name = ? AND column_name = ?",
		schemaName, tableName, columnName,
	).Row().Scan(&count)
	if err != nil {
		panic(err)
	}

	return count > 0
}
