package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "your-project-path/docs" // This is where Swag will generate docs.go
)

// @title News API
// @version 1.0
// @description This is a sample news API server.
// @host localhost:8080
// @BasePath /
type News struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Body      string    `json:"body"`
    Author    string    `json:"author"`
    CreatedAt time.Time `json:"created_at"`
}

var news []News

func main() {
    r := gin.Default()

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
    c.JSON(http.StatusOK, news)
}

// @Summary Get a news
// @Description get news by ID
// @Produce json
// @Param id path string true "News ID"
// @Success 200 {object} News
// @Failure 404 {object} gin.H
// @Router /news/{id} [get]
func getNewsById(c *gin.Context) {
    id := c.Param("id")
    for _, item := range news {
        if item.ID == id {
            c.JSON(http.StatusOK, item)
            return
        }
    }
    c.JSON(http.StatusNotFound, gin.H{"message": "Notícia não encontrada"})
}

// @Summary Create a news
// @Description create new news
// @Accept json
// @Produce json
// @Param news body News true "News object"
// @Success 201 {object} News
// @Failure 400 {object} gin.H
// @Router /news [post]
func createNews(c *gin.Context) {
    var newNews News
    if err := c.ShouldBindJSON(&newNews); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    newNews.ID = uuid.New().String()
    newNews.CreatedAt = time.Now()
    news = append(news, newNews)
    c.JSON(http.StatusCreated, newNews)
}

// @Summary Update a news
// @Description update news by ID
// @Accept json
// @Produce json
// @Param id path string true "News ID"
// @Param news body News true "News object"
// @Success 200 {object} News
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Router /news/{id} [put]
func updateNews(c *gin.Context) {
    id := c.Param("id")
    var updatedNews News
    if err := c.ShouldBindJSON(&updatedNews); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    for i, item := range news {
        if item.ID == id {
            updatedNews.ID = id
            updatedNews.CreatedAt = item.CreatedAt
            news[i] = updatedNews
            c.JSON(http.StatusOK, updatedNews)
            return
        }
    }
    c.JSON(http.StatusNotFound, gin.H{"message": "Notícia não encontrada"})
}

// @Summary Delete a news
// @Description delete news by ID
// @Produce json
// @Param id path string true "News ID"
// @Success 200 {object} gin.H
// @Failure 404 {object} gin.H
// @Router /news/{id} [delete]
func deleteNews(c *gin.Context) {
    id := c.Param("id")
    for i, item := range news {
        if item.ID == id {
            news = append(news[:i], news[i+1:]...)
            c.JSON(http.StatusOK, gin.H{"message": "Notícia deletada com sucesso"})
            return
        }
    }
    c.JSON(http.StatusNotFound, gin.H{"message": "Notícia não encontrada"})
}