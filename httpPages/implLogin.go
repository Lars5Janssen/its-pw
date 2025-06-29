package pages

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Lars5Janssen/its-pw/internal/repository"
	"github.com/Lars5Janssen/its-pw/util"
	"github.com/google/uuid"
)

var (
	sharedSecret     string
	sessionKeyLength int
)

type message struct {
	Sid     string `json:"sid"`
	Message string `json:"ecryptedData"`
}

func MSG(w http.ResponseWriter, r *http.Request) {
	l.Println("MSG RECIVED")

	// get msg json
	var msg message
	err := json.NewDecoder(r.Body).Decode(&msg)
	util.EasyCheck(err, "ERROR during MSG: json decode: ", "err")

	// validate sid
	repo := repository.New(conn)
	sessionKey, err := repo.GetSessionKeyBySID(ctx, msg.Sid)
	if err != nil {
		l.Println("SID not in sessions table, refusing")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	l.Println("SID found")

	// decrypt msg
	decryptedMSG := decrypt(
		string(fromBase64(msg.Message)), 
		string(sessionKey))

	// construct response
	responseMSG := "Your message was:\""+decryptedMSG+"\"."
	encrypted := encrypt(responseMSG, string(sessionKey))
	
	resp := message{
		Sid: msg.Sid,
		Message: toBase64(encrypted),
	}

	body, err := json.Marshal(resp)
	util.EasyCheck(err, "ERROR during response impl: ", "err")

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, string(body))

}

func SendLogin(w http.ResponseWriter, r *http.Request) {
	l.Println("SEND LOGIN")
	sharedSecret = "passphrasewhichneedstobe32bytes!"
	sessionKeyLength = 32

	// Get User Input
	data, err := getSendLoginData(r)
	util.EasyCheck(err, "ERROR in SendLogin during getSendLoginData: ", "err")
	decoded := fromBase64(data.EncryptedData)
	cNounce := decrypt(string(decoded), sharedSecret)

	// Generate Own Response
	ownNounce := strconv.FormatInt(time.Now().UnixNano(), 10)
	sid := uuid.NewString()
	sessionKey, err := generateSessionKey(32)
	util.EasyCheck(err, "ERROR generation Session Key: ", "err")

	repo := repository.New(conn)
	repo.CreateImplSession(ctx, repository.CreateImplSessionParams{
		Sid:          sid,
		Username:     &data.Username,
		ClientNounce: cNounce,
		OwnNounce:    ownNounce,
		SessionKey:   sessionKey,
	})
	sResponse := serverResponse{
		ClientNounce: cNounce,
		ServerNounce: ownNounce,
		Sid:          sid,
		Sessionkey:   sessionKey,
	}
	l.Println("ServerRsp: ", sResponse)

	respData, err := json.Marshal(sResponse)
	util.EasyCheck(err, "ERROR during marshaling of resp in sendLogin: ", "err")
	encRespToClient := encrypt(string(respData), sharedSecret)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, toBase64(encRespToClient))
}

func generateSessionKey(length int) ([]byte, error) {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func toBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

func fromBase64(input string) []byte {
	decoded, err := base64.StdEncoding.DecodeString(input)
	util.EasyCheck(err, "ERROR in Decoding String: fromBase64: ", "err")
	return decoded
}

type serverResponse struct {
	ClientNounce string `json:"clientNounce"`
	ServerNounce string `json:"serverNounce"`
	Sid          string `json:"sid"`
	Sessionkey   []byte `json:"sessionkey"`
}

func encrypt(input string, secret string) string {
	key := []byte(secret)
	plaintext := []byte(input)

	// PKCS Padding
	padding := aes.BlockSize - len(plaintext)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	plaintext = append(plaintext, padtext...)

	block, err := aes.NewCipher(key)
	util.EasyCheck(err, "ERROR during new cipher encryption: ", "err")

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		util.EasyCheck(err, "ERROR during IV generation: ", "err")
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return hex.EncodeToString(ciphertext)
}

func decrypt(input string, secret string) string {
	key := []byte(secret)
	ciphertext, err := hex.DecodeString(input)
	util.EasyCheck(err, "Decrypt: hex decode failed: ", "err")

	block, err := aes.NewCipher(key)
	util.EasyCheck(err, "Decrypt: NewCipher failed: ", "err")

	if len(ciphertext) < aes.BlockSize {
		l.Fatal("Decrypt: ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	if len(ciphertext)%aes.BlockSize != 0 {
		l.Fatal("Decrypt: ciphertext not a multiple of block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// Remove padding
	padding := int(ciphertext[len(ciphertext)-1])
	return string(ciphertext[:len(ciphertext)-padding])
}

type implSendLoginData struct {
	Username      string `json:"username"`
	EncryptedData string `json:"encryptedData"`
}

func getSendLoginData(r *http.Request) (implSendLoginData, error) {
	var data implSendLoginData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return implSendLoginData{}, err
	}
	return data, nil
}

type implAwnserLoginData struct {
	EncryptedData string `json:"encryptedData"`
}

type implMessageData struct {
	Sid           string `json:"sid"`
	EncryptedData string `json:"encryptedData"`
}
