package proj2

// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder, and we will be Very Upset.

import (

	// You neet to add with
	// go get github.com/nweaver/cs161-p2/userlib
	"github.com/nweaver/cs161-p2/userlib"

	// Life is much easier with json:  You are
	// going to want to use this so you can easily
	// turn complex structures into strings etc...
	"encoding/json"

	// Likewise useful for debugging etc
	"encoding/hex"

	// UUIDs are generated right based on the crypto RNG
	// so lets make life easier and use those too...
	//
	// You need to add with "go get github.com/google/uuid"
	"github.com/google/uuid"

	// For the useful little debug printing function
	"fmt"
	"os"
	"strings"
	"time"

	// I/O
	"io"

	// Want to import errors
	"errors"

	// These are imported for the structure definitions.  You MUST
	// not actually call the functions however!!!
	// You should ONLY call the cryptographic functions in the
	// userlib, as for testing we may add monitoring functions.
	// IF you call functions in here directly, YOU WILL LOSE POINTS
	// EVEN IF YOUR CODE IS CORRECT!!!!!
	"crypto/rsa"
)

// This serves two purposes: It shows you some useful primitives and
// it suppresses warnings for items not being imported
func someUsefulThings() {
	// Creates a random UUID
	f := uuid.New()
	debugMsg("UUID as string:%v", f.String())

	// Example of writing over a byte of f
	f[0] = 10
	debugMsg("UUID as string:%v", f.String())

	// takes a sequence of bytes and renders as hex
	h := hex.EncodeToString([]byte("fubar"))
	debugMsg("The hex: %v", h)

	// Marshals data into a JSON representation
	// Will actually work with go structures as well
	d, _ := json.Marshal(f)
	debugMsg("The json data: %v", string(d))
	var g uuid.UUID
	json.Unmarshal(d, &g)
	debugMsg("Unmashaled data %v", g.String())

	// debugMsg("marshal size %v, unmarshal size %v, original %v", len(d), len(g), len(f))

	// userdata := User{"a","b","c","d"}
	// mUD, _ := json.Marshal(userdata)
	// debugMsg("The json userdata: %v", string(mUD))
	// var us User
	// json.Unmarshal(mUD, &us)
	// debugMsg("Unmashaled userdata %v", us)
	// debugMsg("marshal size %v", len(mUD))

	// This creates an error type
	debugMsg("Creation of error %v", errors.New("This is an error"))

	// And a random RSA key.  In this case, ignoring the error
	// return value
	var key *rsa.PrivateKey
	key, _ = userlib.GenerateRSAKey()
	debugMsg("Key is %v", key)
}

// Helper function: Takes the first 16 bytes and
// converts it into the UUID type
func bytesToUUID(data []byte) (ret uuid.UUID) {
	for x := range ret {
		ret[x] = data[x]
	}
	return
}

// Helper function: Returns a byte slice of the specificed
// size filled with random data
func randomBytes(bytes int) (data []byte) {
	data = make([]byte, bytes)
	if _, err := io.ReadFull(userlib.Reader, data); err != nil {
		panic(err)
	}
	return
}

var DebugPrint = false

// Helper function: Does formatted printing to stderr if
// the DebugPrint global is set.  All our testing ignores stderr,
// so feel free to use this for any sort of testing you want
func debugMsg(format string, args ...interface{}) {
	if DebugPrint {
		msg := fmt.Sprintf("%v ", time.Now().Format("15:04:05.00000"))
		fmt.Fprintf(os.Stderr,
			msg+strings.Trim(format, "\r\n ")+"\n", args...)
	}
}

func CFBEncrypt(key []byte, data []byte) ([]byte) {
	// key is 16 bytes == BlockSize is 16 bytes == 128 bits
	ciphertext := make([]byte, userlib.BlockSize + len(data))
	iv := ciphertext[:userlib.BlockSize]

	// Load random data to iv
	if _, err := io.ReadFull(userlib.Reader, iv); err != nil {
		panic(err)
	}

	cipher := userlib.CFBEncrypter(key, iv)
	cipher.XORKeyStream(ciphertext[userlib.BlockSize:], data)

	return ciphertext
}

func CFBDecrypt(key []byte, ciphertext []byte) []byte {
	// key is 16 bytes == BlockSize is 16 bytes == 128 bits
	plaintext := make([]byte, len(ciphertext[userlib.BlockSize:]))
	iv := ciphertext[:userlib.BlockSize]

	cipher := userlib.CFBDecrypter(key, iv)
	cipher.XORKeyStream(plaintext, ciphertext[userlib.BlockSize:])

	return plaintext
}

func HMAC(key []byte, data []byte) []byte {
	hmac := userlib.NewHMAC(key)
	hmac.Write(data)
	return hmac.Sum(nil)
}

func VerifyHMAC(key []byte, data []byte, MAC []byte) bool {
	// Return true if correct, false otherwise
	hmac := userlib.NewHMAC(key)
	hmac.Write(data)
	expectedMAC := hmac.Sum(nil)
	return userlib.Equal(MAC, expectedMAC)
}

func Hash(dataToHash []byte) []byte {
	hasher := userlib.NewSHA256()
	hasher.Write(dataToHash)
	hash := hasher.Sum(nil)
	return hash
}

// The structure definition for a user record
type User struct {
	Username   string
	Password   string
	PrivateKey *rsa.PrivateKey
	PublicKey  string
	// You can add other fields here if you want...
	// Note for JSON to marshal/unmarshal, the fields need to
	// be public (start with a capital letter)
}

// This creates a user.  It will only be called once for a user
// (unless the keystore and datastore are cleared during testing purposes)

// It should store a copy of the userdata, suitably encrypted, in the
// datastore and should store the user's public key in the keystore.

// The datastore may corrupt or completely erase the stored
// information, but nobody outside should be able to get at the stored
// User data: the name used in the datastore should not be guessable
// without also knowing the password and username.

// You are not allowed to use any global storage other than the
// keystore and the datastore functions in the userlib library.

// You can assume the user has a STRONG password
func InitUser(username string, password string) (userdataptr *User, err error) {
	var userdata User
	userdata.Username = username
	userdata.Password = password
	userdata.PrivateKey, err = userlib.GenerateRSAKey()
	if err != nil {
		panic(err)
	}

	// TODO: Use API created above
	// TODO: Marshal any data structure that's not in string or []byte
	// TODO: Convert marshalled data and strings to []byte()

	// Old code below:
	// userdata.PublicKey = userdata.PrivateKey.PublicKey
	// userlib.KeystoreSet(username, userdata.PublicKey)
	// data := []byte(userdata)
	// ciphertext := make([]byte, userlib.BlockSize+len(data))

	// iv := ciphertext[:userlib.BlockSize]
	// if _, err := io.ReadFull(userlib.Reader, iv); err != nil {
	// 	panic(err)
	// }
	// symmetric_key := PBKDF2Key(
	// 	password,
	// 	[]byte("nosalt"), // TODO: change this
	// 	32,
	// )

	// cipher := CFBEncrypter(symmetric_key, iv)
	// cipher.XORKeyStream(ciphertext[userlib.BlockSize:], data)

	// h := userlib.NewSHA256() // TODO: Change this!! cannot hash username (low entropy)
	// h.Write(userdata.Username)
	// DatastoreSet(h.Sum(nil), ciphertext)
	return &userdata, err
}

// This fetches the user information from the Datastore.  It should
// fail with an error if the user/password is invalid, or if the user
// data was corrupted, or if the user can't be found.
func GetUser(username string, password string) (userdataptr *User, err error) {
	return
}

// This stores a file in the datastore.
//
// The name of the file should NOT be revealed to the datastore!
func (userdata *User) StoreFile(filename string, data []byte) {
}

// This adds on to an existing file.
//
// Append should be efficient, you shouldn't rewrite or reencrypt the
// existing file, but only whatever additional information and
// metadata you need.

func (userdata *User) AppendFile(filename string, data []byte) (err error) {
	return
}

// This loads a file from the Datastore.
//
// It should give an error if the file is corrupted in any way.
func (userdata *User) LoadFile(filename string) (data []byte, err error) {
	return
}

// You may want to define what you actually want to pass as a
// sharingRecord to serialized/deserialize in the data store.
type sharingRecord struct {
}

// This creates a sharing record, which is a key pointing to something
// in the datastore to share with the recipient.

// This enables the recipient to access the encrypted file as well
// for reading/appending.

// Note that neither the recipient NOR the datastore should gain any
// information about what the sender calls the file.  Only the
// recipient can access the sharing record, and only the recipient
// should be able to know the sender.

func (userdata *User) ShareFile(filename string, recipient string) (
	msgid string, err error) {
	return
}

// Note recipient's filename can be different from the sender's filename.
// The recipient should not be able to discover the sender's view on
// what the filename even is!  However, the recipient must ensure that
// it is authentically from the sender.
func (userdata *User) ReceiveFile(filename string, sender string,
	msgid string) error {
	return nil
}

// Removes access for all others.
func (userdata *User) RevokeFile(filename string) (err error) {
	return
}
