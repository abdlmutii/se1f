package main

import (
	"math/rand"
	"time"
  "fmt"
  "os"
  "net/http"
  "io/ioutil"
  "strconv"
  "github.com/gofiber/template/html/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/nanobox-io/golang-scribble"
)

type Data struct {
	Name    string `json:"name"`
	Code    string `json:"code"`
  Lang    string `jsom:"lang"`
}

/* Random UUID */
func uuid(length int) string {
	rand.Seed(time.Now().UnixNano())
  pool := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789-_$€"
  
	uuid := make([]byte, length)
	for i := 0; i < length; i++ {
		uuid[i] = pool[rand.Intn(len(pool))]
	}

	return string(uuid)
}

func main() { 
  /* SETTING UP ENGINE, APP, DB, UPTIME */
  engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
        Views: engine,
    })
	db, _ := scribble.New("./bins", nil)

  /* MAIN ROUTE */
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("welcome.html")
	})
  
  /* /:id API ENDPOINT */
	app.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var tbin Data
		err := db.Read("bins", id, &tbin)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(map[string]interface{}{
				"error": "Bin not found",
			})
		}
    unesc, err := strconv.Unquote(tbin.Code)
    tbin.Code = unesc
		return c.Render("index", tbin)
	})

  /* /create API ENDPOINT */
	app.Post("/create", func(c *fiber.Ctx) error {
		var bin Data
		err := c.BodyParser(&bin)
    fmt.Println("err: ", err)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
				"error": "Invalid request payload",
			})
		}
		id := uuid(4)
		err = db.Write("bins", id, bin)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{
				"error": "Failed to create bin",
			})
		}
		return c.JSON(map[string]string{
			"url": "https://se1f.repl.co/" + id,
		})
	})

  /* SERVER LISTEN */
	err := app.Listen(":3000")
	if err != nil {
		panic(err)
	}

  /* UPTIME */
  ticker := time.NewTicker(1 * time.Minute)

	go func() {
		for range ticker.C {
			url := fmt.Sprintf("https://api.shuruhatik.com/uptime/%s/%s/%s%s", os.Getenv("REPL_ID"), os.Getenv("REPL_SLUG"), os.Getenv("REPL_OWNER"), "/")
			response, err := http.Get(url)
			if err != nil {
				fmt.Println(err)
				return
			}

			defer response.Body.Close()
			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				fmt.Println("Error reading response body:", err)
        fmt.Println(body)
				return
			}
		}
	}()

	select {}
}
