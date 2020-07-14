package token

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"strconv"
	"time"
)

var secretKey []byte = []byte("hwT35e0X$lb&oly%")

func Generate(uid string) (string, error) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": uid,
		"exp": strconv.FormatInt(time.Now().Unix()+86400, 10),
	})
	return jwtToken.SignedString(secretKey)
}

func Verify(uid, tokenStr string) error {
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// secretKey is a []byte containing your secret, e.g. []byte("my_secret_key")
		return secretKey, nil
	})

	if err != nil {
		return err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		uidParam, ok := claims["uid"]
		if !ok {
			return fmt.Errorf("uid not in token, uid=%v", uid)
		}

		uidStr, ok := uidParam.(string)
		if !ok {
			return fmt.Errorf("uid in token is not string, uid=%v", uid)
		}

		expParam, ok := claims["exp"]
		if !ok {
			return fmt.Errorf("exp not in token, uid=%v", uid)
		}

		expStr, ok := expParam.(string)
		if !ok {
			return fmt.Errorf("exp in token is not string, uid=%v", uid)
		}

		var expInt int64
		expInt, err = strconv.ParseInt(expStr, 10, 64)
		if err != nil {
			return err
		}

		if uidStr == uid {
			if time.Now().Unix() > expInt {
				//token已过期
				return fmt.Errorf("token expire, uid=%v", uid)
			} else {
				//验证成功
				return nil
			}
		}
	}
	return fmt.Errorf("token error, uid=%v", uid)
}
