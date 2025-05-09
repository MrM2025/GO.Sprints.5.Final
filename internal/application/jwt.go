package application

import (
	"time"
	"fmt"
	"log"

	"github.com/golang-jwt/jwt/v5"
)

const hmacSampleSecret = "really_secret_signature"

func AddJWT(u string) string {
	now := time.Now() 
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": u,
		"nbf":  now.Add(time.Second).Unix(),
		"exp":  now.Add(5 * time.Minute).Unix(),
		"iat":  now.Unix(),
	})

	tokenString, err := token.SignedString([]byte(hmacSampleSecret))
	if err != nil {
		panic(err)
	}

	return tokenString
}

func strimJWT(u, t string) {
	tokenFromString, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(hmacSampleSecret), nil
	})

	if err != nil {
		log.Fatal(err)
	}

	if claims, ok := tokenFromString.Claims.(jwt.MapClaims); ok {
		fmt.Println(u, claims["name"])
	} else {
		fmt.Errorf("%s", err)
	}
}