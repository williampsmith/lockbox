package proj2

import "testing"
import "github.com/nweaver/cs161-p2/userlib"
// You can actually import other stuff if you want IN YOUR TEST
// HARNESS ONLY.  Note that this is NOT considered part of your
// solution, but is how you make sure your solution is correct.

func TestInit(t *testing.T){
	t.Log("Initialization test")
	DebugPrint = false
	someUsefulThings()

	DebugPrint = false
	u, err := InitUser("alice","fubar")
	if err != nil {
		// t.Error says the test fails 
		t.Error("Failed to initialize user", err)
	}
	// t.Log() only produces output if you run with "go test -v"
	t.Log("Got user", u)
	// You probably want many more tests here.
}


func TestStorage(t *testing.T){
	// And some more tests, because
	v, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to reload user", err)
		return
	}
	t.Log("Loaded user", v)
}

func TestLenCap(t *testing.T){
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
	if err != nil {
		t.Error("Failed LoadFile:", err)
		return
	}

	if !isEqualByteArrays(array, file) {
		t.Error("Load failure, wrong contents")
	}

	userlib.DatastoreClear()
	file, err = user.LoadFile("LenCap")
	if err == nil {
		// Expecting there to be an error, but there was no error
		t.Error("Failed LoadFile expecting an error, but got no error")
		return
	}
}

func TestUserCollisions(t *testing.T){
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
	if (userA.Username == userB.Username) || (userA.Username == userC.Username) || (userB.Username == userC.Username) {
		t.Error("User data login caused collision")
		return
	}
	if (userA.PublicKey == userB.PublicKey) || (userA.PublicKey == userC.PublicKey) || (userB.PublicKey == userC.PublicKey) {
		t.Error("User data login caused collision")
		return
	}
}


func fillDataStore(t *testing.T) {
	_, err := InitUser("alice","fubar")
	if err != nil {
		// t.Error says the test fails 
		t.Error("Failed to initialize user", err)
	}
	_, err = InitUser("bob","abc123")
	if err != nil {
		// t.Error says the test fails 
		t.Error("Failed to initialize user", err)
	}
	_, err = InitUser("bo","babc123")
	if err != nil {
		// t.Error says the test fails 
		t.Error("Failed to initialize user", err)
	}
	_, err = InitUser("boba","bc123")
	if err != nil {
		// t.Error says the test fails 
		t.Error("Failed to initialize user", err)
	}
}

func isEqualByteArrays(arr1 []byte, arr2 []byte) bool {
	if len(arr1) != len(arr2){
		return false
	}
	for i, _ := range arr1 {
    	if arr1[i] != arr2[i] {
    		return false
    	}
	}
	return true
}