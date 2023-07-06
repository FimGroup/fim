package fimcore

import "testing"

func TestParseSubConnectorToml(t *testing.T) {
	app := newApplication()

	if err := app.AddSubConnectorGeneratorDefinitions(`

[target_connector.db_forum_name]
"@parent" = "&database_postgres"
"database.connect_string" = "configure-static://forum_database"

`); err != nil {
		t.Fatal(err)
	}

	if app.targetConnectorGeneratorDefinitions == nil || app.sourceConnectorGeneratorDefinitions == nil {
		t.Fatal("nil value of definition")
	}
}
