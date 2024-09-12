package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Models
// Point struct
type Point struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"`
}

// Category struct
type Category struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	CategoryName string             `json:"category_name" bson:"category_name"`
}

// Species struct
type Species struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	SpeciesName string             `json:"species_name" bson:"species_name"`
	Image       string             `json:"image" bson:"image"`
	Category    primitive.ObjectID `json:"category,omitempty" bson:"category,omitempty"`
	Location    Point              `json:"location" bson:"location"`
}

// Animal struct
type Animal struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	AnimalName string             `json:"animal_name" bson:"animal_name"`
	Birthdate  time.Time          `json:"birthdate" bson:"birthdate"`
	Species    primitive.ObjectID `json:"species,omitempty" bson:"species,omitempty"`
	Location   Point              `json:"location" bson:"location"`
}

// MongoDB collections
var animalCollection *mongo.Collection
var speciesCollection *mongo.Collection
var categoryCollection *mongo.Collection

// Main function
func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	MONGODB_URI := os.Getenv("MONGODB_URI")
	fmt.Println("MONGODB_URI:", MONGODB_URI)

	clientOptions := options.Client().ApplyURI(MONGODB_URI)
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	animalCollection = client.Database("golang_db").Collection("animals")
	speciesCollection = client.Database("golang_db").Collection("species")
	categoryCollection = client.Database("golang_db").Collection("categories")

	app := fiber.New()

	// Animal routes
	app.Get("/api/animals", getAnimals)
	app.Get("/api/animals/:id", getAnimalByID)
	app.Post("/api/animals", createAnimal)
	app.Patch("/api/animals/:id", updateAnimal)
	app.Delete("/api/animals/:id", deleteAnimal)

	// Species routes
	app.Get("/api/species", getSpecies)
	app.Get("/api/species/:id", getSpeciesByID)
	app.Post("/api/species", createSpecies)
	app.Patch("/api/species/:id", updateSpecies)
	app.Delete("/api/species/:id", deleteSpecies)

	// Category routes
	app.Get("/api/categories", getCategories)
	app.Get("/api/categories/:id", getCategoryByID)
	app.Post("/api/categories", createCategory)
	app.Patch("/api/categories/:id", updateCategory)
	app.Delete("/api/categories/:id", deleteCategory)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Fatal(app.Listen("0.0.0.0:" + port))
}

// Animal handlers
// Get all animals
func getAnimals(c *fiber.Ctx) error {
	var animals []bson.M

	// Filtering
	filter := bson.M{}
	if animalName := c.Query("animal_name"); animalName != "" {
		filter["animal_name"] = bson.M{"$regex": animalName, "$options": "i"}
	}
	if speciesName := c.Query("species_name"); speciesName != "" {
		filter["species_info.species_name"] = bson.M{"$regex": speciesName, "$options": "i"}
	}
	if categoryName := c.Query("category_name"); categoryName != "" {
		filter["category_info.category_name"] = bson.M{"$regex": categoryName, "$options": "i"}
	}

	// Sorting
	sort := bson.D{}
	if sortBy := c.Query("sort_by"); sortBy != "" {
		sortOrder := 1
		if c.Query("sort_order") == "desc" {
			sortOrder = -1
		}
		sort = append(sort, bson.E{Key: sortBy, Value: sortOrder})
	} else {
		// Default sort by animal_name in ascending order
		sort = append(sort, bson.E{Key: "animal_name", Value: 1})
	}

	// Pagination
	limit := c.QueryInt("limit", 10)
	skip := c.QueryInt("skip", 0)

	lookupSpeciesStage := bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "species"},
			{Key: "localField", Value: "species"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "species_info"},
		}},
	}

	unwindSpeciesStage := bson.D{
		{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$species_info"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}},
	}

	lookupCategoryStage := bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "categories"},
			{Key: "localField", Value: "species_info.category"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "category_info"},
		}},
	}

	unwindCategoryStage := bson.D{
		{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$category_info"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}},
	}

	matchStage := bson.D{
		{Key: "$match", Value: filter},
	}

	projectStage := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "animal_name", Value: 1},
			{Key: "birthdate", Value: 1},
			{Key: "species", Value: "$species_info.species_name"},
			{Key: "category", Value: "$category_info.category_name"},
			{Key: "location", Value: 1},
		}},
	}

	sortStage := bson.D{
		{Key: "$sort", Value: sort},
	}

	limitStage := bson.D{
		{Key: "$limit", Value: limit},
	}

	skipStage := bson.D{
		{Key: "$skip", Value: skip},
	}

	cursor, err := animalCollection.Aggregate(context.Background(), mongo.Pipeline{
		lookupSpeciesStage, unwindSpeciesStage, lookupCategoryStage, unwindCategoryStage, matchStage, projectStage, sortStage, skipStage, limitStage,
	})
	if err != nil {
		log.Printf("Error during aggregation: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		})
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var animal bson.M
		if err := cursor.Decode(&animal); err != nil {
			log.Printf("Error decoding animal: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}
		animals = append(animals, animal)
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Cursor error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		})
	}

	return c.JSON(animals)
}

// Get an animal by ID
func getAnimalByID(c *fiber.Ctx) error {
	animalID := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(animalID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	lookupSpeciesStage := bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "species"},
			{Key: "localField", Value: "species"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "species_info"},
		}},
	}

	unwindSpeciesStage := bson.D{
		{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$species_info"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}},
	}

	lookupCategoryStage := bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "categories"},
			{Key: "localField", Value: "species_info.category"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "category_info"},
		}},
	}

	unwindCategoryStage := bson.D{
		{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$category_info"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}},
	}

	matchStage := bson.D{
		{Key: "$match", Value: bson.D{
			{Key: "_id", Value: objID},
		}},
	}

	projectStage := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "animal_name", Value: 1},
			{Key: "birthdate", Value: 1},
			{Key: "species", Value: "$species_info.species_name"},
			{Key: "category", Value: "$category_info.category_name"},
			{Key: "location", Value: 1},
		}},
	}

	cursor, err := animalCollection.Aggregate(context.Background(), mongo.Pipeline{
		matchStage, lookupSpeciesStage, unwindSpeciesStage, lookupCategoryStage, unwindCategoryStage, projectStage,
	})
	if err != nil {
		log.Printf("Error during aggregation: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		})
	}
	defer cursor.Close(context.Background())

	if cursor.Next(context.Background()) {
		var animal bson.M
		if err := cursor.Decode(&animal); err != nil {
			log.Printf("Error decoding animal: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}
		return c.JSON(animal)
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": "Animal not found",
	})
}

// Create an animal
func createAnimal(c *fiber.Ctx) error {
	animal := new(Animal)

	if err := c.BodyParser(animal); err != nil {
		return err
	}

	insertResult, err := animalCollection.InsertOne(context.Background(), animal)
	if err != nil {
		return err
	}

	animal.ID = insertResult.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(animal)
}

// Update an animal
func updateAnimal(c *fiber.Ctx) error {
	id := c.Params("id")
	ObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	filter := bson.M{"_id": ObjectID}
	update := bson.M{"$set": bson.M{
		"animal_name": c.FormValue("animal_name"),
		"birthdate":   c.FormValue("birthdate"),
		"species":     c.FormValue("species"),
		"location":    c.FormValue("location"),
	}}

	_, err = animalCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": "true"})
}

// Delete an animal
func deleteAnimal(c *fiber.Ctx) error {
	id := c.Params("id")
	ObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	filter := bson.M{"_id": ObjectID}
	_, err = animalCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": "true"})
}

// Species handlers
// Get all species with filtering, sorting, and pagination
func getSpecies(c *fiber.Ctx) error {
	var species []Species

	// Filtering
	filter := bson.M{}
	if speciesName := c.Query("species_name"); speciesName != "" {
		filter["species_name"] = bson.M{"$regex": speciesName, "$options": "i"}
	}
	if categoryID := c.Query("category_id"); categoryID != "" {
		objID, err := primitive.ObjectIDFromHex(categoryID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid category ID format",
			})
		}
		filter["category"] = objID
	}

	// Sorting
	sort := bson.D{}
	if sortBy := c.Query("sort_by"); sortBy != "" {
		sortOrder := 1
		if c.Query("sort_order") == "desc" {
			sortOrder = -1
		}
		sort = append(sort, bson.E{Key: sortBy, Value: sortOrder})
	} else {
		// Default sort by species_name in ascending order
		sort = append(sort, bson.E{Key: "species_name", Value: 1})
	}

	// Pagination
	limit := c.QueryInt("limit", 10)
	skip := c.QueryInt("skip", 0)

	findOptions := options.Find()
	findOptions.SetSort(sort)
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(skip))

	cursor, err := speciesCollection.Find(context.Background(), filter, findOptions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		})
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var specie Species
		if err := cursor.Decode(&specie); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}
		species = append(species, specie)
	}

	if err := cursor.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		})
	}

	return c.JSON(species)
}

// Get a species by ID
func getSpeciesByID(c *fiber.Ctx) error {
	id := c.Params("id")
	ObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var specie Species
	filter := bson.M{"_id": ObjectID}
	err = speciesCollection.FindOne(context.Background(), filter).Decode(&specie)
	if err != nil {
		return err
	}

	return c.JSON(specie)
}

// Create a species
func createSpecies(c *fiber.Ctx) error {
	specie := new(Species)

	if err := c.BodyParser(specie); err != nil {
		return err
	}

	insertResult, err := speciesCollection.InsertOne(context.Background(), specie)
	if err != nil {
		return err
	}

	specie.ID = insertResult.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(specie)
}

// Update a species
func updateSpecies(c *fiber.Ctx) error {
	id := c.Params("id")
	ObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Invalid ID: %v", err)
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var updateData struct {
		SpeciesName string `json:"species_name"`
		Image       string `json:"image"`
		Category    string `json:"category"`
		Location    Point  `json:"location"`
	}

	if err := c.BodyParser(&updateData); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(400).JSON(fiber.Map{"error": "Failed to parse request body"})
	}

	log.Printf("Update Data: %+v", updateData)

	update := bson.M{"$set": bson.M{}}

	if updateData.SpeciesName != "" {
		update["$set"].(bson.M)["species_name"] = updateData.SpeciesName
	}
	if updateData.Image != "" {
		update["$set"].(bson.M)["image"] = updateData.Image
	}
	if updateData.Location.Type != "" && len(updateData.Location.Coordinates) == 2 {
		update["$set"].(bson.M)["location"] = updateData.Location
	}
	if updateData.Category != "" {
		categoryID, err := primitive.ObjectIDFromHex(updateData.Category)
		if err != nil {
			log.Printf("Invalid category ID: %v", err)
			return c.Status(400).JSON(fiber.Map{"error": "Invalid category ID"})
		}
		update["$set"].(bson.M)["category"] = categoryID
	}

	log.Printf("Update BSON: %+v", update)

	filter := bson.M{"_id": ObjectID}
	result, err := speciesCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("Error updating species: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update species"})
	}

	if result.MatchedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Species not found"})
	}

	return c.JSON(fiber.Map{"success": "true"})
}

// Delete a species
func deleteSpecies(c *fiber.Ctx) error {
	id := c.Params("id")
	ObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	filter := bson.M{"_id": ObjectID}
	_, err = speciesCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": "true"})
}

// Category handlers
// Get all categories with filtering, sorting, and pagination
func getCategories(c *fiber.Ctx) error {
	var categories []Category

	// Filtering
	filter := bson.M{}
	if categoryName := c.Query("category_name"); categoryName != "" {
		filter["category_name"] = bson.M{"$regex": categoryName, "$options": "i"}
	}

	// Sorting
	sort := bson.D{}
	if sortBy := c.Query("sort_by"); sortBy != "" {
		sortOrder := 1
		if c.Query("sort_order") == "desc" {
			sortOrder = -1
		}
		sort = append(sort, bson.E{Key: sortBy, Value: sortOrder})
	} else {
		// Default sort by category_name in ascending order
		sort = append(sort, bson.E{Key: "category_name", Value: 1})
	}

	// Pagination
	limit := c.QueryInt("limit", 10)
	skip := c.QueryInt("skip", 0)

	findOptions := options.Find()
	findOptions.SetSort(sort)
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(skip))

	cursor, err := categoryCollection.Find(context.Background(), filter, findOptions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		})
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var category Category
		if err := cursor.Decode(&category); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}
		categories = append(categories, category)
	}

	if err := cursor.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		})
	}

	return c.JSON(categories)
}

// Get a category by ID
func getCategoryByID(c *fiber.Ctx) error {
	id := c.Params("id")
	ObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var category Category
	filter := bson.M{"_id": ObjectID}
	err = categoryCollection.FindOne(context.Background(), filter).Decode(&category)
	if err != nil {
		return err
	}

	return c.JSON(category)
}

// Create a category
func createCategory(c *fiber.Ctx) error {
	category := new(Category)

	if err := c.BodyParser(category); err != nil {
		return err
	}

	insertResult, err := categoryCollection.InsertOne(context.Background(), category)
	if err != nil {
		return err
	}

	category.ID = insertResult.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(category)
}

// Update a category
func updateCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	ObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var updateData struct {
		CategoryName string `json:"category_name"`
	}

	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Failed to parse request body"})
	}

	fmt.Println("Update Data:", updateData)

	update := bson.M{
		"$set": bson.M{
			"category_name": updateData.CategoryName,
		},
	}

	filter := bson.M{"_id": ObjectID}
	result, err := categoryCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update category"})
	}

	fmt.Println("Update Result:", result)

	return c.JSON(fiber.Map{"message": "Category updated successfully"})
}

// Delete a category
func deleteCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	ObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	filter := bson.M{"_id": ObjectID}
	_, err = categoryCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": "true"})
}
