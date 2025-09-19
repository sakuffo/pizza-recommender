package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// connectDB establishes a connection to the SQLite database file.
func connectDB(path string) (*sql.DB, error) {
	// Enable foreign key constraints for SQLite
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database at %s: %w", path, err)
	}
	// Ping verifies the connection is alive
	if err := db.Ping(); err != nil {
		db.Close() // Close the connection if ping fails
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	log.Println("Database connection established.")
	return db, nil
}

// migrate ensures the necessary tables exist in the database.
func migrate(db *sql.DB) error {
	// Schema definition (should match db.sql)
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS pizzas (
           id   INTEGER PRIMARY KEY AUTOINCREMENT,
           name TEXT    NOT NULL UNIQUE
         );`,
		`CREATE TABLE IF NOT EXISTS ingredients (
           id   INTEGER PRIMARY KEY AUTOINCREMENT,
           name TEXT    NOT NULL UNIQUE
         );`,
		`CREATE TABLE IF NOT EXISTS pizza_ingredients (
           pizza_id      INTEGER NOT NULL,
           ingredient_id INTEGER NOT NULL,
           FOREIGN KEY (pizza_id)      REFERENCES pizzas(id) ON DELETE CASCADE,
           FOREIGN KEY (ingredient_id) REFERENCES ingredients(id) ON DELETE CASCADE,
           UNIQUE(pizza_id, ingredient_id)
         );`,
	}

	log.Println("Running database migrations...")
	tx, err := db.Begin() // Use a transaction for migrations
	if err != nil {
		return fmt.Errorf("failed to begin transaction for migration: %w", err)
	}
	// Use defer with a named return variable or a closure to handle rollback correctly
	var txErr error
	defer func() {
		if txErr != nil {
			log.Printf("Rolling back migration transaction due to error: %v", txErr)
			tx.Rollback()
		}
	}()

	for _, s := range stmts {
		if _, txErr = tx.Exec(s); txErr != nil {
			// Wrap the error with context
			txErr = fmt.Errorf("migration failed for statement [%s]: %w", s, txErr)
			return txErr // Trigger rollback via defer
		}
	}

	txErr = tx.Commit() // Try to commit
	if txErr != nil {
		txErr = fmt.Errorf("failed to commit migration transaction: %w", txErr)
		return txErr // Trigger rollback via defer
	}

	log.Println("Database migrations completed successfully.")
	return nil // Success
}

// seed populates the database with initial data from the provided map.
// It uses INSERT OR IGNORE to avoid errors if data already exists.
func seed(db *sql.DB, pizzasMap map[string][]string) error {
	log.Println("Seeding initial data...")
	tx, err := db.Begin() // Use a transaction for seeding
	if err != nil {
		return fmt.Errorf("failed to begin transaction for seeding: %w", err)
	}

	// Use defer with a named return variable or closure for rollback
	var txErr error
	defer func() {
		if txErr != nil {
			log.Printf("Rolling back seed transaction due to error: %v", txErr)
			tx.Rollback()
		}
	}()

	// Prepare statements for efficiency within the transaction
	pizzaInsertStmt, txErr := tx.Prepare(`INSERT OR IGNORE INTO pizzas(name) VALUES(?)`)
	if txErr != nil {
		return fmt.Errorf("failed to prepare pizza insert: %w", txErr)
	}
	defer pizzaInsertStmt.Close()

	ingredientInsertStmt, txErr := tx.Prepare(`INSERT OR IGNORE INTO ingredients(name) VALUES(?)`)
	if txErr != nil {
		return fmt.Errorf("failed to prepare ingredient insert: %w", txErr)
	}
	defer ingredientInsertStmt.Close()

	pizzaIngredientInsertStmt, txErr := tx.Prepare(`INSERT OR IGNORE INTO pizza_ingredients(pizza_id, ingredient_id) VALUES(?,?)`)
	if txErr != nil {
		return fmt.Errorf("failed to prepare pizza_ingredient insert: %w", txErr)
	}
	defer pizzaIngredientInsertStmt.Close()

	pizzaSelectStmt, txErr := tx.Prepare(`SELECT id FROM pizzas WHERE name = ?`)
	if txErr != nil {
		return fmt.Errorf("failed to prepare pizza select: %w", txErr)
	}
	defer pizzaSelectStmt.Close()

	ingredientSelectStmt, txErr := tx.Prepare(`SELECT id FROM ingredients WHERE name = ?`)
	if txErr != nil {
		return fmt.Errorf("failed to prepare ingredient select: %w", txErr)
	}
	defer ingredientSelectStmt.Close()

	insertedCount := 0
	for name, ingredients := range pizzasMap {
		// 1. Insert Pizza or get ID if exists
		var pizzaID int64
		res, err := pizzaInsertStmt.Exec(name)
		if err != nil {
			txErr = fmt.Errorf("failed to insert pizza %s: %w", name, err)
			return txErr
		}
		if id, err := res.LastInsertId(); err == nil && id != 0 {
			pizzaID = id // Get ID if newly inserted
		} else { // Pizza already existed, find its ID
			// Check for QueryRow error separately
			err = pizzaSelectStmt.QueryRow(name).Scan(&pizzaID)
			if err != nil {
				txErr = fmt.Errorf("failed to find existing pizza %s: %w", name, err)
				return txErr
			}
		}

		// 2. Insert Ingredients or get ID if exists, and link to Pizza
		for _, ing := range ingredients {
			var ingredientID int64
			res, err := ingredientInsertStmt.Exec(ing)
			if err != nil {
				txErr = fmt.Errorf("failed to insert ingredient %s for pizza %s: %w", ing, name, err)
				return txErr
			}
			if id, err := res.LastInsertId(); err == nil && id != 0 {
				ingredientID = id // Get ID if newly inserted
			} else { // Ingredient already existed, find its ID
				err = ingredientSelectStmt.QueryRow(ing).Scan(&ingredientID)
				if err != nil {
					txErr = fmt.Errorf("failed to find existing ingredient %s: %w", ing, err)
					return txErr
				}
			}

			// 3. Link Pizza and Ingredient
			_, err = pizzaIngredientInsertStmt.Exec(pizzaID, ingredientID)
			if err != nil {
				// Ignore unique constraint errors specifically
				if !isUniqueConstraintError(err) {
					txErr = fmt.Errorf("failed to link pizza %d (%s) to ingredient %d (%s): %w", pizzaID, name, ingredientID, ing, err)
					return txErr
				}
				// Log the ignored unique constraint error? Optional.
				// log.Printf("Ignoring unique constraint violation for pizza %d / ingredient %d", pizzaID, ingredientID)
			} else {
				insertedCount++
			}
		}
	}

	txErr = tx.Commit() // Try to commit
	if txErr != nil {
		txErr = fmt.Errorf("failed to commit seed transaction: %w", txErr)
		return txErr // Trigger rollback via defer
	}

	log.Printf("Seeding finished. Inserted %d new pizza-ingredient links.", insertedCount)
	return nil // Success
}

// Helper to check for SQLite unique constraint error
// NOTE: The exact error string might vary slightly. Check your logs if this doesn't work.
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	// Check for common SQLite unique constraint error messages
	errMsg := err.Error()
	return strings.Contains(errMsg, "UNIQUE constraint failed") || strings.Contains(errMsg, "constraint failed")
}
