package database

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type columnTestResult struct {
	Name       string
	isPrimary  bool
	isForeign  bool
	isUnique   bool
	isNullable bool
}

type connectionParameter struct {
	connectionString string
	schema           string
}

var (
	testConnectionPostgres connectionParameter = connectionParameter{connectionString: "postgresql://user:password@localhost:5432/mermerd_test", schema: "public"}
	testConnectionMySql    connectionParameter = connectionParameter{connectionString: "mysql://root:password@tcp(127.0.0.1:3306)/mermerd_test", schema: "mermerd_test"}
	testConnectionMsSql    connectionParameter = connectionParameter{connectionString: "sqlserver://sa:securePassword1!@localhost:1433?database=mermerd_test", schema: "dbo"}
	testConnectionAzure    connectionParameter = connectionParameter{connectionString: "sqlserver://sa:securePassword1!@localhost:1434?database=mermerd_test", schema: "dbo"}
	testConnectionSqlite   connectionParameter = connectionParameter{connectionString: "sqlite3://../mermerd_test.db", schema: "mermerd_test"}
)

func TestDatabaseIntegrations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	connectorFactory := NewConnectorFactory()
	testCases := []struct {
		dbType           DbType
		connectionString string
		schema           string
	}{
		{
			dbType:           Postgres,
			connectionString: testConnectionPostgres.connectionString,
			schema:           testConnectionPostgres.schema,
		},
		{
			dbType:           MySql,
			connectionString: testConnectionMySql.connectionString,
			schema:           testConnectionMySql.schema,
		},
		{
			dbType:           MsSql,
			connectionString: testConnectionMsSql.connectionString,
			schema:           testConnectionMsSql.schema,
		},
		{
			// Azure SQL is the same as MSSQL
			dbType:           MsSql,
			connectionString: testConnectionAzure.connectionString,
			schema:           testConnectionAzure.schema,
		},
		{
			dbType:           Sqlite3,
			connectionString: testConnectionSqlite.connectionString,
			schema:           testConnectionSqlite.schema,
		},
	}

	for _, testCase := range testCases {
		connector, _ := connectorFactory.NewConnector(testCase.connectionString)

		getConnectionAndConnect := func(t *testing.T) Connector {
			err := connector.Connect()
			if err != nil {
				logrus.Error(err)
				t.FailNow()
			}
			return connector
		}

		t.Run(testCase.dbType.String(), func(t *testing.T) {
			t.Run("Connect", func(t *testing.T) {
				// Arrange
				connector := connector

				// Act
				err := connector.Connect()

				// Assert
				assert.Nil(t, err)
			})

			t.Run("GetSchemas", func(t *testing.T) {
				// Arrange
				connector := getConnectionAndConnect(t)

				// Act
				schemas, err := connector.GetSchemas()

				// Assert
				assert.Nil(t, err)
				assert.NotEmpty(t, schemas)
			})

			t.Run("GetTables", func(t *testing.T) {
				// Arrange
				connector := getConnectionAndConnect(t)
				schema := testCase.schema

				// Act
				tables, err := connector.GetTables([]string{schema})

				// Assert
				expectedResult := []TableDetail{
					{Schema: schema, Name: "article"},
					{Schema: schema, Name: "article_detail"},
					{Schema: schema, Name: "article_comment"},
					{Schema: schema, Name: "label"},
					{Schema: schema, Name: "article_label"},
					{Schema: schema, Name: "test_1_a"},
					{Schema: schema, Name: "test_1_b"},
					{Schema: schema, Name: "test_2_enum"},
					{Schema: schema, Name: "test_3_a"},
				}
				if testCase.dbType == Postgres {
					expectedResult = append(expectedResult, []TableDetail{
						{Schema: schema, Name: "test_not_unique_constraint_name_a"},
						{Schema: schema, Name: "test_not_unique_constraint_name_b"},
						{Schema: schema, Name: "test_not_unique_constraint_name_c"},
						{Schema: schema, Name: "tenants"},
					}...)
				}

				assert.Nil(t, err)
				assert.ElementsMatch(t, expectedResult, tables)
			})

			t.Run("GetColumns", func(t *testing.T) {
				connector := getConnectionAndConnect(t)
				subTestCases := []struct {
					tableName       string
					expectedColumns []columnTestResult
				}{
					{tableName: "article", expectedColumns: []columnTestResult{
						{Name: "id", isPrimary: true, isForeign: false, isUnique: false, isNullable: false},
						{Name: "title", isPrimary: false, isForeign: false, isUnique: false, isNullable: false},
						{Name: "subtitle", isPrimary: false, isForeign: false, isUnique: false, isNullable: true},
					}},
					{tableName: "article_detail", expectedColumns: []columnTestResult{
						{Name: "id", isPrimary: true, isForeign: true, isUnique: false, isNullable: false},
						{Name: "created_at", isPrimary: false, isForeign: false, isUnique: false, isNullable: false},
					}},
					{tableName: "article_comment", expectedColumns: []columnTestResult{
						{Name: "id", isPrimary: true, isForeign: false, isUnique: false, isNullable: false},
						{Name: "article_id", isPrimary: false, isForeign: true, isUnique: false, isNullable: false},
						{Name: "comment", isPrimary: false, isForeign: false, isUnique: false, isNullable: false},
					}},
					{tableName: "label", expectedColumns: []columnTestResult{
						{Name: "id", isPrimary: true, isForeign: false, isUnique: false, isNullable: false},
						{Name: "label", isPrimary: false, isForeign: false, isUnique: true, isNullable: false},
					}},
					{tableName: "article_label", expectedColumns: []columnTestResult{
						{Name: "article_id", isPrimary: true, isForeign: true, isUnique: false, isNullable: false},
						{Name: "label_id", isPrimary: true, isForeign: true, isUnique: false, isNullable: false},
					}},
					{tableName: "test_1_a", expectedColumns: []columnTestResult{
						{Name: "id", isPrimary: true, isForeign: false, isUnique: false, isNullable: false},
						{Name: "xid", isPrimary: true, isForeign: false, isUnique: false, isNullable: false},
					}},
					{tableName: "test_1_b", expectedColumns: []columnTestResult{
						{Name: "aid", isPrimary: true, isForeign: true, isUnique: false, isNullable: false},
						{Name: "bid", isPrimary: true, isForeign: true, isUnique: false, isNullable: false},
					}},
				}

				for index, subTestCase := range subTestCases {
					t.Run(fmt.Sprintf("run #%d", index), func(t *testing.T) {
						// Arrange
						tableName := TableDetail{Schema: testCase.schema, Name: subTestCase.tableName}
						var columnResult []columnTestResult

						// Act
						columns, err := connector.GetColumns(tableName)

						// Assert
						for _, column := range columns {
							columnResult = append(columnResult, columnTestResult{
								Name:       column.Name,
								isPrimary:  column.IsPrimary,
								isForeign:  column.IsForeign,
								isUnique:   column.IsUnique,
								isNullable: column.IsNullable,
							})
						}

						assert.Nil(t, err)
						assert.ElementsMatch(t, subTestCase.expectedColumns, columnResult)
					})
				}
			})

			t.Run("GetConstraints", func(t *testing.T) {
				connector := getConnectionAndConnect(t)

				t.Run("One-to-one relation", func(t *testing.T) {
					// Arrange
					tableName := TableDetail{Schema: testCase.schema, Name: "article_detail"}

					// Act
					constraintResults, err := connector.GetConstraints(tableName)

					// Assert
					assert.Nil(t, err)
					assert.Len(t, constraintResults, 1)
					if len(constraintResults) >= 1 {
						constraint := constraintResults[0]
						assert.True(t, constraint.IsPrimary)
						assert.False(t, constraint.HasMultiplePK)
					}
				})

				t.Run("Many-to-one relation #1", func(t *testing.T) {
					// Arrange
					tableName := TableDetail{Schema: testCase.schema, Name: "article_comment"}

					// Act
					constraintResults, err := connector.GetConstraints(tableName)

					// Assert
					assert.Nil(t, err)
					assert.Len(t, constraintResults, 1)
					if len(constraintResults) >= 1 {
						constraint := constraintResults[0]
						assert.False(t, constraint.IsPrimary)
						assert.False(t, constraint.HasMultiplePK)
					}
				})

				t.Run("Many-to-one relation #2", func(t *testing.T) {
					// Arrange
					pkTableName := TableDetail{Schema: testCase.schema, Name: "article"}
					fkTableName := TableDetail{Schema: testCase.schema, Name: "article_label"}

					// Act
					constraintResults, err := connector.GetConstraints(pkTableName)

					// Assert
					assert.Nil(t, err)
					var constraint *ConstraintResult
					for _, item := range constraintResults {
						if item.FkTable == fkTableName.Name {
							constraint = &item
							break
						}
					}
					assert.NotNil(t, constraint)
					if constraint != nil {
						assert.True(t, constraint.IsPrimary)
						assert.True(t, constraint.HasMultiplePK)
					}
				})

				// Multiple primary keys (https://github.com/KarnerTh/mermerd/issues/8)
				t.Run("Multiple primary keys (Issue #8)", func(t *testing.T) {
					// Arrange
					pkTableName := TableDetail{Schema: testCase.schema, Name: "test_1_b"}

					// Act
					constraintResults, err := connector.GetConstraints(pkTableName)

					// Assert
					assert.Nil(t, err)
					assert.NotNil(t, constraintResults)
					assert.Len(t, constraintResults, 2)
					if len(constraintResults) >= 2 {
						assert.True(t, constraintResults[0].IsPrimary)
						assert.True(t, constraintResults[0].HasMultiplePK)
						assert.Equal(t, constraintResults[0].ColumnName, "aid")
						assert.True(t, constraintResults[1].IsPrimary)
						assert.True(t, constraintResults[1].HasMultiplePK)
						assert.Equal(t, constraintResults[1].ColumnName, "bid")
					}
				})
			})

			t.Run("Multiple schemas (Issue #23)", func(t *testing.T) {
				if testCase.dbType == Sqlite3 {
					t.Skip("Sqlite does not support multiple schemas")
				}

				connector := getConnectionAndConnect(t)

				t.Run("GetTables", func(t *testing.T) {
					// Arrange
					secondSchema := "other_db"
					schemas := []string{testCase.schema, secondSchema}

					// Act
					tables, err := connector.GetTables(schemas)

					// Assert
					expectedResult := []TableDetail{
						{Schema: testCase.schema, Name: "article"},
						{Schema: testCase.schema, Name: "article_detail"},
						{Schema: testCase.schema, Name: "article_comment"},
						{Schema: testCase.schema, Name: "label"},
						{Schema: testCase.schema, Name: "article_label"},
						{Schema: testCase.schema, Name: "test_1_a"},
						{Schema: testCase.schema, Name: "test_1_b"},
						{Schema: testCase.schema, Name: "test_2_enum"},
						{Schema: testCase.schema, Name: "test_3_a"},
						{Schema: secondSchema, Name: "test_3_b"},
						{Schema: secondSchema, Name: "test_3_c"},
					}

					if testCase.dbType == Postgres {
						expectedResult = append(expectedResult, []TableDetail{
							{Schema: testCase.schema, Name: "test_not_unique_constraint_name_a"},
							{Schema: testCase.schema, Name: "test_not_unique_constraint_name_b"},
							{Schema: testCase.schema, Name: "test_not_unique_constraint_name_c"},
							{Schema: testCase.schema, Name: "tenants"},
						}...)
					}

					assert.Nil(t, err)
					assert.ElementsMatch(t, expectedResult, tables)
				})

				t.Run("Get Cross-Schema-Constraints", func(t *testing.T) {
					// Arrange
					tableName := TableDetail{Schema: "other_db", Name: "test_3_b"}

					// Act
					constraintResults, err := connector.GetConstraints(tableName)

					// Assert
					assert.Nil(t, err)
					assert.Len(t, constraintResults, 1)

					if len(constraintResults) >= 1 {
						assert.False(t, constraintResults[0].IsPrimary)
						assert.False(t, constraintResults[0].HasMultiplePK)
						assert.Equal(t, constraintResults[0].ColumnName, "aid")
						assert.Equal(t, constraintResults[0].FkTable, "test_3_b")
						assert.Equal(t, constraintResults[0].PkTable, "test_3_a")
					}
				})

				t.Run("Get schema from FK and PK table", func(t *testing.T) {
					// Arrange
					tableName := TableDetail{Schema: "other_db", Name: "test_3_b"}

					// Act
					constraintResults, err := connector.GetConstraints(tableName)

					// Assert
					assert.Nil(t, err)
					assert.Len(t, constraintResults, 1)
					if len(constraintResults) >= 1 {
						assert.Equal(t, constraintResults[0].FkTable, "test_3_b")
						assert.Equal(t, constraintResults[0].FkSchema, "other_db")
						assert.Equal(t, constraintResults[0].PkTable, "test_3_a")
						assert.Equal(t, constraintResults[0].PkSchema, testCase.schema)
					}
				})
			})

			t.Run("Schema with hyphen (special characters)", func(t *testing.T) {
				if testCase.dbType != Postgres {
					t.Skip("Test specific to Postgres")
				}

				connector := getConnectionAndConnect(t)

				t.Run("GetTables with hyphenated schema", func(t *testing.T) {
					// Arrange
					hyphenSchema := "tenant-alfa"
					schemas := []string{testCase.schema, hyphenSchema}

					// Act
					tables, err := connector.GetTables(schemas)

					// Assert
					assert.Nil(t, err)
					// Check that tables from hyphenated schema are included
					expectedHyphenTables := []TableDetail{
						{Schema: hyphenSchema, Name: "users"},
						{Schema: hyphenSchema, Name: "refresh_tokens"},
						{Schema: hyphenSchema, Name: "tenant_configs"},
					}
					for _, expectedTable := range expectedHyphenTables {
						assert.Contains(t, tables, expectedTable, "Expected table %s.%s to be present", expectedTable.Schema, expectedTable.Name)
					}
				})

				t.Run("GetColumns with hyphenated schema", func(t *testing.T) {
					// Arrange
					tableName := TableDetail{Schema: "tenant-alfa", Name: "users"}

					// Act
					columns, err := connector.GetColumns(tableName)

					// Assert
					assert.Nil(t, err)
					assert.Len(t, columns, 2)
					if len(columns) >= 2 {
						assert.Equal(t, "id", columns[0].Name)
						assert.True(t, columns[0].IsPrimary)
						assert.Equal(t, "name", columns[1].Name)
						assert.False(t, columns[1].IsPrimary)
					}
				})

				t.Run("GetConstraints same-schema FK", func(t *testing.T) {
					// Arrange
					tableName := TableDetail{Schema: "tenant-alfa", Name: "refresh_tokens"}

					// Act
					constraintResults, err := connector.GetConstraints(tableName)

					// Assert
					assert.Nil(t, err)
					assert.NotEmpty(t, constraintResults)
					
					// Find the FK constraint (not the PK)
					var fkConstraint *ConstraintResult
					for i := range constraintResults {
						if !constraintResults[i].IsPrimary {
							fkConstraint = &constraintResults[i]
							break
						}
					}
					
					assert.NotNil(t, fkConstraint, "Should have found a foreign key constraint")
					if fkConstraint != nil {
						assert.Equal(t, "refresh_tokens", fkConstraint.FkTable)
						assert.Equal(t, "tenant-alfa", fkConstraint.FkSchema)
						assert.Equal(t, "users", fkConstraint.PkTable)
						assert.Equal(t, "tenant-alfa", fkConstraint.PkSchema)
						assert.Equal(t, "user_id", fkConstraint.ColumnName)
					}
				})

				t.Run("GetConstraints cross-schema FK", func(t *testing.T) {
					// Arrange
					tableName := TableDetail{Schema: "tenant-alfa", Name: "tenant_configs"}

					// Act
					constraintResults, err := connector.GetConstraints(tableName)

					// Assert
					assert.Nil(t, err)
					assert.NotEmpty(t, constraintResults)
					
					// Find the FK constraint (not the PK)
					var fkConstraint *ConstraintResult
					for i := range constraintResults {
						if !constraintResults[i].IsPrimary {
							fkConstraint = &constraintResults[i]
							break
						}
					}
					
					assert.NotNil(t, fkConstraint, "Should have found a foreign key constraint")
					if fkConstraint != nil {
						assert.Equal(t, "tenant_configs", fkConstraint.FkTable)
						assert.Equal(t, "tenant-alfa", fkConstraint.FkSchema)
						assert.Equal(t, "tenants", fkConstraint.PkTable)
						assert.Equal(t, testCase.schema, fkConstraint.PkSchema)
						assert.Equal(t, "tenant_id", fkConstraint.ColumnName)
					}
				})
			})
		})
	}
}
