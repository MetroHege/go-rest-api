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

type Point struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"`
}

type Category struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	CategoryName string             `json:"category_name" bson:"category_name"`
}

type Species struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	SpeciesName string             `json:"species_name" bson:"species_name"`
	Image       string             `json:"image" bson:"image"`
	Category    primitive.ObjectID `json:"category,omitempty" bson:"category,omitempty"`
	Location    Point              `json:"location" bson:"location"`
}

type Animal struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	AnimalName string             `json:"animal_name" bson:"animal_name"`
	Birthdate  time.Time          `json:"birthdate" bson:"birthdate"`
	Species    primitive.ObjectID `json:"species,omitempty" bson:"species,omitempty"`
	Location   Point              `json:"location" bson:"location"`
}

var animalCollection *mongo.Collection
var speciesCollection *mongo.Collection
var categoryCollection *mongo.Collection

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	MONGODB_URI := os.Getenv("MONGODB_URI")
	fmt.Println("MONGODB_URI:", MONGODB_URI) // Debug print

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

	app.Get("/api/animals", getAnimals)
	app.Post("/api/animals", createAnimal)
	app.Patch("/api/animals/:id", updateAnimal)
	app.Delete("/api/animals/:id", deleteAnimal)

	app.Get("/api/species", getSpecies)
	app.Post("/api/species", createSpecies)
	app.Patch("/api/species/:id", updateSpecies)
	app.Delete("/api/species/:id", deleteSpecies)

	app.Get("/api/categories", getCategories)
	app.Post("/api/categories", createCategory)
	app.Patch("/api/categories/:id", updateCategory)
	app.Delete("/api/categories/:id", deleteCategory)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Fatal(app.Listen("0.0.0.0:" + port))
}

func getAnimals(c *fiber.Ctx) error {
	var animals []Animal

	cursor, err := animalCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var animal Animal
		if err := cursor.Decode(&animal); err != nil {
			return err
		}
		animals = append(animals, animal)
	}

	return c.JSON(animals)
}

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

func getSpecies(c *fiber.Ctx) error {
	var species []Species

	cursor, err := speciesCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var specie Species
		if err := cursor.Decode(&specie); err != nil {
			return err
		}
		species = append(species, specie)
	}

	return c.JSON(species)
}

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

func updateSpecies(c *fiber.Ctx) error {
	id := c.Params("id")
	ObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	filter := bson.M{"_id": ObjectID}
	update := bson.M{"$set": bson.M{
		"species_name": c.FormValue("species_name"),
		"image":        c.FormValue("image"),
		"category":     c.FormValue("category"),
		"location":     c.FormValue("location"),
	}}

	_, err = speciesCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": "true"})
}

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

func getCategories(c *fiber.Ctx) error {
	var categories []Category

	cursor, err := categoryCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var category Category
		if err := cursor.Decode(&category); err != nil {
			return err
		}
		categories = append(categories, category)
	}

	return c.JSON(categories)
}

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

func updateCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	ObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	filter := bson.M{"_id": ObjectID}
	update := bson.M{"$set": bson.M{
		"category_name": c.FormValue("category_name"),
	}}

	_, err = categoryCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": "true"})
}

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
