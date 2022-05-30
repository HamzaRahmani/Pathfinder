package main

import (
	"fmt"
	"context"
	"encoding/json"
	"log"
	"os"
	"time"
	"net/http"
	
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	_ "go.mongodb.org/mongo-driver/mongo/readpref"
)
var client *mongo.Client

type MonthEntry struct {
	ID    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	User  primitive.ObjectID `json:"user,omitempty" bson:"user,omitempty"`
	Year  int32              `json:"year,omitempty" bson:"year,omitempty"`
	Month int32              `json:"month,omitempty" bson:"month,omitempty"`
	Days  []DayEntry         `json:"days,omitempty" bson:"days,omitempty"`
}

type DayEntry struct {
	Week        int32              `json:"week,omitempty" bson:"week,omitempty"`
	DateTime    primitive.DateTime `json:"datetime,omitempty" bson:"datetime,omitempty"`
	Description string             `jsons:"description,omitempty" bson:"description,omitempty"`
	Priority    int32              `json:"priority,omitempty" bson:"priority,omitempty"`
	IsPermanent bool               `json:"ispermanent,omitempty" bson:"ispermanent,omitempty"`
}

type User struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name,omitempty" bson:"name,omitempty"`
	Email    string             `json:"email,omitempty" bson:"email,omitempty"`
	Password string             `json:"password,omitempty" bson:"password,omitempty"`
}

type Config struct {
	Username     string `json:"dbUser"`
	Password     string `json:"dbPass"`
	ClusterName  string `json:"cluster"`
	MongoURI string `json:"mongoURI"`
}

func LoadConfiguration() (Config, error) {
	var config Config
	configFile, err := os.Open("config.json")
	
	defer configFile.Close()
	if err != nil {
		return config, err
	}
	parse := json.NewDecoder(configFile)
	err = parse.Decode(&config)
	
	return config, err
}

func userSignupEndpoint(c *gin.Context){
	var user User

	if err := c.BindJSON(&user);err!=nil{
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{
				"error": "VALIDATEERR-1",
				"message": "Invalid inputs. Please check your inputs"})
		return
	}
	
	userCollection := client.Database("myDB").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	hash, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	user.Password = string(hash)

	result, err := userCollection.InsertOne(ctx, user)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{
				"error": "Failed to write to db",
				"message": "Please check your inputs"})
		return
	}

	c.JSON(http.StatusAccepted, result)
}
func userLoginEndpoint(c *gin.Context){
	
	var inputUser User
	
	if err := c.BindJSON(&inputUser);err!=nil{
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{
				"error": "VALIDATEERR-1",
				"message": "Invalid inputs. Please check your inputs"})
		return
	}

	var matchedUser User
	userCollection := client.Database("myDB").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	if err := userCollection.FindOne(ctx, bson.M{"email": inputUser.Email}).Decode(&matchedUser); err != nil {
		log.Fatal(err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(matchedUser.Password), []byte(inputUser.Password)); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			gin.H{
				"error": "Not Authorized",
				"message": "Username or Password incorrect"})
		return
	}
	c.JSON(http.StatusAccepted, matchedUser)
}
func createMonthEntryEndpoint(c *gin.Context){

}
func createDayEntryEndpoint(c *gin.Context){

}
func getEntryEndpoint(c *gin.Context){

}

func main() {
	fmt.Println("Pathfinder is launching...")

	var err error
	config, err := LoadConfiguration()
	if err != nil {
		log.Fatal(err)
		return
	}

	client, err = mongo.NewClient(options.Client().ApplyURI("mongodb+srv://" + config.MongoURI))
	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	
	defer client.Disconnect(ctx)

	router := gin.Default()

	router.GET("/user", userLoginEndpoint)
	router.POST("/user", userSignupEndpoint)

	router.GET("/entry/{id}", getEntryEndpoint)
	router.POST("/entry/month", createMonthEntryEndpoint)
	router.POST("/entry/day/{id}", createDayEntryEndpoint)

	router.Run(":8080")
}


