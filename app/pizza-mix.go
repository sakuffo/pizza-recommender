package main

import (
	"log"
	"net/http"
	"os"            // For checking if db file exists, path manipulation
	"path/filepath" // For creating data directory

	// For building IN clauses safely
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	// NOTE: database/sql and the driver (_ "github.com/mattn/go-sqlite3")
	// are now only directly imported in db.go
)

// --- Data Structures for API ---

// APIUserPreferences contains user's ingredient preferences from the request body.
type APIUserPreferences struct {
	Disliked  []string `json:"disliked"`
	Preferred []string `json:"preferred"`
}

// APIPizzaRecommendation is the structure returned by the /recommend endpoint.
type APIPizzaRecommendation struct {
	Name        string   `json:"name"`
	Ingredients []string `json:"ingredients"`
	Score       float64  `json:"score"`
}

// --- Initial Data (Used ONLY for seeding the database via db.go) ---

// pizzas holds the initial data to populate the database if it's empty.
// This map is NOT used for serving requests after the initial seed.
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

// --- Main Application ---

func main() {
	// Define path for the database file relative to the executable
	dbDir := "./data" // Store DB in a 'data' subdirectory
	dbPath := filepath.Join(dbDir, "pizza.db")
	log.Printf("Database path set to: %s", dbPath)

	// Ensure the data directory exists
	if err := os.MkdirAll(dbDir, 0755); err != nil { // 0755 permissions
		log.Fatalf("FATAL: Failed to create data directory '%s': %v", dbDir, err)
	}

	// Check if the database needs seeding (e.g., doesn't exist)
	// This is a simple check; more robust checks could verify table existence/counts.
	needsSeed := false
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Printf("Database file '%s' not found. Will attempt to create and seed.", dbPath)
		needsSeed = true
	} else if err != nil {
		// Handle other potential stat errors (e.g., permissions)
		log.Fatalf("FATAL: Error checking database file '%s': %v", dbPath, err)
	}

	// 1. Connect to the database (function from db.go)
	db, err := connectDB(dbPath)
	if err != nil {
		log.Fatalf("FATAL: DB connection failed: %v", err)
	}
	// Use defer to ensure the database connection is closed cleanly on exit
	defer func() {
		log.Println("Closing database connection...")
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// 2. Run Migrations (function from db.go)
	if err := migrate(db); err != nil {
		log.Fatalf("FATAL: Migration failed: %v", err)
	}

	// 3. Seed Data (only if the file didn't exist initially) (function from db.go)
	if needsSeed {
		if err := seed(db, pizzas); err != nil {
			// Log as fatal because if seeding fails on first run, app is likely unusable
			log.Fatalf("FATAL: Seeding failed: %v", err)
		}
	} else {
		log.Println("Database already exists. Skipping seed.")
	}

	// --- Setup Gin Router and Middleware ---
	gin.SetMode(gin.ReleaseMode) // Use ReleaseMode for production, or gin.DebugMode
	r := gin.New()               // Use New instead of Default to have more control
	r.Use(gin.Logger())          // Add logger middleware
	r.Use(gin.Recovery())        // Add recovery middleware to handle panics

	// CORS Configuration (adjust origins for production)
	config := cors.DefaultConfig()
	// config.AllowOrigins = []string{"http://your-frontend-domain.com", "http://localhost:xxxx"} // Production example
	config.AllowOrigins = []string{"*"} // Allow all for now, restrict in production
	config.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"} // Add common headers
	r.Use(cors.New(config))

	// --- API Endpoints ---

	// Health Check Endpoint (Good Practice)
	r.GET("/health", func(c *gin.Context) {
		// Check DB connection as part of health
		if err := db.Ping(); err != nil {
			log.Printf("WARN: Health check failed - DB ping error: %v", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": "database connection issue"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Endpoint for getting all available ingredients
	r.GET("/ingredients", func(c *gin.Context) {
		// Call repository function, passing the DB connection
		ingredientsList, err := getAllIngredients(db) // Function from repository.go
		if err != nil {
			log.Printf("ERROR fetching ingredients: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve ingredients"})
			return
		}
		c.JSON(http.StatusOK, ingredientsList)
	})

	// Endpoint for getting pizza recommendations
	r.POST("/recommend", func(c *gin.Context) {
		var prefs APIUserPreferences // Use the API-specific struct for binding
		if err := c.ShouldBindJSON(&prefs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}

		// Call repository function, passing DB connection and preferences
		// The repository function returns []PizzaRecommendation
		recommendations, err := getRecommendations(db, prefs.Disliked, prefs.Preferred) // Function from repository.go
		if err != nil {
			log.Printf("ERROR generating recommendations: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate recommendations"})
			return
		}

		// Convert internal []PizzaRecommendation to []APIPizzaRecommendation for the response
		// (In this case, they have the same fields, but this shows the pattern if they differed)
		apiRecs := make([]APIPizzaRecommendation, len(recommendations))
		for i, rec := range recommendations {
			apiRecs[i] = APIPizzaRecommendation{
				Name:        rec.Name,
				Ingredients: rec.Ingredients,
				Score:       rec.Score,
			}
		}

		c.JSON(http.StatusOK, apiRecs)
	})

	// --- Start Server ---
	listenAddr := "0.0.0.0:8080" // Listen on all interfaces
	log.Printf("Starting HTTP server on %s", listenAddr)
	if err := r.Run(listenAddr); err != nil && err != http.ErrServerClosed {
		// Log fatal only if it's not a clean shutdown
		log.Fatalf("FATAL: Failed to run HTTP server: %v", err)
	}
}
