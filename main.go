package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	_ "github.com/psousaj/go-news/docs" // This is where Swag will generate docs.go
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title News API
// @version 1.0
// @description API simples de notícias
// @host localhost:8080
// @BasePath /

// Define o tipo H como um alias para map[string]interface{}
type H = map[string]interface{}

type News struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Body      string    `json:"body"`
    Author    string    `json:"author"`
    CreatedAt time.Time `json:"created_at"`
}

var dbPool *pgxpool.Pool

func main() {
    r := gin.Default()

    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
    connStr := os.Getenv("DATABASE_URL")
    
    // Criando um pool de conexões
    dbPool, err = pgxpool.New(context.Background(), connStr)
    if err != nil {
        log.Fatalf("Unable to create connection pool: %v\n", err)
    }
    defer dbPool.Close()

    // Verificar a conexão
    if err := dbPool.Ping(context.Background()); err != nil {
        log.Fatalf("Unable to connect to database: %v\n", err)
    }

    // Criar a tabela se não existir
    _, err = dbPool.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS news (
            id VARCHAR(36) PRIMARY KEY,
            title TEXT NOT NULL,
            body TEXT NOT NULL,
            author TEXT NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        )
    `)
    if err != nil {
        log.Fatalf("Unable to create table: %v\n", err)
    }

    // Rotas da API
    r.GET("/news", getNews)
    r.GET("/news/:id", getNewsById)
    r.POST("/news", createNews)
    r.PUT("/news/:id", updateNews)
    r.DELETE("/news/:id", deleteNews)

    // Rota do Swagger
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // Iniciar o servidor na porta 8080
    r.Run(":8080")
}

// @Summary List all news
// @Description get all news
// @Produce json
// @Success 200 {array} News
// @Router /news [get]
func getNews(c *gin.Context) {
    rows, err := dbPool.Query(context.Background(), "SELECT id, title, body, author, created_at FROM news")
    if err != nil {
        c.JSON(http.StatusInternalServerError, H{"error": "Error fetching news"})
        return
    }
    defer rows.Close()

    var newsList []News
    for rows.Next() {
        var n News
        err := rows.Scan(&n.ID, &n.Title, &n.Body, &n.Author, &n.CreatedAt)
        if err != nil {
            c.JSON(http.StatusInternalServerError, H{"error": "Error scanning news"})
            return
        }
        newsList = append(newsList, n)
    }

    c.JSON(http.StatusOK, newsList)
}

// @Summary Get a news
// @Description get news by ID
// @Produce json
// @Param id path string true "News ID"
// @Success 200 {object} News
// @Failure 404 {object} H
// @Router /news/{id} [get]
func getNewsById(c *gin.Context) {
    id := c.Param("id")
    var n News
    err := dbPool.QueryRow(context.Background(), "SELECT id, title, body, author, created_at FROM news WHERE id = $1", id).Scan(&n.ID, &n.Title, &n.Body, &n.Author, &n.CreatedAt)
    if err == pgx.ErrNoRows {
        c.JSON(http.StatusNotFound, H{"message": "Notícia não encontrada"})
        return
    } else if err != nil {
        c.JSON(http.StatusInternalServerError, H{"error": "Error fetching news"})
        return
    }
    c.JSON(http.StatusOK, n)
}

// @Summary Create a news
// @Description create new news
// @Accept json
// @Produce json
// @Param news body News true "News object"
// @Success 201 {object} News
// @Failure 400 {object} H
// @Router /news [post]
func createNews(c *gin.Context) {
    var newNews News
    if err := c.ShouldBindJSON(&newNews); err != nil {
        c.JSON(http.StatusBadRequest, H{"error": "Erro ao ler o JSON", "details": err.Error()})
        return
    }

    // Validações
    if newNews.Title == "" {
        c.JSON(http.StatusBadRequest, H{"error": "O título é obrigatório"})
        return
    }
    if newNews.Body == "" {
        c.JSON(http.StatusBadRequest, H{"error": "O corpo da notícia é obrigatório"})
        return
    }
    if newNews.Author == "" {
        c.JSON(http.StatusBadRequest, H{"error": "O autor é obrigatório"})
        return
    }

    newNews.ID = uuid.New().String()
    newNews.CreatedAt = time.Now()

    _, err := dbPool.Exec(context.Background(), 
        "INSERT INTO news (id, title, body, author, created_at) VALUES ($1, $2, $3, $4, $5)",
        newNews.ID, newNews.Title, newNews.Body, newNews.Author, newNews.CreatedAt)
    if err != nil {
        c.JSON(http.StatusInternalServerError, H{"error": "Erro ao criar notícia", "details": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, newNews)
}

// @Summary Update a news
// @Description update news by ID
// @Accept json
// @Produce json
// @Param id path string true "News ID"
// @Param news body News true "News object"
// @Success 200 {object} News
// @Failure 400 {object} H
// @Failure 404 {object} H
// @Router /news/{id} [put]
func updateNews(c *gin.Context) {
    id := c.Param("id")
    var updatedNews News
    if err := c.ShouldBindJSON(&updatedNews); err != nil {
        c.JSON(http.StatusBadRequest, H{"error": err.Error()})
        return
    }

    _, err := dbPool.Exec(context.Background(), "UPDATE news SET title = $1, body = $2, author = $3 WHERE id = $4",
        updatedNews.Title, updatedNews.Body, updatedNews.Author, id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, H{"error": "Error updating news"})
        return
    }

    updatedNews.ID = id
    c.JSON(http.StatusOK, updatedNews)
}

// @Summary Delete a news
// @Description delete news by ID
// @Produce json
// @Param id path string true "News ID"
// @Success 200 {object} H
// @Failure 404 {object} H
// @Router /news/{id} [delete]
func deleteNews(c *gin.Context) {
    id := c.Param("id")
    
    res, err := dbPool.Exec(context.Background(), "DELETE FROM news WHERE id = $1", id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, H{"error": "Error deleting news"})
        return
    }

    if res.RowsAffected() == 0 {
        c.JSON(http.StatusNotFound, H{"message": "Notícia não encontrada"})
        return
    }

    c.JSON(http.StatusOK, H{"message": "Notícia deletada com sucesso"})
}