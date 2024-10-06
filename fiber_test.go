package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/mustache/v2"
	"github.com/stretchr/testify/assert"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var engine = mustache.New("./template", ".mustache")

var app = fiber.New(fiber.Config{
	Views:        engine,
	IdleTimeout:  time.Minute * 5,
	ReadTimeout:  time.Minute * 5,
	WriteTimeout: time.Minute * 5,
	ErrorHandler: func(ctx *fiber.Ctx, err error) error {
		ctx.Status(fiber.StatusInternalServerError)
		return ctx.SendString("Error: " + err.Error())
	},
})

func TestRoutingHelloWorld(t *testing.T) {
	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello World")
	})

	request := httptest.NewRequest("GET", "/", nil)
	response, err := app.Test(request)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, 200, response.StatusCode)
}

func TestCtx(t *testing.T) {
	app.Get("/hello", func(ctx *fiber.Ctx) error {
		query := ctx.Query("name", "Guest")
		return ctx.SendString("Hello " + query)
	})

	request := httptest.NewRequest("GET", "/hello?name=Brian", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	bytes, err := io.ReadAll(response.Body)
	assert.Equal(t, "Hello Brian", string(bytes))
}

func TestHttpRequest(t *testing.T) {
	app.Get("/request", func(ctx *fiber.Ctx) error {
		first := ctx.Get("firstname")
		last := ctx.Cookies("lastname")
		return ctx.SendString("Hello " + first + " " + last)
	})

	request := httptest.NewRequest("GET", "/request", nil)
	request.Header.Set("firstname", "Brian")
	request.AddCookie(&http.Cookie{Name: "lastname", Value: "Anashari"})
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	bytes, err := io.ReadAll(response.Body)
	assert.Equal(t, "Hello Brian Anashari", string(bytes))
}

func TestRouteParameter(t *testing.T) {
	app.Get("/users/:userId/orders/:orderId", func(ctx *fiber.Ctx) error {
		userId := ctx.Params("userId")
		orderId := ctx.Params("orderId")
		return ctx.SendString("User: " + userId + " with order: " + orderId)
	})

	request := httptest.NewRequest("GET", "/users/2/orders/5", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	bytes, err := io.ReadAll(response.Body)
	assert.Equal(t, "User: 2 with order: 5", string(bytes))
}

func TestFormRequest(t *testing.T) {
	app.Get("/hello", func(ctx *fiber.Ctx) error {
		name := ctx.FormValue("name")
		return ctx.SendString("Hello " + name)
	})

	body := strings.NewReader("name=Brian")
	request := httptest.NewRequest("GET", "/hello", body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	bytes, err := io.ReadAll(response.Body)
	assert.Equal(t, "Hello Brian", string(bytes))
}

//go:embed source/file.txt
var contohFile []byte

func TestFormUpload(t *testing.T) {
	app.Post("/upload", func(ctx *fiber.Ctx) error {
		file, err := ctx.FormFile("file")
		if err != nil {
			panic(err)
		}

		err = ctx.SaveFile(file, "./target/"+file.Filename)
		if err != nil {
			panic(err)
		}

		return ctx.SendString("Uploaded successfully")
	})

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	file, err2 := writer.CreateFormFile("file", "file.txt")
	assert.Nil(t, err2)
	_, err2 = file.Write(contohFile)
	assert.Nil(t, err2)
	err := writer.Close()
	assert.Nil(t, err)

	request := httptest.NewRequest("POST", "/upload", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	bytes, err := io.ReadAll(response.Body)
	assert.Equal(t, "Uploaded successfully", string(bytes))
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func TestRequestBody(t *testing.T) {
	app.Post("/login", func(ctx *fiber.Ctx) error {
		body := ctx.Body()

		request := new(LoginRequest)
		err := json.Unmarshal(body, request)
		if err != nil {
			panic(err)
		}

		return ctx.SendString("Hello " + request.Username)
	})

	body := strings.NewReader(`{"username":"Brian", "password":"12345"}`)
	request := httptest.NewRequest("POST", "/login", body)
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	bytes, err := io.ReadAll(response.Body)
	assert.Equal(t, "Hello Brian", string(bytes))
}

type RegisterRequest struct {
	Username string `json:"username" form:"username" xml:"username"`
	Password string `json:"password" form:"password" xml:"password"`
}

func TestBodyParser(t *testing.T) {
	app.Post("/register", func(ctx *fiber.Ctx) error {
		request := new(RegisterRequest)
		err := ctx.BodyParser(request)
		if err != nil {
			panic(err)
		}

		return ctx.SendString("Hello " + request.Username)
	})
}

func TestBodyParserJSON(t *testing.T) {
	TestBodyParser(t)

	body := strings.NewReader(`{"username":"Brian", "password":"12345"}`)
	request := httptest.NewRequest("POST", "/register", body)
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	bytes, err := io.ReadAll(response.Body)
	assert.Equal(t, "Hello Brian", string(bytes))
}

func TestBodyParserForm(t *testing.T) {
	TestBodyParser(t)

	body := strings.NewReader(`username=Brian&password=12345`)
	request := httptest.NewRequest("POST", "/register", body)
	request.Header.Set("Content-Type", "x-www-form-urlencoded")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	bytes, err := io.ReadAll(response.Body)
	assert.Equal(t, "Hello Brian", string(bytes))
}

func TestBodyParserXML(t *testing.T) {
	TestBodyParser(t)

	body := strings.NewReader(`
		<RegisterRequest>
			<username>Brian</username>
			<password>12345</password>
		</RegisterRequest>	
	`)
	request := httptest.NewRequest("POST", "/register", body)
	request.Header.Set("Content-Type", "application/xml")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	bytes, err := io.ReadAll(response.Body)
	assert.Equal(t, "Hello Brian", string(bytes))
}

func TestResponseJSON(t *testing.T) {
	app.Get("/user", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"username": "Brian",
			"password": "12345",
		})
	})

	request := httptest.NewRequest(http.MethodGet, "/user", nil)
	request.Header.Set("Accept", "application/json")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	bytes, err := io.ReadAll(response.Body)
	assert.Equal(t, `{"password":"12345","username":"Brian"}`, string(bytes))
}

func TestDownloadFile(t *testing.T) {
	app.Get("/download", func(ctx *fiber.Ctx) error {
		return ctx.Download("./source/file.txt", "file.txt")
	})

	request := httptest.NewRequest(http.MethodGet, "/download", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, `attachment; filename="file.txt"`, response.Header.Get("Content-Disposition"))
	bytes, err := io.ReadAll(response.Body)
	assert.Equal(t, `this a sample file`, string(bytes))
}

func TestRoutingGroup(t *testing.T) {
	helloWorld := func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello World")
	}

	api := app.Group("/api")
	api.Get("/hello", helloWorld)
	api.Get("/world", helloWorld)

	web := app.Group("/web")
	web.Get("/hello", helloWorld)
	web.Get("/world", helloWorld)

	request1 := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	response, err := app.Test(request1)
	assert.Nil(t, err)
	bytes, err := io.ReadAll(response.Body)
	assert.Equal(t, `Hello World`, string(bytes))

	request2 := httptest.NewRequest(http.MethodGet, "/web/world", nil)
	response, err = app.Test(request2)
	assert.Nil(t, err)
	bytes, err = io.ReadAll(response.Body)
	assert.Equal(t, `Hello World`, string(bytes))
}

func TestStatic(t *testing.T) {
	app.Static("/public", "./source")

	request := httptest.NewRequest(http.MethodGet, "/public/file.txt", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	bytes, err := io.ReadAll(response.Body)
	assert.Equal(t, `this a sample file`, string(bytes))
}

func TestErrorHandling(t *testing.T) {
	app.Get("/error", func(ctx *fiber.Ctx) error {
		return errors.New("ups")
	})

	request := httptest.NewRequest(http.MethodGet, "/error", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 500, response.StatusCode)
	bytes, err := io.ReadAll(response.Body)
	assert.Equal(t, "Error: ups", string(bytes))
}

func TestView(t *testing.T) {
	app.Get("/view", func(ctx *fiber.Ctx) error {
		return ctx.Render("index", fiber.Map{
			"title":   "Hello Title",
			"header":  "Hello Header",
			"content": "Hello Content",
		})
	})

	request := httptest.NewRequest(http.MethodGet, "/view", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	bytes, err := io.ReadAll(response.Body)
	fmt.Println(string(bytes))
	assert.Contains(t, string(bytes), "Hello Title")
	assert.Contains(t, string(bytes), "Hello Header")
	assert.Contains(t, string(bytes), "Hello Content")
}

func TestClient(t *testing.T) {
	client := fiber.AcquireClient()
	defer fiber.ReleaseClient(client)

	agent := client.Get("https://example.com")
	status, response, err := agent.String()
	assert.Nil(t, err)
	assert.Equal(t, 200, status)
	assert.Contains(t, response, "Example Domain")
}
