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

	// This creates an error type
	debugMsg("Creation of error %v", errors.New("This is an error"))

	// And a random RSA key.  In this case, ignoring the error
	// return value
	var key *rsa.PrivateKey
	key, _ = userlib.GenerateRSAKey()
	debugMsg("Key is %v", key)

	// userdata := User{"a","b", *key,"d"}
	// debugMsg("Unmarshaled userdata %v", userdata)
	// mUD, _ := json.Marshal(userdata)
	// debugMsg("The marshaled json userdata: %v", string(mUD))
	// var ud User
	// json.Unmarshal(mUD, &ud)
	// debugMsg("Unmarshaled userdata %v", ud)
	// mUD2, _ := json.Marshal(ud)
	// debugMsg("Marshal size %v", len(mUD))
	// debugMsg("Marshal size %v", len(mUD2))
	// debugMsg("Marshal equal %v", string(mUD) == string(mUD2))
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

func CFBEncrypt(key []byte, data []byte) []byte {
	// key is 16 bytes == BlockSize is 16 bytes == 128 bits
	ciphertext := make([]byte, userlib.BlockSize+len(data))
	iv := ciphertext[:userlib.BlockSize]

	// Load random data to iv
	if _, err := io.ReadFull(userlib.Reader, iv); err != nil {
		panic(err)
	}
	debugMsg("CFBEncrypt IV is: %v", iv)
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
	return userlib.Equal(MAC, HMAC(key, data))
}

func Hash(dataToHash []byte) []byte {
	hasher := userlib.NewSHA256()
	hasher.Write(dataToHash)
	hash := hasher.Sum(nil)
	return hash
}

type FileMetadata struct {
	FileID     uuid.UUID
	EncryptKey []byte
	MACKey     []byte
}

// The structure definition for a user record
type User struct {
	Username    string
	PrivateKey  rsa.PrivateKey
	PublicKey   rsa.PublicKey
	OwnedFiles  map[string]FileMetadata
	SharedFiles map[string]FileMetadata
	// You can add other fields here if you want...
	// Note for JSON to marshal/unmarshal, the fields need to
	// be public (start with a capital letter)
}

// Helper struct to hold pairs of E(data), MAC(E(data))
type EMAC struct {
	Ciphertext []byte
	Mac        []byte
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
	key, err := userlib.GenerateRSAKey()
	if err != nil {
		panic(err)
	}

	userdata.OwnedFiles = make(map[string]FileMetadata)
	userdata.SharedFiles = make(map[string]FileMetadata)

	userdata.PrivateKey = *key
	userdata.PublicKey = key.PublicKey

	masterKey := userlib.PBKDF2Key(
		[]byte(password),
		Hash([]byte(username)),
		userlib.AESKeySize*2,
	)

	macKey, encryptKey := masterKey[:userlib.AESKeySize], masterKey[userlib.AESKeySize:]
	path := "logins/" + string(HMAC(macKey, []byte(username)))
	userJSON, err := json.Marshal(userdata)
	if err != nil {
		panic(err)
	}

	ciphertext := CFBEncrypt(encryptKey, userJSON)
	emac := EMAC{
		Ciphertext: ciphertext,
		Mac:        HMAC(macKey, ciphertext),
	}
	emacJSON, err := json.Marshal(emac)
	if err != nil {
		panic(err)
	}

	userlib.KeystoreSet(username, userdata.PublicKey)
	userlib.DatastoreSet(path, emacJSON)
	return &userdata, err
}

// This fetches the user information from the Datastore.  It should
// fail with an error if the user/password is invalid, or if the user
// data was corrupted, or if the user can't be found.
func GetUser(username string, password string) (userdataptr *User, err error) {
	masterKey := userlib.PBKDF2Key(
		[]byte(password),
		Hash([]byte(username)),
		userlib.AESKeySize*2,
	)

	macKey, encryptKey := masterKey[:userlib.AESKeySize], masterKey[userlib.AESKeySize:]
	path := "logins/" + string(HMAC(macKey, []byte(username)))
	data, ok := userlib.DatastoreGet(path)

	// If this fails, did not find the file. Either
	// Username/password is bad or path tampered with
	if !ok {
		return nil, errors.New("Error finding the file.")
	}

	var emac EMAC
	err = json.Unmarshal(data, &emac)
	if err != nil {
		panic(err)
	}

	// check MAC of encrypted data to ensure no tampering
	if !VerifyHMAC(macKey, emac.Ciphertext, emac.Mac) {
		return nil, errors.New("User Data has been tampered with.")
	}

	var userdata User
	plaintext := CFBDecrypt(encryptKey, emac.Ciphertext)
	err = json.Unmarshal(plaintext, &userdata)
	if err != nil {
		panic(err)
	}

	return &userdata, err
}

type RevisionMetadata struct {
	FileSize      uint
	NumRevisions  uint
	RevisionSizes []uint
}

// Unencrypted metadata for file r, located at meta/r, where r is random
type SharedMetadata struct {
	Metadata    []byte // NOTE: Marshaled RevisionMetadata
	MetadataMAC []byte
}

// helper function to append byte array to byte array
func extend(a []byte, newData []byte) []byte {
	for _, b := range newData {
		a = append(a, b)
	}
	return a
}

// This stores a file in the datastore.
//
// The name of the file should NOT be revealed to the datastore!
// This will determine where the file metadata is stored in user struct
func (userdata *User) StoreFile(filename string, data []byte) {
	fileID := uuid.New() // random identifier
	fileEncryptKey := randomBytes(16)
	fileMacKey := randomBytes(16)
	fileMetadata := FileMetadata{
		FileID:     fileID,
		EncryptKey: fileEncryptKey,
		MACKey:     fileMacKey,
	}
	userdata.OwnedFiles[filename] = fileMetadata

	// initialize data revision metadata
	ciphertext := CFBEncrypt(fileEncryptKey, data)
	dataLen := uint(len(ciphertext))
	var revisionSizes []uint
	revisionSizes = append(revisionSizes, dataLen)

	revisionMetadata := RevisionMetadata{
		FileSize:      dataLen,
		NumRevisions:  1,
		RevisionSizes: revisionSizes,
	}
	debugMsg("StoreFile revisionMetadata is: %v", revisionMetadata)
	revisionJSON, err := json.Marshal(revisionMetadata)
	if err != nil {
		panic(err)
	}

	sharedMetadata := SharedMetadata{
		Metadata:    revisionJSON,
		MetadataMAC: HMAC(fileMacKey, revisionJSON),
	}

	sharedMetadataJSON, err := json.Marshal(sharedMetadata)
	if err != nil {
		panic(err)
	}
	metaDataPath := "meta/" + fileID.String()
	userlib.DatastoreSet(metaDataPath, sharedMetadataJSON)

	debugMsg("StoreFile cipher is: %v", ciphertext)
	fileMAC := HMAC(fileMacKey, ciphertext)
	debugMsg("StoreFile MAC is: %v", fileMAC)
	// ciphertext size == len(IV) + dataLen; len(IV) == userlib.Blocksize
	// MAC size == 32 bytes
	// fileData := make([]byte, 0, userlib.HashSize+len(ciphertext)) // enough to hold MAC if file size is small
	var fileData []byte
	fileData = extend(fileData, fileMAC)
	fileData = extend(fileData, ciphertext)

	filePath := "file/" + fileID.String()
	debugMsg("StoreFile filepath is: %v", filePath)
	userlib.DatastoreSet(filePath, fileData)
}

// This adds on to an existing file.
//
// Append should be efficient, you shouldn't rewrite or reencrypt the
// existing file, but only whatever additional information and
// metadata you need.

func (userdata *User) AppendFile(filename string, data []byte) (err error) {
	fileMetaData, ok := userdata.OwnedFiles[filename]
	if !ok {
		debugMsg("AppendFile -- Filename: %s", filename)
		return errors.New("File not found, please check filename")
	}

	ciphertext := CFBEncrypt(fileMetaData.EncryptKey, data)

	filePath := "file/" + fileMetaData.FileID.String()
	metadataPath := "meta/" + fileMetaData.FileID.String()

	// fetch and verify file metadata
	sharedMetadataJSON, ok := userlib.DatastoreGet(metadataPath)
	if !ok {
		return errors.New("Metadata not in datastore, may have been moved")
	}
	var sharedMetadata SharedMetadata
	err = json.Unmarshal(sharedMetadataJSON, &sharedMetadata)
	if err != nil {
		panic(err)
	}
	if !VerifyHMAC(
		fileMetaData.MACKey,
		sharedMetadata.Metadata,
		sharedMetadata.MetadataMAC,
	) {
		return errors.New("File metadata has been tampered with")
	}
	var revisionMetadata RevisionMetadata
	err = json.Unmarshal(sharedMetadata.Metadata, &revisionMetadata)
	if err != nil {
		panic(err)
	}

	// update metadata
	dataLen := uint(len(ciphertext))
	revisionMetadata.FileSize += dataLen
	revisionMetadata.NumRevisions++
	revisionMetadata.RevisionSizes = append(
		revisionMetadata.RevisionSizes,
		dataLen,
	)
	newRevisionMetadataJSON, err := json.Marshal(revisionMetadata)
	if err != nil {
		panic(err)
	}
	sharedMetadata.Metadata = newRevisionMetadataJSON
	sharedMetadata.MetadataMAC = HMAC(fileMetaData.MACKey, newRevisionMetadataJSON)
	newSharedMetadataJSON, err := json.Marshal(sharedMetadata)
	if err != nil {
		panic(err)
	}
	userlib.DatastoreSet(metadataPath, newSharedMetadataJSON)

	// append to file
	file, ok := userlib.DatastoreGet(filePath)
	if !ok {
		return errors.New("File not in datastore, may have been moved")
	}
	mac := HMAC(fileMetaData.MACKey, ciphertext)
	file = extend(file, mac)
	file = extend(file, ciphertext)
	userlib.DatastoreSet(filePath, file)

	return err
}

// This loads a file from the Datastore.
//
// It should give an error if the file is corrupted in any way.
func (userdata *User) LoadFile(filename string) (data []byte, err error) {
	fileMetaData, ok := userdata.OwnedFiles[filename]
	if !ok { // could be shared instead of owned
		fileMetaData, ok = userdata.SharedFiles[filename]
		if !ok {
			return nil, errors.New("File not found, please check filename")
		}
	}

	filePath := "file/" + fileMetaData.FileID.String()
	metadataPath := "meta/" + fileMetaData.FileID.String()

	debugMsg("LoadFile filepath is: %v", filePath)

	// fetch and verify file metadata
	sharedMetadataJSON, ok := userlib.DatastoreGet(metadataPath)
	if !ok {
		return nil, errors.New("Metadata not in datastore, may have been moved")
	}
	var sharedMetadata SharedMetadata
	err = json.Unmarshal(sharedMetadataJSON, &sharedMetadata)
	if err != nil {
		panic(err)
	}
	if !VerifyHMAC(
		fileMetaData.MACKey,
		sharedMetadata.Metadata,
		sharedMetadata.MetadataMAC,
	) {
		return nil, errors.New("File metadata has been tampered with")
	}
	var revisionMetadata RevisionMetadata
	err = json.Unmarshal(sharedMetadata.Metadata, &revisionMetadata)
	if err != nil {
		panic(err)
	}
	debugMsg("LoadFile revisionMetadata is: %v", revisionMetadata)
	// verify, decrypt, and copy file data
	file, ok := userlib.DatastoreGet(filePath)
	if !ok {
		return nil, errors.New("File not in datastore, may have been moved")
	}

	var fileData []byte
	j := 0
	// TODO: are these int conversion safe?
	for i := 0; i < int(revisionMetadata.NumRevisions); i++ {
		offset := j + userlib.HashSize
		mac := file[j:offset]

		ciphertext := file[offset : offset+int(revisionMetadata.RevisionSizes[i])]
		debugMsg("LoadFile MAC is: %v", mac)
		debugMsg("LoadFile cipher is: %v", ciphertext)
		// check MAC of encrypted data to ensure no tampering
		if !VerifyHMAC(fileMetaData.MACKey, ciphertext, mac) {
			return nil, errors.New("File Data has been tampered with.")
		}
		plaintext := CFBDecrypt(fileMetaData.EncryptKey, ciphertext)
		fileData = extend(fileData, plaintext)
		j += userlib.HashSize + int(revisionMetadata.RevisionSizes[i])
	}

	return fileData, err
}

// You may want to define what you actually want to pass as a
// SharingRecord to serialized/deserialize in the data store.
type SharingRecord struct {
	PrivateFileMetadata []byte // Marshaled FileMetaData
	Signature           []byte
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
	recipientKey, ok := userlib.KeystoreGet(recipient)
	if !ok {
		return "", errors.New("Recipient public key not found. Check recipient name.")
	}

	// retrieve private file metadata
	fileMetadata, err := json.Marshal(userdata.OwnedFiles[filename])
	if err != nil {
		panic(err)
	}

	// encrypt and sign
	ciphertext, err := userlib.RSAEncrypt(&recipientKey, fileMetadata, nil)
	if err != nil {
		panic(err)
	}
	rsaSignature, err := userlib.RSASign(&userdata.PrivateKey, ciphertext)
	if err != nil {
		panic(err)
	}

	sharedRecord := SharingRecord{
		PrivateFileMetadata: ciphertext,
		Signature:           rsaSignature,
	}
	message, err := json.Marshal(sharedRecord)
	if err != nil {
		panic(err)
	}
	debugMsg("ShareFie -- ciphertext: %s", ciphertext)
	debugMsg("ShareFie -- Shared Record: %s", sharedRecord)
	debugMsg("ShareFie -- Shared Record JSON: %s", message)

	return string(message), err
}

// Note recipient's filename can be different from the sender's filename.
// The recipient should not be able to discover the sender's view on
// what the filename even is!  However, the recipient must ensure that
// it is authentically from the sender.
func (userdata *User) ReceiveFile(filename string, sender string,
	msgid string) error {
	sharedRecordJSON := []byte(msgid)

	var sharedRecord SharingRecord
	err := json.Unmarshal(sharedRecordJSON, &sharedRecord)
	if err != nil {
		panic(err)
	}

	debugMsg("ReceiveFile -- Shared Record: %s", sharedRecord)
	debugMsg("ReceiveFile -- Shared Record JSON: %s", sharedRecordJSON)

	// verify message signature and decrypt
	senderKey, ok := userlib.KeystoreGet(sender)
	debugMsg("Sender: %s", sender)
	debugMsg("Recipient: %s", userdata.Username)
	if !ok {
		return errors.New("Sender public key not found. Check sender name.")
	}
	if userlib.RSAVerify(
		&senderKey,
		sharedRecord.PrivateFileMetadata,
		sharedRecord.Signature,
	) != nil {
		debugMsg("File Metadata is: %s", string(sharedRecord.PrivateFileMetadata))
		return errors.New("Message verification failed.")
	}

	message, err := userlib.RSADecrypt(
		&userdata.PrivateKey,
		sharedRecord.PrivateFileMetadata,
		nil,
	)

	var fileMetadata FileMetadata
	json.Unmarshal(message, &fileMetadata)
	userdata.SharedFiles[filename] = fileMetadata
	return err
}

// Removes access for all others.
func (userdata *User) RevokeFile(filename string) (err error) {
	// Check if file belongs to user
	fileMetaData, ok := userdata.OwnedFiles[filename]
	if !ok { // could be shared instead of owned
		return errors.New("Revoke File not found or you are not the original owner")
	}

	originalFileData, err := userdata.LoadFile(filename)
	if err != nil {
		return err
	}

	// First we re-encrypt current location with random keys
	// Can also load it with random data
	oldFileID := fileMetaData.FileID
	randomEncryptKey := randomBytes(16)
	randomMACKey := randomBytes(16)

	oldFilePath := "file/" + oldFileID.String()
	oldMetadataPath := "meta/" + oldFileID.String()

	// encrypt and mac metadata with random keys =====================
	randomCiphertext := CFBEncrypt(randomEncryptKey, originalFileData)

	dataLen := uint(len(randomCiphertext))

	revisionMetadata := RevisionMetadata{
		FileSize:      dataLen,
		NumRevisions:  1,
		RevisionSizes: []uint{dataLen},
	}

	revisionJSON, err := json.Marshal(revisionMetadata)
	if err != nil {
		panic(err)
	}

	sharedMetadata := SharedMetadata{
		Metadata:    revisionJSON,
		MetadataMAC: HMAC(randomMACKey, revisionJSON),
	}

	sharedMetadataJSON, err := json.Marshal(sharedMetadata)
	if err != nil {
		panic(err)
	}
	// Done: encrypt and mac metadata with random keys =====================

	// Revoke access by to everyone by storing random contents =====================
	userlib.DatastoreSet(oldMetadataPath, sharedMetadataJSON)
	var randomData []byte
	mac := HMAC(randomMACKey, randomCiphertext)
	randomData = extend(randomData, mac)
	randomData = extend(randomData, randomCiphertext)
	userlib.DatastoreSet(oldFilePath, randomData)
	// Done: Revoke access by to everyone by storing random contents =====================

	userlib.DatastoreDelete(oldFilePath)

	// Copy file to new location. StoreFile generates new location and new keys
	userdata.StoreFile(filename, originalFileData)
	return err
}
