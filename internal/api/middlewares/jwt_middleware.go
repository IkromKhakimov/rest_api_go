package middlewares

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"restapi/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
)

type ContextKey string

func JWTMiddleware(next http.Handler) http.Handler {
	fmt.Println("---------------- JWT Middleware ------------------")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("++++++++++++++++++++ Inside JWT Middleware")

		token, err := r.Cookie("Bearer")
		if err != nil {
			http.Error(w, "Authorization Header Missing", http.StatusUnauthorized)
			return
		}
		fmt.Println(token.Value)

		jwtSecret := os.Getenv("JWT_SECRET")

		// Parse takes the token string and a function for looking up the key. The latter is especially
		// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
		// head of the token to identify which key to use, but the parsed token (head and claims) is provided
		// to the callback, providing flexibility.
		parsedToken, err := jwt.Parse(token.Value, func(token *jwt.Token) (any, error) {
			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return []byte(jwtSecret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				http.Error(w, "Token Expired", http.StatusUnauthorized)
				return
			} else if errors.Is(err, jwt.ErrTokenMalformed) {
				http.Error(w, "Token Expired", http.StatusUnauthorized)
				return
			}
			utils.ErrorHandler(err, "Token not Valid")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if parsedToken.Valid {
			log.Println("Valid JWT")
		} else {
			http.Error(w, "Invalid Login Token", http.StatusUnauthorized)
			log.Println("Invalid JWT:", token.Value)
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if ok {
			fmt.Println(claims["uid"], claims["exp"], claims["role"])
		} else {
			//fmt.Println(err)
			http.Error(w, "Invalid Login Token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ContextKey("role"), claims["role"])
		ctx = context.WithValue(ctx, ContextKey("role"), claims["exp"])
		ctx = context.WithValue(ctx, ContextKey("role"), claims["user"])
		ctx = context.WithValue(ctx, ContextKey("role"), claims["uid"])

		fmt.Println(ctx)

		//fmt.Println(r.Cookie("Bearer"))
		next.ServeHTTP(w, r.WithContext(ctx))
		fmt.Println("Sent Response from JWT Middleware")
	})
}
