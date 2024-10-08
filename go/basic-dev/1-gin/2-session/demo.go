package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
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
		db:                  db,
		sessionName:         "sid",
		sessionUserIDKey:    "UserID",
		sessionUpdatedAtKey: "UpdatedAt",
		sessionMaxAge:       60,
	}

	//var store = cookie.NewStore([]byte("secret")) // 使用 cookie 存储 session 信息不太安全
	//var store = memstore.NewStore(
	//	[]byte("ENVnX0XMCYkmUTPKNLmVczmsSDsDOFfG"),
	//	[]byte("1RXEP1cC8uyXIrOG9mR8gvcGT560sRsu"),
	//) // 存储在本地内存中
	var store sessions.Store
	store, err = redis.NewStore(
		16,
		"tcp",
		"192.168.8.51:6379",
		"",
		[]byte("ENVnX0XMCYkmUTPKNLmVczmsSDsDOFfG"),
		[]byte("1RXEP1cC8uyXIrOG9mR8gvcGT560sRsu"),
	)

	server.Use(sessions.Sessions(u.sessionName, store))

	ug := server.Group("/users")
	// POST /users/signup
	ug.POST("/signup", u.SignUp)
	// POST /users/login
	ug.POST("/login", u.Login)

	// 需要登录才能访问的接口组
	var loginAccessGroup = server.Group("/login-access")
	loginAccessGroup.Use(func(ctx *gin.Context) {
		var sess = sessions.Default(ctx) // 本身就会发起查询将该 session 数据全部加载出来
		var userID = sess.Get(u.sessionUserIDKey)
		if userID == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var nowUnixMilli = time.Now().UnixMilli()

		var refresh = func() {
			sess.Set(u.sessionUserIDKey, userID)
			sess.Set(u.sessionUpdatedAtKey, nowUnixMilli)
			sess.Options(sessions.Options{MaxAge: u.sessionMaxAge})
			_ = sess.Save()
		}

		var updatedAtValue = sess.Get(u.sessionUpdatedAtKey)
		if updatedAtValue == nil {
			log.Println("刚登录，还未刷新过")
			// 刚登录，还未刷新过
			refresh()
			return
		}

		var updatedAt, ok = updatedAtValue.(int64)
		if !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if nowUnixMilli-updatedAt > int64(u.sessionMaxAge/6+1)*1000 {
			log.Println("达到允许刷新时间")
			// 达到允许刷新时间
			refresh()
		}
	})
	loginAccessGroup.GET("/hello", u.Hello)

	server.Run(":8080")
}

type user struct {
	db                  *gorm.DB
	sessionName         string
	sessionUserIDKey    string
	sessionUpdatedAtKey string
	sessionMaxAge       int
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

func (u *user) Login(ctx *gin.Context) {
	type Login struct {
		Email    string
		Password string
	}

	var req Login
	var err = ctx.Bind(&req)
	if err != nil {
		return
	}

	if len(req.Email) == 0 {
		ctx.String(http.StatusBadRequest, "Email Required")
		return
	}
	if len(req.Password) == 0 {
		ctx.String(http.StatusBadRequest, "Password Required")
		return
	}

	var userInfo User
	if err = u.db.Model(&User{}).Where("email = ?", req.Email).First(&userInfo).Error; err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(userInfo.Password), []byte(req.Password)); err != nil {
		ctx.String(http.StatusBadRequest, "密码错误")
		return
	}

	// 登录成功设置 session 信息
	var sess = sessions.Default(ctx)
	sess.Set(u.sessionUserIDKey, userInfo.ID)
	sess.Options(sessions.Options{
		MaxAge: u.sessionMaxAge, // 设置过期时间为 60s，用于调试
	})
	_ = sess.Save()

	ctx.String(http.StatusOK, "登录成功")
}

func (u *user) Hello(ctx *gin.Context) {
	ctx.String(http.StatusOK, "Hello")
}
