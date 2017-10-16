package proj2

import "testing"
import "github.com/nweaver/cs161-p2/userlib"
import "errors"

// You can actually import other stuff if you want IN YOUR TEST
// HARNESS ONLY.  Note that this is NOT considered part of your
// solution, but is how you make sure your solution is correct.

func corruptData(path string) (err error) {
	data, ok := userlib.DatastoreGet(path)
	if !ok {
		return errors.New("Failed to corrupt data. Could not get data from datastore")
	}
	if len(data) <= 0 {
		return errors.New("Failed to corrupt data. Data is empty")
	}
	bite := data[2]
	data[2] = byte(int(bite) - 1)
	userlib.DatastoreSet(path, data)
	return err
}

func TestInit(t *testing.T) {
	t.Log("Initialization test")
	DebugPrint = false
	someUsefulThings()

	DebugPrint = false
	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
	}
	// t.Log() only produces output if you run with "go test -v"
	t.Log("Got user", u)

	if u.Username == "" {
		t.Error("Username not initialized")
	} else if u.Password == "" {
		t.Error("Password not initialized")
	} else if u.OwnedFiles == nil {
		t.Error("OwnedFiles not initialized")
	} else if u.SharedFiles == nil {
		t.Error("SharedFiles not initialized")
	}
}

func TestStorage(t *testing.T) {
	DebugPrint = true
	fillDataStore(t)

	v, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to reload user", err)
		return
	}
	t.Log("Loaded user", v)

	v, err = GetUser("alice", "barfu")
	if err == nil {
		t.Error("Invalid login credentials passed when it should have failed", err)
		return
	}

	v, err = GetUser("bob", "fubar")
	if err == nil {
		t.Error("Invalid login credentials passed when it should have failed", err)
		return
	}

	// test corrupted data
	masterKey := userlib.PBKDF2Key(
		[]byte("fubar"),
		Hash([]byte("alice")),
		userlib.AESKeySize*2,
	)

	macKey := masterKey[:userlib.AESKeySize]
	path := "logins/" + bytesToUUID(HMAC(macKey, []byte("alice"))).String()
	corruptData(path)

	v, err = GetUser("alice", "fubar")
	if err == nil {
		t.Error("Corrupted userdata passed when it should have failed:", err)
		return
	}
	userlib.DatastoreClear()
}

func TestTamperedLoadMetadata(t *testing.T) {
	DebugPrint = true
	fillDataStore(t)

	user, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed GetUser:", err)
		return
	}

	user.StoreFile("meta", []byte("Don't mess with my metadata"))

	privateMetadata, err := user.getPrivateMetadata("meta")
	if err != nil {
		t.Error("Failed to retrieve private metadata:", err)
		return
	}

	revisionMetadata, err := user.loadMetadata("meta")
	if err != nil || revisionMetadata == nil {
		t.Error("Failed to retrieve revision metadata:", err)
		return
	}

	metaDataPath := "meta/" + privateMetadata.FileID.String()
	corruptData(metaDataPath)

	revisionMetadata, err = user.loadMetadata("meta")
	if err == nil || revisionMetadata != nil {
		t.Error("Expected corrupted metadata to fail, but passed")
		return
	}

	userlib.DatastoreClear()
}

func TestTamperedLoadFile(t *testing.T) {
	DebugPrint = false
	fillDataStore(t)

	user, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed GetUser:", err)
		return
	}

	user.StoreFile("meta", []byte("Don't mess with my metadata"))

	privateMetadata, err := user.getPrivateMetadata("meta")
	if err != nil {
		t.Error("Failed to retrieve private metadata:", err)
		return
	}

	// successful file load
	file, err := user.LoadFile("meta")
	if err != nil || file == nil {
		t.Error("Failed to load file:", err)
		return
	}

	// non-existent file load
	file, err = user.LoadFile("beta")
	if err != nil {
		t.Error("Error should not be thrown on file not found", err)
		return
	} else if file != nil {
		t.Error("LoadFile on non-existent file returned non-nil value")
		return
	}

	// corrupted data file load
	filepath := "file/" + privateMetadata.FileID.String()
	err = corruptData(filepath)
	if err != nil {
		t.Error("data corruption failed:", err)
	}
	file, err = user.LoadFile("meta")
	if err == nil || file != nil {
		t.Error("Expected error and nil file return for corrupted file")
		return
	}

	userlib.DatastoreClear()
}

func TestLenCap(t *testing.T) {
	fillDataStore(t)
	user, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed GetUser:", err)
		return
	}
	// t.Log("Loaded user:", user)

	array := make([]byte, 10, 10)
	t.Log("Original file len:", len(array))
	t.Log("Original file cap:", cap(array))

	DebugPrint = false
	user.StoreFile("LenCap", array)

	var file []byte
	file, err = user.LoadFile("LenCap")
	if err != nil || file == nil {
		t.Error("Failed LoadFile")
		return
	}

	if !isEqualByteArrays(array, file) {
		t.Error("Load failure, wrong contents")
	}

	userlib.DatastoreClear()
	file, err = user.LoadFile("LenCap")
	if err != nil {
		// Expecting there to be an error, but there was no error
		t.Error("File not found should not return an error")
		return
	} else if file != nil {
		t.Error("Expected LoadFile on non-existent file to return nil, but did not.")
	}
	userlib.DatastoreClear()
}

func TestUserCollisions(t *testing.T) {
	fillDataStore(t)
	u, err := GetUser("bob", "abc123")
	if err != nil {
		t.Error("Failed to reload user", err)
		return
	}
	userA := u
	t.Log("Loaded user", userA)
	u, err = GetUser("bo", "babc123")
	if err != nil {
		t.Error("Failed to reload user", err)
		return
	}
	userB := u
	t.Log("Loaded user", userB)
	u, err = GetUser("boba", "bc123")
	if err != nil {
		t.Error("Failed to reload user", err)
		return
	}
	userC := u
	t.Log("Loaded user", userC)
	_, err = GetUser("bo", "bc123")
	if err == nil {
		// Expecting error, but there was no error
		t.Error("Failed GetUser expecting an error, but got no error")
		return
	}
	if (userA.Username == userB.Username) ||
		(userA.Username == userC.Username) || (userB.Username == userC.Username) {
		t.Error("User data login caused collision")
		return
	}
	if (userA.PublicKey == userB.PublicKey) ||
		(userA.PublicKey == userC.PublicKey) || (userB.PublicKey == userC.PublicKey) {
		t.Error("User data login caused collision")
		return
	}
}

func TestSameUser(t *testing.T) {
	DebugPrint = false
	fillDataStore(t)
	// Having previously created a user "alice" with password "fubar"...
	alice, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to reload user Alice", err)
		return
	}
	also_alice, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to reload second user Alice", err)
		return
	}

	alice.StoreFile("todo", []byte("write tests"))
	todo, err := also_alice.LoadFile("todo")
	if err != nil || todo == nil {
		t.Error("Failed to load file.", err)
		return
	}
	if string(todo) != "write tests" {
		t.Error("Same user and password could not access file: ", err)
	}

	_, ok := also_alice.OwnedFiles["todo"]
	if !ok {
		t.Error("Failed to load metadata. OwnedFiles dictionaries not consistent")
	}
}

func TestSharedAppendAndRevoke(t *testing.T) {
	DebugPrint = false
	fillDataStore(t)
	alice, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to reload user Alice", err)
		return
	}
	bob, err := GetUser("bob", "abc123")
	if err != nil {
		t.Error("Failed to reload user Bob", err)
		return
	}
	bo, err := GetUser("bo", "babc123")
	if err != nil {
		t.Error("Failed to reload user bob", err)
		return
	}

	verse1 := []byte("Simple Simon ")
	alice.StoreFile("rhyme", verse1)

	var file []byte
	file, err = alice.LoadFile("rhyme")
	if err != nil || file == nil {
		t.Error("Failed LoadFile:", err)
		return
	}

	if !isEqualByteArrays(verse1, file) {
		t.Error("Load failure, wrong contents")
	}

	verse2 := []byte("met a pie-man ")
	err = alice.AppendFile("rhyme", verse2)
	if err != nil {
		t.Error("Failed Alice AppendFile:", err)
		return
	}

	verses1And2 := []byte("Simple Simon met a pie-man ")
	file, err = alice.LoadFile("rhyme")
	if err != nil || file == nil {
		t.Error("Failed LoadFile:", err)
		return
	}

	t.Logf("Array A: %s, Array B: %s", verses1And2, file)
	if !isEqualByteArrays(verses1And2, file) {
		t.Error("Append failure. Arrays not equal")
	}

	msgid, err := alice.ShareFile("rhyme", "bob")
	if err != nil {
		t.Error("Failed ShareFile:", err)
		return
	}
	err = bob.ReceiveFile("flow", "alice", msgid)
	if err != nil {
		t.Error("Failed ReceiveFile:", err)
		return
	}

	// does not have access to file under other users' filename
	msgid2, err := bob.ShareFile("rhyme", "bo")
	if err == nil || msgid2 != "" {
		t.Error("User should not have access to file under other user's local filename")
		return
	}

	msgid2, err = bob.ShareFile("flow", "bo")
	if err != nil || msgid2 == "" {
		t.Error("Failed ShareFile:", err)
		return
	}
	err = bo.ReceiveFile("goo", "bob", msgid2)
	if err != nil {
		t.Error("Failed ReceiveFile:", err)
		return
	}

	verse3 := []byte("going to the fair.")
	err = bob.AppendFile("flow", verse3)
	if err != nil {
		t.Error("Failed Bob AppendFile:", err)
		return
	}

	verse4 := []byte(" The end.")
	err = alice.AppendFile("rhyme", verse4)
	if err != nil {
		t.Error("Failed Alice AppendFile:", err)
		return
	}

	// ==================================================
	finalVerse := []byte("Simple Simon met a pie-man going to the fair. The end.")
	// ==================================================

	file, err = alice.LoadFile("rhyme")
	if err != nil || file == nil {
		t.Error("Failed LoadFile:", err)
		return
	}

	if !isEqualByteArrays(finalVerse, file) {
		t.Error("Shared Append failure. Edits by Bob not reflected in Alice's file.")
	}

	file, err = bo.LoadFile("goo")
	if err != nil || file == nil {
		t.Error("Failed LoadFile:", err)
		return
	}

	if !isEqualByteArrays(finalVerse, file) {
		t.Error("Shared Append failure. Edits by Bob not reflected in Bo's file.")
	}

	err = bo.RevokeFile("goo")
	if err == nil {
		t.Error("Expected not original owner error, but got no error")
	}

	err = bob.RevokeFile("flow")
	if err == nil {
		t.Error("Expected not original owner error, but got no error")
	}

	err = alice.RevokeFile("rhyme")
	if err != nil {
		t.Error("Alice revoke file failed: ", err)
	}

	file, err = alice.LoadFile("rhyme")
	if err != nil || file == nil {
		t.Error("Failed LoadFile:", err)
		return
	}
	if !isEqualByteArrays(finalVerse, file) {
		t.Error("File contents changed after revoke")
	}

	file, err = bob.LoadFile("flow")
	if err == nil || file != nil {
		t.Error("Bob Expected load file failure")
		return
	}

	file, err = bo.LoadFile("goo")
	if err == nil || file != nil {
		t.Error("Bo Expected load file failure")
		return
	}

	msgid, err = alice.ShareFile("rhyme", "bo")
	if err != nil {
		t.Error("Failed ShareFile:", err)
		return
	}
	err = bo.ReceiveFile("goo2", "alice", msgid)
	if err != nil {
		t.Error("Failed ReceiveFile:", err)
		return
	}

	file, err = bo.LoadFile("goo2")
	if err != nil || file == nil {
		t.Error("Failed LoadFile:", err)
		return
	}
	if !isEqualByteArrays(finalVerse, file) {
		t.Error("File contents changed after revoke")
	}

	alice2, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to reload user Alice", err)
		return
	}
	file, err = alice2.LoadFile("rhyme")
	if err != nil || file == nil {
		t.Error("Failed LoadFile userdata contents did not persist:", err)
		return
	}
	if !isEqualByteArrays(finalVerse, file) {
		t.Error("File contents changed")
	}

	bo2, err := GetUser("bo", "babc123")
	if err != nil {
		t.Error("Failed to reload user Bo", err)
		return
	}
	file, err = bo2.LoadFile("goo2")
	if err != nil || file == nil {
		t.Error("Failed LoadFile userdata shared contents did not persist:", err)
		return
	}
	if !isEqualByteArrays(finalVerse, file) {
		t.Error("File contents changed")
	}

	userlib.DatastoreClear()
}

func fillDataStore(t *testing.T) {
	_, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
	}
	_, err = InitUser("bob", "abc123")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
	}
	_, err = InitUser("bo", "babc123")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
	}
	_, err = InitUser("boba", "bc123")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
	}
}

func isEqualByteArrays(arr1 []byte, arr2 []byte) bool {
	if len(arr1) != len(arr2) {
		return false
	}
	for i, _ := range arr1 {
		if arr1[i] != arr2[i] {
			return false
		}
	}
	return true
}
