package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"app/initializers"
	"app/models"

	"github.com/google/uuid"
)

func init() {
	config, err := initializers.LoadConfig(".")
	if err != nil {
		log.Fatal("🚀 Could not load environment variables", err)
	}

	initializers.ConnectDB(&config)
}

// CategorySeed represents the JSON structure for categories
type CategorySeed struct {
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
	Image       string         `json:"image,omitempty"`
	Children    []CategorySeed `json:"children,omitempty"`
}

func main() {
	// Enable UUID extension
	initializers.DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")

	// Migrate models in order (dependencies first)
	err := initializers.DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Post{},
	)

	if err != nil {
		log.Fatal("❌ Migration failed:", err)
	}

	fmt.Println("👍 Migration complete")

	// Check if seed parameter is provided
	if len(os.Args) > 1 && os.Args[1] == "seed" {
		fmt.Println("\n🌴 Starting category seeding...")
		seedCategories()
	}
}

// seedCategories loads and seeds categories from JSON file
func seedCategories() {
	// Check if categories already exist
	var count int64
	initializers.DB.Model(&models.Category{}).Count(&count)
	if count > 0 {
		fmt.Printf("⚠️  Found %d existing categories. Skipping seed.\n", count)
		fmt.Println("💡 To re-seed, clear the categories table first")
		return
	}

	// Load categories from JSON
	fmt.Println("📂 Loading categories from JSON file...")
	categoriesJSON, err := loadCategoriesFromJSON("migrate/seeds/uae_categories.json")
	if err != nil {
		log.Fatal("❌ Failed to load JSON:", err)
	}

	// Seed categories
	fmt.Println("📦 Seeding UAE travel categories...")
	totalCount := seedCategoriesRecursive(categoriesJSON, nil, 0, "")

	fmt.Println("\n🎉 Successfully seeded", totalCount, "UAE travel categories!")

	// Statistics
	var rootCount, level1Count int64
	initializers.DB.Model(&models.Category{}).Where("level = ?", 0).Count(&rootCount)
	initializers.DB.Model(&models.Category{}).Where("level = ?", 1).Count(&level1Count)

	fmt.Println("\n📊 Category Statistics:")
	fmt.Printf("  • Root categories: %d\n", rootCount)
	fmt.Printf("  • Child categories: %d\n", level1Count)
	fmt.Printf("  • Total: %d\n", totalCount)

	fmt.Println("\n✨ UAE Travel Blog is ready!")
}

// loadCategoriesFromJSON reads and parses the JSON file
func loadCategoriesFromJSON(filepath string) ([]CategorySeed, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var categories []CategorySeed
	if err := json.Unmarshal(data, &categories); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return categories, nil
}

// seedCategoriesRecursive recursively creates categories and their children
func seedCategoriesRecursive(categories []CategorySeed, parentID *uuid.UUID, level int, parentPath string) int {
	count := 0
	now := time.Now()

	for _, cat := range categories {
		// Calculate path
		var path string
		if parentPath != "" {
			path = parentPath + "/" + cat.Slug
		} else {
			path = cat.Slug
		}

		// Create category
		category := models.Category{
			Name:        cat.Name,
			Slug:        cat.Slug,
			Description: cat.Description,
			Image:       cat.Image,
			ParentID:    parentID,
			Level:       level,
			Path:        path,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		result := initializers.DB.Create(&category)
		if result.Error != nil {
			log.Printf("❌ Failed to create category '%s': %v", cat.Name, result.Error)
			continue
		}

		count++
		indent := ""
		for i := 0; i < level; i++ {
			indent += "  "
		}
		fmt.Printf("%s✓ %s\n", indent, cat.Name)

		// Recursively create children
		if len(cat.Children) > 0 {
			childCount := seedCategoriesRecursive(cat.Children, &category.ID, level+1, path)
			count += childCount
		}
	}

	return count
}
