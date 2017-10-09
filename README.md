# lockbox

## A cloud based file storage system, implemented on top of a secure cryptographic protocol for trustless server hosting.

### Part 1:

Notation -- key : value

#### User struct
```
Username
PrivateKey
PublicKey
... // add to this
```

#### keyStore
store `username : publicKey`

#### dataStore:
```
// kh is a 256 bit key, and kp is a 128 bit key
(kh, kp) = PBKDF2(password, username, salt=SHA256(password), length=384)
```

`x = E_kp(Userdata)` where `E_kp` is a block cipher encryption on key `kp`

`u = HMAC(kh, username)  // this will fail if password is incorrect`

store `logins/<u> : (x, RSA_SIG(publick_key(username), x))`

TODO:
write `HMAC()` function that generates random IV already, etc. Makes code cleaner.


Design 1:
O(n) updates, n = number of files stored in server. Not sure how they got O(n) here??

	<user>/dir_keys  : x || SignRSA(x)			x = RSA(ke, ka)
	<user>/directory : y, MAC_ka(directory, y)	y = E_ke(r, k1, k2)
	<user>files/r	    : z, MAC_k2(r, z)			z = E_k1(contents)

		r = random ID from directory listing
		New files uploaded: generate new random r, k1, k2
			^ which is probably where O(n) comes from since have to search to make sure it’s not used already

Design 2:
O(1) updates

	info/<user> : x || SignRSA(x)     			x = RSA(ke, ka, kn)
ID	<user>/r	    : y || MAC_ka(filename, y)		y = E_ke(contents)

		including filename in MAC prevents malicious server from swapping files between IDs
		E = CFB
		MAC = SHA256_HMAC
		r = ID = E_kn(filename) = SHA256_HMAC



A file ID (FID) will be generated via HMAC on the filename

And a file will be stored at some location of FID in the data store.

Something like:
userID/FID

Where userID is the HMAC on the username (to keep confidentiality).

My approach for efficient append is as follows:

FID at HMAC(filename) will contain a counter of how many appends were made for the filename
Then I can store the same file at HMAC(filename0), HMAC(filename1), etc.


### Part 2:

Each user has 3 symmetric master keys: ke, ka, kn
		2 new sharing keys ke', ka' per file shared
		For each user they share with, generate 2 new keys ke'', ka''

<original user>/r1 : w, HMAC_ka(w)		w = E_ke(ke’, ka’, r2, filename)
			 r2 : x, HMAC_ka’(x)		x = E_ke’(file contents)

r1 = HMAC_kn(filename)
For each new user we share r2 with, generate new r3, ke’’, ka’’, stores this info.
Say we share with user 2.
Original user stores information in r3, used to decrypt stuff at r2

			r3 : y, HMAC_ka’’(y)			y = E_ke’’(ke’, ka’, r2)

Original user then sends to user 2: m, signRSA1(m) 	m = RSA2(ke’’, ka’’, r3)
RSA2 = user 2 public key
signRSA1 = signed with user 1 private key

User 2 names it with a different filename and stores at r4. User 2 also uses his universal master keys, ke2, ka2
	<user 2> / r4 : z, HMAC_ka2(z)			z = ke2(ke’’, ka’’, r3)

r4 = HMAC_kn2(new filename)

Lets say there’s also a user 3, but original user decides to revoke access to user 3.
Original user changes ka’, ke’ and needs to re-encrypt location r2 and also update location r3.



