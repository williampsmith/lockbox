package proj2

import "testing"
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

	array := make([]byte, 10, 20)
	t.Log("Original file len:", len(array))
	t.Log("Original file cap:", cap(array))

	DebugPrint = true
	user.StoreFile("LenCap", array)

	var file []byte
	file, err = user.LoadFile("LenCap")
	if err != nil {
		t.Error("Failed LoadFile:", err)
		return
	}
	t.Log("Original file len:", len(file))
	t.Log("Original file cap:", cap(file))
}
