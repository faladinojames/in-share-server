package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/scrypt"
	"log"
	"os"
)


type Crypto struct {

}


func (crypto *Crypto) EncryptPassword(password string)string{
	return crypto.textToScrypt(password)
}

func (crypto *Crypto) CheckPassword(passwordText string, encryptedPassword string)bool{
	encryptedPasswordToCompare := crypto.textToScrypt(passwordText)
	return encryptedPasswordToCompare == encryptedPassword
}

func (crypto *Crypto) textToScrypt(text string) string {
	var salt = []byte(os.Getenv("JWT_TOKEN"))


	dk, err := scrypt.Key([]byte(text), salt, 1<<15, 8, 1, 32)
	if err != nil {
		log.Fatal(err)
	}

	return base64.StdEncoding.EncodeToString(dk)
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}


func (crypto *Crypto) GenerateJWT(claims jwt.MapClaims) (string, error){
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string using the secret

	a, err :=token.SignedString([]byte(os.Getenv("JWT_TOKEN")))

	return a, err
}
