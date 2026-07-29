package authpass

import "testing"

func TestVerifySha1(t *testing.T) {
	plain := "password"
	sha := "5baa61e4c9b93f3f0682250b6cf8331b7ee68fd8"
	if !Verify(plain, sha) {
		t.Fatal("sha1 verify failed")
	}
	if !NeedsRehash(sha) {
		t.Fatal("sha1 should need rehash")
	}
}

func TestVerifyArgon2(t *testing.T) {
	plain := "password"
	h, err := Hash(plain)
	if err != nil {
		t.Fatal(err)
	}
	if !Verify(plain, h) {
		t.Fatal("argon2 verify failed")
	}
	if NeedsRehash(h) {
		t.Fatal("argon2 should not need rehash")
	}
}

func TestVerifyBcrypt(t *testing.T) {
	h := "$2y$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi"
	if !Verify("password", h) {
		t.Fatal("bcrypt verify failed")
	}
	if !NeedsRehash(h) {
		t.Fatal("bcrypt should need rehash")
	}
}
