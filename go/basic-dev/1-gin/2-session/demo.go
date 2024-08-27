package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	server := gin.Default()

	db, err := gorm.Open(sqlite.Open("user.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	_ = db.AutoMigrate(&User{})

	var u = &user{
		db: db,
	}

	ug := server.Group("/users")
	// POST /users/signup
	ug.POST("/signup", u.SignUp)

	server.Run(":8080")
}

type user struct {
	db *gorm.DB
}

func (u *user) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
		Name            string `json:"name"`
	}

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入密码不对")
		return
	}

	if len(req.Name) > 8 {
		ctx.String(http.StatusOK, "昵称不得超过8个字符")
		return
	}

	// mock register user ...
	var hashedPassword, err = bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	var timeNowUnixMilli = time.Now().UnixMilli()
	err = u.db.Model(&User{}).Create(&User{
		Name:  req.Name,
		Email: req.Email,
		//Password:  req.Password,
		Password:  string(hashedPassword),
		CreatedAt: timeNowUnixMilli,
		UpdatedAt: timeNowUnixMilli,
	}).Error
	if err != nil {
		ctx.String(http.StatusOK, "注册失败, "+err.Error())
		return
	}

	ctx.String(http.StatusOK, "注册成功")
}
