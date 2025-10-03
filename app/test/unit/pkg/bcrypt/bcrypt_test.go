package bcrypt_test

import (
	"backend/service-platform/app/pkg/bcrypt"
	"fmt"
	"strings"
	"testing"

	gocrypt "golang.org/x/crypto/bcrypt"
)

func TestNewBcrypt(t *testing.T) {
	tests := []struct {
		name         string
		cost         int
		expectedCost int
	}{
		{
			name:         "valid cost within range",
			cost:         12,
			expectedCost: 12,
		},
		{
			name:         "cost below minimum defaults to default",
			cost:         3,
			expectedCost: gocrypt.DefaultCost,
		},
		{
			name:         "cost above maximum defaults to default",
			cost:         32,
			expectedCost: gocrypt.DefaultCost,
		},
		{
			name:         "minimum cost",
			cost:         gocrypt.MinCost,
			expectedCost: gocrypt.MinCost,
		},
		{
			name:         "maximum cost",
			cost:         gocrypt.MaxCost,
			expectedCost: gocrypt.MaxCost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := bcrypt.NewBcrypt(tt.cost)
			if hasher.Cost() != tt.expectedCost {
				t.Errorf("NewBcrypt() cost = %v, expected %v", hasher.Cost(), tt.expectedCost)
			}
		})
	}
}

func TestNewBcryptWithDefaultCost(t *testing.T) {
	hasher := bcrypt.NewBcryptWithDefaultCost()
	if hasher.Cost() != gocrypt.DefaultCost {
		t.Errorf("NewBcryptWithDefaultCost() cost = %v, expected %v", hasher.Cost(), gocrypt.DefaultCost)
	}
}

func TestBcrypt_HashPassword(t *testing.T) {
	hasher := bcrypt.NewBcryptWithDefaultCost()

	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "valid password",
			password:    "mySecurePassword123!",
			expectError: false,
		},
		{
			name:        "simple password",
			password:    "password",
			expectError: false,
		},
		{
			name:        "long password",
			password:    strings.Repeat("a", 72), // bcrypt max input length
			expectError: false,
		},
		{
			name:        "password with special characters",
			password:    "p@$$w0rd!@#$%^&*()",
			expectError: false,
		},
		{
			name:        "empty password",
			password:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := hasher.HashPassword(tt.password)

			if tt.expectError {
				if err == nil {
					t.Errorf("HashPassword() expected error for password: %q", tt.password)
				}
				if hash != "" {
					t.Errorf("HashPassword() expected empty hash on error, got: %q", hash)
				}
				return
			}

			if err != nil {
				t.Errorf("HashPassword() unexpected error: %v", err)
				return
			}

			if hash == "" {
				t.Error("HashPassword() returned empty hash")
				return
			}

			// Verify hash format (bcrypt hashes start with $2a$, $2b$, or $2y$)
			if !strings.HasPrefix(hash, "$2") {
				t.Errorf("HashPassword() returned invalid hash format: %q", hash)
			}

			// Verify the hash can be used to compare against the original password
			err = gocrypt.CompareHashAndPassword([]byte(hash), []byte(tt.password))
			if err != nil {
				t.Errorf("HashPassword() produced hash that doesn't match original password: %v", err)
			}
		})
	}
}

func TestBcrypt_CheckPassword(t *testing.T) {
	hasher := bcrypt.NewBcryptWithDefaultCost()
	testPassword := "testPassword123!"

	// Create a valid hash for testing
	validHash, err := hasher.HashPassword(testPassword)
	if err != nil {
		t.Fatalf("Failed to create test hash: %v", err)
	}

	tests := []struct {
		name           string
		password       string
		hash           string
		expectedResult bool
		expectError    bool
	}{
		{
			name:           "correct password and hash",
			password:       testPassword,
			hash:           validHash,
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "incorrect password",
			password:       "wrongPassword",
			hash:           validHash,
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "empty password",
			password:       "",
			hash:           validHash,
			expectedResult: false,
			expectError:    true,
		},
		{
			name:           "empty hash",
			password:       testPassword,
			hash:           "",
			expectedResult: false,
			expectError:    true,
		},
		{
			name:           "invalid hash format",
			password:       testPassword,
			hash:           "invalid_hash",
			expectedResult: false,
			expectError:    true,
		},
		{
			name:           "empty password and hash",
			password:       "",
			hash:           "",
			expectedResult: false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := hasher.CheckPassword(tt.password, tt.hash)

			if tt.expectError {
				if err == nil {
					t.Errorf("CheckPassword() expected error but got none")
				}
				if result != false {
					t.Errorf("CheckPassword() expected false result on error, got: %v", result)
				}
				return
			}

			if err != nil {
				t.Errorf("CheckPassword() unexpected error: %v", err)
				return
			}

			if result != tt.expectedResult {
				t.Errorf("CheckPassword() result = %v, expected %v", result, tt.expectedResult)
			}
		})
	}
}

func TestBcrypt_Interface(t *testing.T) {
	var _ bcrypt.Hasher = &bcrypt.Bcrypt{}

	// Test that both constructors return types that implement the interface
	hasher1 := bcrypt.NewBcrypt(10)
	hasher2 := bcrypt.NewBcryptWithDefaultCost()

	// Test interface methods exist and work
	password := "testPassword"

	hash1, err := hasher1.HashPassword(password)
	if err != nil {
		t.Errorf("Interface method HashPassword failed: %v", err)
	}

	valid1, err := hasher1.CheckPassword(password, hash1)
	if err != nil || !valid1 {
		t.Errorf("Interface method CheckPassword failed: err=%v, valid=%v", err, valid1)
	}

	hash2, err := hasher2.HashPassword(password)
	if err != nil {
		t.Errorf("Interface method HashPassword failed: %v", err)
	}

	valid2, err := hasher2.CheckPassword(password, hash2)
	if err != nil || !valid2 {
		t.Errorf("Interface method CheckPassword failed: err=%v, valid=%v", err, valid2)
	}
}

func TestBcrypt_DifferentCosts(t *testing.T) {
	password := "testPassword"
	// Test only reasonable costs to keep tests fast
	costs := []int{gocrypt.MinCost, 6, 8, 10}

	for _, cost := range costs {
		t.Run(fmt.Sprintf("cost_%d", cost), func(t *testing.T) {
			hasher := bcrypt.NewBcrypt(cost)

			hash, err := hasher.HashPassword(password)
			if err != nil {
				t.Errorf("HashPassword with cost %d failed: %v", cost, err)
				return
			}

			valid, err := hasher.CheckPassword(password, hash)
			if err != nil {
				t.Errorf("CheckPassword with cost %d failed: %v", cost, err)
				return
			}

			if !valid {
				t.Errorf("CheckPassword with cost %d returned false for valid password", cost)
			}
		})
	}
}

func TestBcrypt_HashUniqueness(t *testing.T) {
	hasher := bcrypt.NewBcryptWithDefaultCost()
	password := "samePassword"

	// Generate multiple hashes of the same password
	hashes := make([]string, 5)
	for i := range hashes {
		hash, err := hasher.HashPassword(password)
		if err != nil {
			t.Fatalf("HashPassword failed on iteration %d: %v", i, err)
		}
		hashes[i] = hash
	}

	// Verify all hashes are different (due to salt)
	for i, hash1 := range hashes {
		for j, hash2 := range hashes {
			if i != j && hash1 == hash2 {
				t.Errorf("Hashes should be unique due to salt, but got duplicate: %q", hash1)
			}
		}
	}

	// Verify all hashes validate the same password
	for i, hash := range hashes {
		valid, err := hasher.CheckPassword(password, hash)
		if err != nil {
			t.Errorf("CheckPassword failed for hash %d: %v", i, err)
		}
		if !valid {
			t.Errorf("Hash %d should validate the original password", i)
		}
	}
}

func TestBcrypt_EdgeCases(t *testing.T) {
	t.Run("very long password", func(t *testing.T) {
		hasher := bcrypt.NewBcrypt(gocrypt.MinCost) // Use minimum cost for speed
		// bcrypt rejects passwords longer than 72 bytes
		longPassword := strings.Repeat("a", 100)

		_, err := hasher.HashPassword(longPassword)
		if err == nil {
			t.Error("HashPassword should fail with password longer than 72 bytes")
		}
		if !strings.Contains(err.Error(), "failed to hash password") {
			t.Errorf("Expected hash error for long password, got: %v", err)
		}

		// Test with maximum valid length (72 bytes)
		maxPassword := strings.Repeat("a", 72)
		hash, err := hasher.HashPassword(maxPassword)
		if err != nil {
			t.Errorf("HashPassword failed with 72-byte password: %v", err)
		}

		valid, err := hasher.CheckPassword(maxPassword, hash)
		if err != nil || !valid {
			t.Errorf("CheckPassword failed with 72-byte password: err=%v, valid=%v", err, valid)
		}
	})

	t.Run("unicode password", func(t *testing.T) {
		hasher := bcrypt.NewBcrypt(gocrypt.MinCost)
		unicodePassword := "测试密码🔐"

		hash, err := hasher.HashPassword(unicodePassword)
		if err != nil {
			t.Errorf("HashPassword failed with unicode password: %v", err)
		}

		valid, err := hasher.CheckPassword(unicodePassword, hash)
		if err != nil || !valid {
			t.Errorf("CheckPassword failed with unicode password: err=%v, valid=%v", err, valid)
		}
	})

	t.Run("password with null bytes", func(t *testing.T) {
		hasher := bcrypt.NewBcrypt(gocrypt.MinCost)
		passwordWithNull := "password\x00with\x00nulls"

		hash, err := hasher.HashPassword(passwordWithNull)
		if err != nil {
			t.Errorf("HashPassword failed with null bytes: %v", err)
		}

		valid, err := hasher.CheckPassword(passwordWithNull, hash)
		if err != nil || !valid {
			t.Errorf("CheckPassword failed with null bytes: err=%v, valid=%v", err, valid)
		}
	})
}

func TestBcrypt_ErrorMessages(t *testing.T) {
	hasher := bcrypt.NewBcryptWithDefaultCost()

	t.Run("hash password error messages", func(t *testing.T) {
		_, err := hasher.HashPassword("")
		if err == nil {
			t.Error("Expected error for empty password")
		}
		if !strings.Contains(err.Error(), "password cannot be empty") {
			t.Errorf("Error message should mention empty password, got: %v", err)
		}
	})

	t.Run("check password error messages", func(t *testing.T) {
		validHash, _ := hasher.HashPassword("test")

		// Test empty password error
		_, err := hasher.CheckPassword("", validHash)
		if err == nil {
			t.Error("Expected error for empty password")
		}
		if !strings.Contains(err.Error(), "password cannot be empty") {
			t.Errorf("Error message should mention empty password, got: %v", err)
		}

		// Test empty hash error
		_, err = hasher.CheckPassword("test", "")
		if err == nil {
			t.Error("Expected error for empty hash")
		}
		if !strings.Contains(err.Error(), "hash cannot be empty") {
			t.Errorf("Error message should mention empty hash, got: %v", err)
		}

		// Test invalid hash error
		_, err = hasher.CheckPassword("test", "invalid")
		if err == nil {
			t.Error("Expected error for invalid hash")
		}
		if !strings.Contains(err.Error(), "failed to check password") {
			t.Errorf("Error message should mention failed check, got: %v", err)
		}
	})
}
