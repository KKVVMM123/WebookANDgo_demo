package middleware

import (
	"encoding/gob"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go_demo/webook/internal/web"
	"log"
	"net/http"
	"strings"
	"time"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

// CheckLoginJWT 登录校验
func (l *LoginJWTMiddlewareBuilder) CheckLoginJWT() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		//sess := sessions.Default(ctx)
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}
		//JWT校验
		tokenHeader := ctx.GetHeader("Authorization")
		if tokenHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
			})
			return
		}
		segs := strings.Split(tokenHeader, " ")
		if len(segs) != 2 {
			//没有登录
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
			})
			return
		}
		tokenStr := segs[1]
		claims := &web.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("3f6e1f6f8c0e15a6c8ef634d0f6f4791e7b1f8f2d7d8a1e1d3f6b2e2c6d1c9e2f"), nil
		})
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
			})
			return
		}
		if token == nil || !token.Valid || claims.Uid == 0 {
			//没有登录
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
			})
			return
		}
		now := time.Now()
		if claims.ExpiresAt.Sub(now) < time.Second*50 {
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute))
			tokenStr, err = token.SignedString([]byte("3f6e1f6f8c0e15a6c8ef634d0f6f4791e7b1f8f2d7d8a1e1d3f6b2e2c6d1c9e2f"))
			if err != nil {
				log.Println("jwt续约失败", err)
			}
			ctx.Header("x-jwt-token", tokenStr)
		}

		ctx.Set("claims", claims)
		//ctx.Set("userId", claims.Uid)
	}
}
