package middleware

import (
	"awesomeProject/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth 校验请求中的 JWT（支持 Authorization Header 或 Cookie），有效后在 Context 中设置 userEmail
func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string
		// 优先从 Authorization Header 获取
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			tokenString = strings.TrimPrefix(auth, "Bearer ")
		} else {
			// 再尝试从 Cookie 获取
			t, err := c.Cookie("token")
			if err != nil || t == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "请输入您的身份"})
				return
			}
			tokenString = t
		}

		// 解析并验证 JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "您的身份无效"})
			return
		}
		// 从 Claims 提取 email
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "身份信息解析失败"})
			return
		}
		email, _ := claims["email"].(string)
		c.Set("userEmail", email)
		c.Next()
	}
}
