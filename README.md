# lockbox
A cloud based file storage system, implemented on top of a secure cryptographic protocol for trustless server hosting.

Part 1:

key : value

keyStore:
username : publicKey

dataStore:

logins/<u> : (x, MAC(x))
x = Ep(Userdata) where Ep = block cipher encryption
verify should not be done with RSA public key from keystore: leaks data about username??
Maybe verify using public key, idk…
verify should be done with MAC

u = HMAC(k, username)
k = PBKDF2(password, username, 256)
p = PBKDF2(password, username, 128) 	// k guaranteed not to leak info about p?

TODO:
write HMAC() function that generates random IV already, etc. Makes code cleaner.


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