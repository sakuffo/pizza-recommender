package main

import (
	"net/http"
	"sort"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Pizza represents a pizza with a name and a list of ingredients
type Pizza struct {
	Name        string   `json:"name"`
	Ingredients []string `json:"ingredients"`
}

// UserPreferences contains user's ingredient preferences
type UserPreferences struct {
	Disliked  []string `json:"disliked"`
	Preferred []string `json:"preferred"`
}

// PizzaRecommendation is a pizza with a score based on user preferences
type PizzaRecommendation struct {
	Name        string   `json:"name"`
	Ingredients []string `json:"ingredients"`
	Score       float64  `json:"score"`
}

// Global pizza data
var pizzas = map[string][]string{
	"Margherita":       {"tomato sauce", "mozzarella", "basil"},
	"Pepperoni":        {"tomato sauce", "mozzarella", "pepperoni"},
	"Hawaiian":         {"tomato sauce", "mozzarella", "ham", "pineapple"},
	"Vegetarian":       {"tomato sauce", "mozzarella", "bell peppers", "mushrooms", "onions", "olives"},
	"Meat Lovers":      {"tomato sauce", "mozzarella", "pepperoni", "sausage", "bacon", "ham"},
	"BBQ Chicken":      {"bbq sauce", "mozzarella", "chicken", "red onions", "cilantro"},
	"Supreme":          {"tomato sauce", "mozzarella", "pepperoni", "sausage", "mushrooms", "bell peppers", "onions", "olives"},
	"Buffalo Chicken":  {"buffalo sauce", "mozzarella", "chicken", "red onions", "ranch drizzle"},
	"Veggie Delight":   {"tomato sauce", "mozzarella", "spinach", "tomatoes", "mushrooms", "feta"},
	"Four Cheese":      {"tomato sauce", "mozzarella", "cheddar", "parmesan", "feta"},
	"Mediterranean":    {"olive oil", "mozzarella", "feta", "olives", "sun-dried tomatoes", "spinach"},
	"Pesto Chicken":    {"pesto sauce", "mozzarella", "chicken", "sun-dried tomatoes", "pine nuts"},
	"Bacon & Mushroom": {"tomato sauce", "mozzarella", "bacon", "mushrooms"},
	"White Pizza":      {"olive oil", "mozzarella", "ricotta", "garlic", "spinach"},
	"Taco Pizza":       {"tomato sauce", "mozzarella", "ground beef", "lettuce", "tomatoes", "cheddar", "sour cream"},
}

// getAllIngredients extracts and returns all unique ingredients from our pizza database
func getAllIngredients() []string {
	ingredientsMap := make(map[string]bool)

	for _, ingredients := range pizzas {
		for _, ingredient := range ingredients {
			ingredientsMap[ingredient] = true
		}
	}

	// Convert map keys to slice
	ingredients := make([]string, 0, len(ingredientsMap))
	for ingredient := range ingredientsMap {
		ingredients = append(ingredients, ingredient)
	}

	sort.Strings(ingredients)
	return ingredients
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getRecommendations generates pizza recommendations based on user preferences
func getRecommendations(disliked, preferred []string) []PizzaRecommendation {
	var recommendations []PizzaRecommendation

	for name, ingredients := range pizzas {
		// Skip pizzas with disliked ingredients
		hasDisliked := false
		for _, ingredient := range ingredients {
			if contains(disliked, ingredient) {
				hasDisliked = true
				break
			}
		}

		if hasDisliked {
			continue
		}

		// Calculate score based on preferred ingredients
		score := 0.0

		// Count preferred ingredients in this pizza
		preferredCount := 0
		for _, ingredient := range ingredients {
			if contains(preferred, ingredient) {
				preferredCount++
			}
		}

		// Score is double the number of preferred ingredients
		score = float64(preferredCount * 2)

		// Add to recommendations (even if score is 0)
		recommendations = append(recommendations, PizzaRecommendation{
			Name:        name,
			Ingredients: ingredients,
			Score:       score,
		})
	}

	// Sort recommendations by score (highest first)
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	// Return top 3 or fewer if not enough
	if len(recommendations) > 3 {
		return recommendations[:3]
	}
	return recommendations
}

func main() {
	// Create a Gin router with default middleware
	r := gin.Default()

	// Configure CORS
	// Allow all origins for simplicity. In production, restrict this.
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // Replace with your frontend's origin in production/specific setup
	config.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	r.Use(cors.New(config))

	// --- REMOVED Static File Serving ---
	// r.Static("/static", "./static")

	// --- REMOVED HTML Template Loading ---
	// r.LoadHTMLGlob("templates/*")

	// --- REMOVED Home page route ---
	// r.GET("/", func(c *gin.Context) {
	// 	// Pass all available ingredients to the template
	// 	c.HTML(http.StatusOK, "index.html", gin.H{
	// 		"ingredients": getAllIngredients(),
	// 	})
	// })

	// API endpoint for recommendations
	r.POST("/recommend", func(c *gin.Context) {
		var prefs UserPreferences
		if err := c.ShouldBindJSON(&prefs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		recommendations := getRecommendations(prefs.Disliked, prefs.Preferred)
		c.JSON(http.StatusOK, recommendations)
	})

	// API endpoint for getting all ingredients
	r.GET("/ingredients", func(c *gin.Context) {
		c.JSON(http.StatusOK, getAllIngredients())
	})

	// Start the server - listen on all interfaces for the API
	r.Run("0.0.0.0:8080")
}
