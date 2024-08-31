package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func main() {
	server := gin.Default()

	server.Use(cors.New(cors.Config{
		//AllowAllOrigins: true,
		//AllowOrigins:     []string{"http://localhost:3000"},
		AllowCredentials: true,

		AllowHeaders:  []string{"Content-Type", "Authorization"}, // jwt token 通过 header Authorization 设置 Bearer xxx
		ExposeHeaders: []string{xJWTToken},                       // 前后端约定使用 `X-Jwt-Token` 返回后端生成的 jwt token
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "your_company.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	var ug = server.Group("/users")
	// POST /users/signup
	ug.POST("/signup", SignUp)
	ug.POST("/login", Login)

	var testGroup = server.Group("/login-access")
	// 需要登录才能访问
	testGroup.Use(func(ctx *gin.Context) {
		var authCode = ctx.GetHeader("Authorization")
		log.Println("authCode:", authCode)
		if !strings.HasPrefix(authCode, bearer) {
			// 没有登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var (
			tokenStr   = authCode[len(bearer):]
			userClaims UserClaims
		)
		log.Println("tokenStr:", tokenStr)

		var token, err = jwt.ParseWithClaims(tokenStr, &userClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		})
		if err != nil {
			// token 伪造的或者过期了
			log.Println(err)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 校验用户身份通过
		var now = time.Now()
		if userClaims.ExpiresAt.Sub(now) < time.Minute {
			// 主动刷新 token
			userClaims.ExpiresAt = jwt.NewNumericDate(now.Add(2 * time.Minute))
			var newTokenStr string
			if newTokenStr, err = token.SignedString([]byte(jwtKey)); err != nil {
				// 用户仍是登录状态，只是刷新 token 失败了
				return
			}

			// 回写到 header 中
			ctx.Header(xJWTToken, newTokenStr)
		}

		// 设置用户信息
		ctx.Set("user", userClaims)
	})
	testGroup.GET("/hello", Hello)

	server.Run(":8080")
}

func SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入密码不对")
		return
	}

	// mock register user ...

	ctx.String(http.StatusOK, "注册成功")
}

func Login(ctx *gin.Context) {
	type Login struct {
		Email    string
		Password string
	}

	var req Login
	var err = ctx.Bind(&req)
	if err != nil {
		return
	}

	// 一些校验

	// 模拟登录
	log.Printf("email: %s, password: %s 登录成功\n", req.Email, req.Password)

	// 生成 token
	var token = jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Minute)), // 2 分钟过期
		},
		Uid: 1, // mock user id
	})
	var tokenStr string
	if tokenStr, err = token.SignedString([]byte(jwtKey)); err != nil {
		ctx.String(http.StatusOK, "系统错误")
	}

	// 回写到 header 中
	ctx.Header(xJWTToken, tokenStr)
	ctx.String(http.StatusOK, "登录成功")
}

func Hello(ctx *gin.Context) {
	ctx.String(http.StatusOK, "Hello, World!")
}

const (
	jwtKey    = "codeporter.pages.dev"
	xJWTToken = "X-Jwt-Token"
	bearer    = "Bearer "
)

type UserClaims struct {
	jwt.RegisteredClaims

	Uid int64 `json:"uid"` // user id
}
