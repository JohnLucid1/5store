package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	_ "github.com/lib/pq"
)

const (
	host                   = "localhost"
	port                   = 5432
	user                   = "postgres"
	password               = "mysecretpassword"
	dbname                 = "postgres"
	app_port               = ":3000"
	url_length             = 5
	time_to_delete_minutes = 5
)

var DBCon *sql.DB = ConnectDB()

/* NOTE: Api resposnse should be
{
	success: bool
	data: somedata or nil (or if success if false, message like "not found in db, no record, so on...")
}

  NOTE: API
	POST "/create/" + json{"data": "someomseom"}
	Get "/old/" + json {"url": "someurl"}
  NOTE: returns (All requests will return {
	"success": bool,
	"data": data||nil
  })
*/

func create_text(c *fiber.Ctx) error {
	text := struct {
		Data string `json:"data"`
	}{}

	if err := c.BodyParser(&text); err != nil {
		log.Println(err)
		response := fiber.Map{"success": false, "data": nil}
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	new_url := Create_url(url_length)
	newStr, err := CreateData(DBCon, text.Data, new_url)
	if err != nil {
		log.Println(err)
		response := fiber.Map{"success": false, "data": nil}
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	response := fiber.Map{
		"success": true,
		"data":    newStr,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func show_text(c *fiber.Ctx) error {
	text := struct {
		Url string `json:"url"`
	}{}

	if err := c.BodyParser(&text); err != nil {
		log.Println(err)
		response := fiber.Map{"success": false, "data": nil}
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	old_data, err := Find_url_data(DBCon, text.Url)
	if err != nil {
		log.Println(err)
		response := fiber.Map{"success": false, "data": nil}
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	response := fiber.Map{
		"success": true,
		"data":    old_data,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func main() {
	app := fiber.New()

	go DeleteOldRecords()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "*",
		AllowMethods: "POST, GET",
	}))
	app.Post("/create/", create_text)
	app.Get("/old/", show_text)

	app.Listen(app_port)
}

func Create_url(length int) string {
	buff := ""
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_"
	min := 0
	max := len(charset)
	for i := 0; i < length; i++ {
		idx := rand.Intn(max-min) + min
		buff += string(charset[idx])
	}
	return buff
}

func ConnectDB() *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalln(err)
	}

	return db
}

type Data struct {
	Data_id     int    `json:"data_id"`
	Data        string `json:"data"`
	Created_url string `json:"created_url"`
	Created_at  string `json:"created_at"`
}

func CreateData(db *sql.DB, data string, new_url string) (string, error) {
	newUrl := Create_url(url_length)

	_, err := db.Exec("INSERT INTO Links (data, created_url) VALUES($1, $2)", data, newUrl)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return newUrl, nil

}

func Find_url_data(db *sql.DB, old_url string) (Data, error) {
	query := "SELECT data_id, data, created_at FROM Links WHERE created_url = $1 LIMIT 1"

	row := db.QueryRow(query, old_url)

	var data Data
	if err := row.Scan(&data.Data_id, &data.Data, &data.Created_at); err != nil {
		log.Println(err)
		return Data{}, err
	}

	return data, nil
}

func DeleteOldRecords() {
	for {
		cutoffTime := time.Now().Add(-time_to_delete_minutes * time.Minute)
		_, err := DBCon.Exec("DELETE FROM Links WHERE created_at < $1", cutoffTime)
		if err != nil {
			log.Println("Error deleting old records: ", err)
		}
		time.Sleep(1 * time.Minute)
	}
}
