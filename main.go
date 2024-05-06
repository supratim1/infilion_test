package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// Define structs to represent the data
type Person struct {
	Name string `json:"name"`
}
type Phone struct {
	PhoneNumber string `json:"number"`
}
type Address struct {
	City    string `json:"city"`
	State   string `json:"state"`
	Street1 string `json:"street1"`
	Street2 string `json:"street2"`
	ZipCode string `json:"zip_code"`
}

type PersonInfo struct {
	Person      Person  `json:"person"`
	Phone       Phone   `json:"phone"`
	AddressInfo Address `json:"address"`
}

func initializeDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", "username:password@tcp(localhost:3306)/database_name")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	// Initialize Gin engine
	router := gin.Default()
	// Define endpoint handler
	router.GET("/person/:person_id/info", func(c *gin.Context) {

		// Extract person ID from path parameter
		personIDStr := c.Param("person_id")
		personID, err := strconv.Atoi(personIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid person ID"})
			return
		}

		// Connect to the database
		db, err := sql.Open("mysql", "username:password@tcp(localhost:3306)/database_name")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to the database"})
			return
		}
		defer db.Close()

		// Query the person table
		personRow := db.QueryRow("SELECT name FROM person WHERE id = ?", personID)
		var person Person
		err = personRow.Scan(&person.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch person information"})
			return
		}

		phnRow := db.QueryRow("SELECT number FROM phone WHERE person_id = ?", personID)
		var phone Phone
		err = phnRow.Scan(&phone.PhoneNumber)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch person information"})
			return
		}

		// Query the address table
		addressRow := db.QueryRow("SELECT city, state, street1, street2, zip_code FROM address a inner join address_join aj on aj.address_id=a.id WHERE person_id  = ?", personID)
		var address Address
		err = addressRow.Scan(&address.City, &address.State, &address.Street1, &address.Street2, &address.ZipCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch address information"})
			return
		}

		// Construct response
		personInfo := PersonInfo{
			Person:      person,
			Phone:       phone,
			AddressInfo: address,
		}

		// Return JSON response
		c.JSON(http.StatusOK, personInfo)

	})

	// Define endpoint handler for creating a new person entry
	router.POST("/person/create", func(c *gin.Context) {
		// Parse request body
		var personAddress PersonInfo
		if err := c.BindJSON(&personAddress); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Connect to the database
		db, err := initializeDB()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to the database"})
			return
		}
		defer db.Close()
		// Insert new person into the database
		result, err := db.Exec("INSERT INTO person (name) VALUES (?)", personAddress.Person.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert person into the database"})
			return
		}

		// Get the ID of the newly inserted person
		personID, err := result.LastInsertId()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get person ID"})
			return
		}

		_, err = db.Exec("INSERT INTO phone ( person_id, number) VALUES (?, ?)", personID, personAddress.Phone.PhoneNumber)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert person into the database"})
			return
		}

		// Insert address into the database
		res, err := db.Exec("INSERT INTO address (city, state, street1, street2, zip_code) VALUES (?, ?, ?, ?, ?)",
			personAddress.AddressInfo.City, personAddress.AddressInfo.State, personAddress.AddressInfo.Street1,
			personAddress.AddressInfo.Street2, personAddress.AddressInfo.ZipCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert address into the database"})
			return
		}

		addID, err := res.LastInsertId()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get person ID"})
			return
		}
		_, err = db.Exec("INSERT INTO address_join ( person_id, address_id) VALUES (?, ?)", personID, addID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert person into the database"})
			return
		}

		// Return success response
		c.JSON(http.StatusOK, gin.H{"message": "Person created successfully", "person_id": personID})
	})
	// Start server
	router.Run(":8080")
}
