package web

import (
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	jwt "github.com/golang-jwt/jwt/v5"
	"go_demo/webook/internal/domain"
	"go_demo/webook/internal/service"
	"time"

	//"github.com/gin-gonic/contrib"
	"github.com/gin-gonic/gin"
	"net/http"
)

// UserHandler 定义和用户有关的路由
type UserHandler struct {
	svc         *service.UserService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
}

// NewUserHandler 预编译正	则表达式提高校验速度
func NewUserHandler(svc *service.UserService) *UserHandler {
	const (
		emailRegexPattern    = `^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`                     //邮箱校验
		passwordRegexPattern = `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$` //密码校验
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	return &UserHandler{
		svc:         svc,
		emailExp:    emailExp,
		passwordExp: passwordExp,
	}
}

// RegisterRoutes 注册路由组
func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	//server.POST("/users/signup", u.SignUp)
	//server.POST("/users/login", u.LogIn)
	//server.POST("/users/edit", u.Edit)
	//server.GET("/users/profile", u.Profile)
	//注册路由组  usergroup
	ug := server.Group("/users")
	ug.POST("/signup", u.SignUp)
	//ug.POST("/login", u.LogIn)
	ug.POST("/login", u.LogInJWT)
	ug.POST("/edit", u.Edit)
	//ug.GET("/profile", u.Profile)
	ug.GET("/profile", u.ProfileJWT)

}

// SignUp 注册
func (u *UserHandler) SignUp(c *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	var req SignUpReq
	//Bind方法根据Content-type来解析数据到req中 Bind接受请求
	//若解析错误则返回400错误
	if err := c.BindJSON(&req); err != nil {
		return
	}
	//邮箱和密码校验 正则表达式
	//预编译
	//emailExp := regexp.MustCompile(emailRegexPattern, regexp.None) 这样写需要每次都进行预编译 浪费内存和降低速度
	//邮箱校验
	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		c.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		c.String(http.StatusOK, "你的邮箱格式不对")
		return
	}
	//两次密码输入校验
	if req.ConfirmPassword != req.Password {
		c.String(http.StatusOK, "两次输入的密码不一致")
		return
	}
	//密码校验
	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		//需要记录日志
		c.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		c.String(http.StatusOK, "密码长度大于8位，要包含大小写字母和特殊字符")
		return
	}
	//调用svc方法
	err = u.svc.SingUp(c, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err == service.ErrUserDuplicateEmail {
		c.String(http.StatusOK, "邮箱冲突")
		return
	}
	if err != nil {
		c.String(http.StatusOK, "系统错误")
		return
	}
	c.String(http.StatusOK, "注册成功")

}

func (u *UserHandler) LogInJWT(c *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	//接受绑定参数 邮箱和密码 Bind方法 如果为空则返回
	if err := c.Bind(&req); err != nil {
		return
	}
	//调用service层方法 返回消息模型即邮箱和密码
	user, err := u.svc.Login(c, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		c.String(http.StatusOK, "用户名或密码错误")
		return
	}
	if err != nil {
		c.String(http.StatusOK, "系统错误")
		return
	}
	//设置JWT登录态
	//并生成JWT Token用于登录
	loginclaims := UserClaims{
		Uid:        user.Id,
		Authorized: true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)), // 设置过期时间 1 小时
		},
	} //将uid放入token中

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, loginclaims)
	//设置有效载荷
	//claims := token.Claims.(jwt.MapClaims)
	//claims["authorized"] = true
	//claims["exp"] = time.Now().Add(time.Hour * 1).Unix() // 1小时后过期
	tokenStr, err := token.SignedString([]byte("3f6e1f6f8c0e15a6c8ef634d0f6f4791e7b1f8f2d7d8a1e1d3f6b2e2c6d1c9e2f"))
	//if err != nil {
	//	c.String(http.StatusInternalServerError, "系统错误")
	//	return
	//}
	c.Header("x-jwt-token", tokenStr)
	fmt.Println(tokenStr)
	fmt.Println(user)
	c.String(http.StatusOK, "登陆成功")
	return
}

// LogIn 登录
func (u *UserHandler) LogIn(c *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	//接受绑定参数 邮箱和密码 Bind方法 如果为空则返回
	if err := c.Bind(&req); err != nil {
		return
	}
	//调用service层方法 返回消息模型即邮箱和密码
	user, err := u.svc.Login(c, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		c.String(http.StatusOK, "用户名或密码错误")
		return
	}
	if err != nil {
		c.String(http.StatusOK, "系统错误")
		return
	}
	//设置session
	sess := sessions.Default(c)
	//设置session的值
	sess.Set("userId", user.Id)
	//设置cookie的参数
	sess.Options(sessions.Options{
		//Secure: true, //use https connection
		//HttpOnly: true, and so on
	})
	sess.Save()
	c.String(http.StatusOK, "登陆成功")
	return
}

func (u *UserHandler) LogOut(c *gin.Context) {
	//设置session
	sess := sessions.Default(c)
	//设置cookie的参数
	sess.Options(sessions.Options{
		//Secure: true, //use https connection
		//HttpOnly: true, and so on
		MaxAge: -1,
	})
	sess.Save()

	c.String(http.StatusOK, "登出成功")
}

// Edit 编辑
func (u *UserHandler) Edit(c *gin.Context) {

}

// Profile 查看个人信息状态
func (u *UserHandler) Profile(c *gin.Context) {
	c.String(http.StatusOK, "这是你的Profile")
}

// ProfileJWT 查看个人信息状态
func (u *UserHandler) ProfileJWT(c *gin.Context) {
	ctx, ok := c.Get("claims")
	if !ok {
		//监控此处判断错误
		c.String(http.StatusOK, "系统错误")
		return
	}
	claims, ok := ctx.(*UserClaims) //断言
	if !ok {
		c.String(http.StatusOK, "系统错误")
		return
	}
	println(claims.Uid)
}

type UserClaims struct {
	jwt.RegisteredClaims
	//声明要放到token中的数据
	Uid        int64
	Authorized bool `json:"authorized"`
}
