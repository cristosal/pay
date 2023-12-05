package pay

import "github.com/cristosal/orm"

var migrations = []orm.Migration{
	{
		Name:        "customer table",
		Description: "create customers table",
		Up: `
		CREATE TABLE {{ .Schema }}.customer (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT Null,
			email VARCHAR(255) NOT Null,
			provider VARCHAR(32) NOT NUll,
			provider_id VARCHAR(255) NOT Null,
			UNIQUE (provider, provider_id)
		);`,
		Down: "DROP TABLE {{ .Schema }}.customer",
	},
	{
		Name:        "plan table",
		Description: "create plan table",
		Up: `
		CREATE TABLE {{ .Schema }}.plan (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			provider VARCHAR(255) NOT NULL,
			provider_id VARCHAR(255) NOT NULL,
			active BOOL NOT NULL,
			UNIQUE (provider, provider_id)
		);`,
		Down: "DROP TABLE {{ .Schema }}.plan",
	},
	{
		Name:        "subscription table",
		Description: "create subscription table",
		Up: `
		CREATE TABLE {{ .Schema }}.subscription (
			id SERIAL PRIMARY KEY,
			customer_id INT NOT NULL,
			price_id INT NOT NULL,
			provider VARCHAR(255) NOT NULL,
			provider_id VARCHAR(255) NOT NULL,
			active BOOL NOT NULL DEFAULT FALSE,
			FOREIGN KEY (customer_id) REFERENCES {{ .Schema }}.customer(id),
			FOREIGN KEY (price_id) REFERENCES {{ .Schema }}.price(id),
			UNIQUE (provider, provider_id)
		);`,
		Down: "DROP TABLE {{ .Schema }}.subscription",
	},
	{
		Name:        "price table",
		Description: "create price table",
		Up: `
		CREATE TABLE {{ .Schema }}.price (
			id SERIAL PRIMARY KEY,
			plan_id INT NOT NULL,
			provider VARCHAR(255) NOT NULL,
			provider_id VARCHAR(255) NOT NULL,
			currency VARCHAR(3) NOT NULL,
			amount INT NOT NULL DEFAULT 0,
			schedule VARCHAR(32) NOT NULL,
			FOREIGN KEY (plan_id) REFERENCES {{ .Schema }}.plan (id),
			UNIQUE (provider, provider_id)
		);`,
		Down: "DROP TABLE {{ .Schema }}.subscription",
	},
}