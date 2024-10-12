package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	_ "github.com/psousaj/go-news/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title News API
// @version 1.0
// @description API simples de notícias
// @host localhost:8080
// @BasePath /

type H = map[string]interface{}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type News struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

var dbPool *pgxpool.Pool
var jwtSecret []byte

func main() {
	r := gin.Default()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	connStr := os.Getenv("DATABASE_URL")
    jwtSecret = []byte(os.Getenv("JWT_SECRET"))

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

	// Criar as tabelas se não existirem
	_, err = dbPool.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS users (
            id VARCHAR(36) PRIMARY KEY,
            username TEXT NOT NULL UNIQUE,
            password TEXT NOT NULL
        );
        CREATE TABLE IF NOT EXISTS news (
            id VARCHAR(36) PRIMARY KEY,
            title TEXT NOT NULL,
            body TEXT NOT NULL,
            author VARCHAR(36) NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (author) REFERENCES users(id)
        )
    `)
	if err != nil {
		log.Fatalf("Unable to create tables: %v\n", err)
	}

	r.POST("/register", registerUser)
	r.POST("/login", loginUser)

	// Middleware para autenticação
	auth := r.Group("/")
	auth.Use(AuthMiddleware())
	{
		auth.GET("/news", getNews)
		auth.GET("/news/:id", getNewsById) 
		auth.POST("/news", createNews)
		auth.PUT("/news/:id", updateNews) 
		auth.DELETE("/news/:id", deleteNews) 
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Run(":8080")
}

// Função para gerar o token JWT
func generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour * 72).Unix(), // Expiração em 72 horas
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret) 
}

// Middleware de autenticação
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Request.Header.Get("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, H{"error": "Authorization header is missing"})
			c.Abort()
			return
		}

		claims := &jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil 
		})
        print("OI",token,err)

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, H{"error": "Invalid token", "details": token.Valid})
			c.Abort()
			return
		}

		c.Set("userID", (*claims)["sub"]) // Armazena o ID do usuário no contexto
		c.Next()
	}
}

// @Summary Register a user
// @Description register a new user
// @Accept json
// @Produce json
// @Param user body User true "User object"
// @Success 201 {object} User
// @Failure 400 {object} H
// @Router /register [post]
func registerUser(c *gin.Context) {
	var newUser User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, H{"error": "Erro ao ler o JSON", "details": err.Error()})
		return
	}

	// Validações
	if newUser.Username == "" || newUser.Password == "" {
		c.JSON(http.StatusBadRequest, H{"error": "O username e a senha são obrigatórios"})
		return
	}

	newUser.ID = uuid.New().String()
	// Aqui você deve usar uma função para hash a senha antes de salvar
	// Exemplo: newUser.Password = HashPassword(newUser.Password)

	_, err := dbPool.Exec(context.Background(), 
		"INSERT INTO users (id, username, password) VALUES ($1, $2, $3)",
		newUser.ID, newUser.Username, newUser.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, H{"error": "Erro ao criar usuário", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newUser)
}

// @Summary Login a user
// @Description login user and return JWT
// @Accept json
// @Produce json
// @Param user body User true "User object"
// @Success 200 {string} string "token"
// @Failure 400 {object} H
// @Router /login [post]
// Função de login do usuário
func loginUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, H{"error": "Invalid input"})
		return
	}

	var storedUser User
	err := dbPool.QueryRow(context.Background(), "SELECT id, username, password FROM users WHERE username = $1", user.Username).Scan(&storedUser.ID, &storedUser.Username, &storedUser.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, H{"error": "Invalid username or password"})
		return
	}

	// Aqui você deve verificar a senha (assumindo que você armazena a senha como hash)
	if storedUser.Password != user.Password { // Substitua por comparação de hash real
		c.JSON(http.StatusUnauthorized, H{"error": "Invalid username or password"})
		return
	}

	// Gera o token JWT
	token, err := generateToken(storedUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, H{"token": token})
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

// @Summary Get news by ID
// @Description get news by ID
// @Produce json
// @Success 200 {object} News
// @Failure 404 {object} H
// @Router /news/{id} [get]
func getNewsById(c *gin.Context) {
	id := c.Param("id")
	var n News
	err := dbPool.QueryRow(context.Background(), "SELECT id, title, body, author, created_at FROM news WHERE id = $1", id).Scan(&n.ID, &n.Title, &n.Body, &n.Author, &n.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, H{"error": "Notícia não encontrada"})
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

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, H{"error": "Usuário não autenticado"})
		return
	}
	newNews.Author = userID.(string)

	newNews.ID = uuid.New().String()

	_, err := dbPool.Exec(context.Background(), 
		"INSERT INTO news (id, title, body, author) VALUES ($1, $2, $3, $4)",
		newNews.ID, newNews.Title, newNews.Body, newNews.Author)
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
// @Failure 404 {object} H
// @Router /news/{id} [put]
func updateNews(c *gin.Context) {
	id := c.Param("id")
	var updatedNews News
	if err := c.ShouldBindJSON(&updatedNews); err != nil {
		c.JSON(http.StatusBadRequest, H{"error": "Erro ao ler o JSON", "details": err.Error()})
		return
	}

	_, err := dbPool.Exec(context.Background(),
		"UPDATE news SET title = $1, body = $2 WHERE id = $3",
		updatedNews.Title, updatedNews.Body, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, H{"error": "Erro ao atualizar notícia", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedNews)
}

// @Summary Delete a news
// @Description delete news by ID
// @Produce json
// @Param id path string true "News ID"
// @Success 204
// @Failure 404 {object} H
// @Router /news/{id} [delete]
func deleteNews(c *gin.Context) {
	id := c.Param("id")

	_, err := dbPool.Exec(context.Background(), "DELETE FROM news WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, H{"error": "Erro ao deletar notícia", "details": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
