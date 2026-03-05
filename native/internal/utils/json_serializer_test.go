package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
}

type Product struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

func TestMarshal(t *testing.T) {
	t.Run("序列化结构体", func(t *testing.T) {
		user := User{ID: 1, Username: "admin", Email: "admin@example.com", Age: 25}
		jsonStr, err := Marshal(user)
		assert.NoError(t, err)
		assert.Contains(t, jsonStr, "admin")
	})

	t.Run("序列化切片", func(t *testing.T) {
		users := []User{
			{ID: 1, Username: "user1", Email: "user1@example.com", Age: 20},
			{ID: 2, Username: "user2", Email: "user2@example.com", Age: 30},
		}
		jsonStr, err := Marshal(users)
		assert.NoError(t, err)
		assert.Contains(t, jsonStr, "user1")
	})

	t.Run("序列化map", func(t *testing.T) {
		data := map[string]int{"apple": 10, "banana": 20}
		jsonStr, err := Marshal(data)
		assert.NoError(t, err)
		assert.Contains(t, jsonStr, "apple")
	})
}

func TestMarshalIndent(t *testing.T) {
	user := User{ID: 1, Username: "admin", Email: "admin@example.com", Age: 25}
	jsonStr, err := MarshalIndent(user, "", "  ")
	assert.NoError(t, err)
	assert.Contains(t, jsonStr, "\n")
	assert.Contains(t, jsonStr, "  ")
}

func TestToBytes(t *testing.T) {
	user := User{ID: 1, Username: "admin", Email: "admin@example.com", Age: 25}
	bytes, err := ToBytes(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, bytes)
}

func TestUnmarshal(t *testing.T) {
	t.Run("反序列化结构体", func(t *testing.T) {
		jsonStr := `{"id":1,"username":"admin","email":"admin@example.com","age":25}`
		user, err := Unmarshal[User](jsonStr)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "admin", user.Username)
	})

	t.Run("反序列化切片", func(t *testing.T) {
		jsonStr := `[{"id":1,"username":"user1","email":"user1@example.com","age":20}]`
		users, err := Unmarshal[[]User](jsonStr)
		assert.NoError(t, err)
		assert.Len(t, users, 1)
	})

	t.Run("反序列化map", func(t *testing.T) {
		jsonStr := `{"apple":10,"banana":20}`
		data, err := Unmarshal[map[string]int](jsonStr)
		assert.NoError(t, err)
		assert.Equal(t, 10, data["apple"])
	})

	t.Run("空字符串返回错误", func(t *testing.T) {
		_, err := Unmarshal[User]("")
		assert.Error(t, err)
	})
}

func TestFromBytes(t *testing.T) {
	jsonBytes := []byte(`{"id":1,"username":"admin","email":"admin@example.com","age":25}`)
	user, err := FromBytes[User](jsonBytes)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, "admin", user.Username)
}

func TestRoundTrip(t *testing.T) {
	t.Run("结构体往返", func(t *testing.T) {
		original := User{ID: 1, Username: "admin", Email: "admin@example.com", Age: 25}
		jsonStr, err := Marshal(original)
		assert.NoError(t, err)
		result, err := Unmarshal[User](jsonStr)
		assert.NoError(t, err)
		assert.Equal(t, original, result)
	})

	t.Run("切片往返", func(t *testing.T) {
		original := []Product{
			{Name: "Product1", Price: 99.99, Stock: 100},
			{Name: "Product2", Price: 149.99, Stock: 50},
		}
		jsonStr, err := Marshal(original)
		assert.NoError(t, err)
		result, err := Unmarshal[[]Product](jsonStr)
		assert.NoError(t, err)
		assert.Equal(t, original, result)
	})

	t.Run("map往返", func(t *testing.T) {
		original := map[string]int{"apple": 10, "banana": 20}
		jsonStr, err := Marshal(original)
		assert.NoError(t, err)
		result, err := Unmarshal[map[string]int](jsonStr)
		assert.NoError(t, err)
		assert.Equal(t, original, result)
	})

	t.Run("字节数组往返", func(t *testing.T) {
		original := User{ID: 1, Username: "admin", Email: "admin@example.com", Age: 25}
		bytes, err := ToBytes(original)
		assert.NoError(t, err)
		result, err := FromBytes[User](bytes)
		assert.NoError(t, err)
		assert.Equal(t, original, result)
	})
}

func TestComplexTypes(t *testing.T) {
	t.Run("嵌套结构体", func(t *testing.T) {
		type Address struct {
			City    string `json:"city"`
			Country string `json:"country"`
		}
		type UserWithAddress struct {
			User
			Address Address `json:"address"`
		}
		original := UserWithAddress{
			User:    User{ID: 1, Username: "admin", Email: "admin@example.com", Age: 25},
			Address: Address{City: "Beijing", Country: "China"},
		}
		jsonStr, err := Marshal(original)
		assert.NoError(t, err)
		result, err := Unmarshal[UserWithAddress](jsonStr)
		assert.NoError(t, err)
		assert.Equal(t, original.Username, result.Username)
		assert.Equal(t, original.Address.City, result.Address.City)
	})

	t.Run("复杂map", func(t *testing.T) {
		original := map[string][]User{
			"admins": {{ID: 1, Username: "admin1", Email: "admin1@example.com", Age: 25}},
			"users":  {{ID: 2, Username: "user1", Email: "user1@example.com", Age: 20}},
		}
		jsonStr, err := Marshal(original)
		assert.NoError(t, err)
		result, err := Unmarshal[map[string][]User](jsonStr)
		assert.NoError(t, err)
		assert.Len(t, result["admins"], 1)
		assert.Equal(t, "admin1", result["admins"][0].Username)
	})
}
