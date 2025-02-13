package entity

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUser(t *testing.T) {
	// Создаем тестового пользователя
	user := User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Coins:        1000,
		CreatedAt:    time.Now(),
	}

	// Тестируем JSON сериализацию
	jsonData, err := json.Marshal(user)
	if err != nil {
		t.Errorf("Failed to marshal user: %v", err)
	}

	// Тестируем JSON десериализацию
	var unmarshaled User
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal user: %v", err)
	}

	// Проверяем что поля совпадают
	if unmarshaled.ID != user.ID {
		t.Errorf("ID mismatch: got %v want %v", unmarshaled.ID, user.ID)
	}
	if unmarshaled.Username != user.Username {
		t.Errorf("Username mismatch: got %v want %v", unmarshaled.Username, user.Username)
	}
	if unmarshaled.Coins != user.Coins {
		t.Errorf("Coins mismatch: got %v want %v", unmarshaled.Coins, user.Coins)
	}
	// PasswordHash не должен быть в JSON
	if unmarshaled.PasswordHash != "" {
		t.Error("PasswordHash should not be present in JSON")
	}
}
