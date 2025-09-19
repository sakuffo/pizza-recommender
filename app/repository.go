package main

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings" // For building IN clauses safely
)

// Pizza represents a pizza with a name and a list of ingredients.
// Used internally by repository and returned by fetchAllPizzas.
type Pizza struct {
	// ID int // Keep ID internal to repository for now
	Name        string
	Ingredients []string
}

// PizzaRecommendation is a pizza with a score based on user preferences.
// Used internally by repository and returned by getRecommendations.
type PizzaRecommendation struct {
	Name        string
	Ingredients []string
	Score       float64
}

// fetchAllPizzas retrieves all pizzas with their ingredients from the database.
// Uses GROUP_CONCAT for efficiency.
func fetchAllPizzas(db *sql.DB) ([]Pizza, error) {
	query := `
		SELECT
			p.name,
			COALESCE(GROUP_CONCAT(i.name, '|'), '') -- Use '|' as delimiter, handle no ingredients
		FROM pizzas p
		LEFT JOIN pizza_ingredients pi ON p.id = pi.pizza_id
		LEFT JOIN ingredients i ON pi.ingredient_id = i.id
		GROUP BY p.id, p.name -- Group by id and name
		ORDER BY p.name;
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pizzas with ingredients: %w", err)
	}
	defer rows.Close()

	var pizzas []Pizza
	for rows.Next() {
		var p Pizza
		var ingredientsConcat sql.NullString // Use NullString for COALESCE

		if err := rows.Scan(&p.Name, &ingredientsConcat); err != nil {
			return nil, fmt.Errorf("failed to scan pizza row: %w", err)
		}

		// Split the concatenated ingredients string
		if ingredientsConcat.Valid && ingredientsConcat.String != "" {
			p.Ingredients = strings.Split(ingredientsConcat.String, "|")
			sort.Strings(p.Ingredients) // Keep ingredients sorted
		} else {
			p.Ingredients = []string{} // Ensure it's an empty slice, not nil
		}
		pizzas = append(pizzas, p)
	}
	// Check for errors during row iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pizza rows: %w", err)
	}

	return pizzas, nil
}

// getAllIngredients retrieves a sorted list of all unique ingredient names.
func getAllIngredients(db *sql.DB) ([]string, error) {
	query := `SELECT name FROM ingredients ORDER BY name;`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query ingredients: %w", err)
	}
	defer rows.Close()

	var ingredients []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan ingredient row: %w", err)
		}
		ingredients = append(ingredients, name)
	}
	// Check for errors during row iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ingredient rows: %w", err)
	}
	return ingredients, nil
}

// getRecommendations filters pizzas based on disliked ingredients and scores based on preferred ones.
// Performs filtering and scoring primarily within the database query.
func getRecommendations(db *sql.DB, disliked, preferred []string) ([]PizzaRecommendation, error) {

	// --- Prepare arguments and placeholders safely ---
	dislikedArgs := make([]interface{}, len(disliked))
	for i, v := range disliked {
		dislikedArgs[i] = v
	}

	preferredArgs := make([]interface{}, len(preferred))
	for i, v := range preferred {
		preferredArgs[i] = v
	}

	// Create placeholders like (?,?,?) or return NULL if the list is empty
	// This prevents SQL errors with empty IN clauses.
	dislikedPlaceholders := "NULL"
	if len(dislikedArgs) > 0 {
		dislikedPlaceholders = "(" + strings.Repeat("?,", len(dislikedArgs)-1) + "?)"
	}

	preferredPlaceholders := "NULL"
	if len(preferredArgs) > 0 {
		preferredPlaceholders = "(" + strings.Repeat("?,", len(preferredArgs)-1) + "?)"
	}
	// --- End argument preparation ---

	// --- Construct the main query ---
	// This query does the heavy lifting:
	// 1. Joins pizzas and ingredients.
	// 2. Filters out pizzas containing any disliked ingredients using a subquery.
	// 3. Calculates a score based on the count of preferred ingredients (times 2).
	// 4. Groups results by pizza.
	// 5. Orders by score (desc) and then name (asc).
	// 6. Limits to the top 3.
	query := fmt.Sprintf(`
		SELECT
			p.name,
			COALESCE(GROUP_CONCAT(i.name, '|'), '') as ingredients_concat,
			-- Calculate score: 2 points for each preferred ingredient found in this pizza's ingredients
			SUM(CASE WHEN i.name IN %s THEN 2 ELSE 0 END) as score
		FROM pizzas p
		LEFT JOIN pizza_ingredients pi ON p.id = pi.pizza_id
		LEFT JOIN ingredients i ON pi.ingredient_id = i.id
		WHERE p.id NOT IN (
			-- Subquery: Find IDs of pizzas that contain any disliked ingredient
			SELECT DISTINCT p_inner.id
			FROM pizzas p_inner
			JOIN pizza_ingredients pi_inner ON p_inner.id = pi_inner.pizza_id
			JOIN ingredients i_inner ON pi_inner.ingredient_id = i_inner.id
			WHERE i_inner.name IN %s
		)
		GROUP BY p.id, p.name -- Need to group by the non-aggregated columns
		ORDER BY score DESC, p.name ASC
		LIMIT 3;
	`, preferredPlaceholders, dislikedPlaceholders) // Inject safe placeholders
	// --- End query construction ---

	// Combine arguments in the correct order for the final query
	args := append(preferredArgs, dislikedArgs...)

	// Execute the query
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Query Failed:\nSQL: %s\nArgs: %v\nError: %v", query, args, err) // Log details on failure
		return nil, fmt.Errorf("failed to query recommendations: %w", err)
	}
	defer rows.Close()

	// Process results
	var recommendations []PizzaRecommendation
	for rows.Next() {
		var rec PizzaRecommendation
		var ingredientsConcat sql.NullString

		// Scan into temporary variables
		if err := rows.Scan(&rec.Name, &ingredientsConcat, &rec.Score); err != nil {
			return nil, fmt.Errorf("failed to scan recommendation row: %w", err)
		}

		// Process ingredients string
		if ingredientsConcat.Valid && ingredientsConcat.String != "" {
			rec.Ingredients = strings.Split(ingredientsConcat.String, "|")
			sort.Strings(rec.Ingredients)
		} else {
			rec.Ingredients = []string{}
		}
		recommendations = append(recommendations, rec)
	}
	// Check for errors during row iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating recommendation rows: %w", err)
	}

	return recommendations, nil
}
