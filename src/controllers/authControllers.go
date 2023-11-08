package controllers

import (
	"go-gin-auth/src/common"
	"go-gin-auth/src/initializers"
	"go-gin-auth/src/models"
	"go-gin-auth/src/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost = 10
)

type RegisterInput struct {
	Name            string `json:"name" binding:"required,min=3,max=50"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=8"`
}

type SignInInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
	var input RegisterInput

	// get db
	db := initializers.GetDB()

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// validation checks
	if input.Password != input.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password do not match."})
		return
	}

	// check if user already exists
	result := db.Where("email = ?", input.Email).First(&models.User{})
	if result.Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists."})
		return
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcryptCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// create user
	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
	}

	result = db.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// generate jwt tokens
	accessSignedToken, refreshSignedToken, err := utils.GenerateJWTTokensForUser(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// set http only cookies (name, value, maxAge, path, domain, secure, httpOnly)
	c.SetCookie("refresh_token", refreshSignedToken, 86400, "/", "localhost", false, true)

	c.JSON(http.StatusOK, common.APIResponse{
		Success: true,
		Message: "User created successfully",
		Data: gin.H{
			"access_token": accessSignedToken,
			"user":         user,
		},
	})
}

// sign in controller
// func SignIn(c *gin.Context) {
// 	var input SignInInput
// 	if err := c.ShouldBindJSON(&input); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	var user models.User
// 	result := models.DB.Where("email = ?", input.Email).First(&user)
// 	if result.Error != nil {
// 		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
// 			return
// 		}
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user"})
// 		return
// 	}

// 	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
// 		return
// 	}

// 	token, err := utils.GenerateToken(user.ID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"token": token})
// }
