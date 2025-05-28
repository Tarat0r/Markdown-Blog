package notifications

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
)

func NotifyTelegram(text string) error {
	token := os.Getenv("TG_TOKEN")
	chatID := os.Getenv("TG_CHAT_ID")

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	body := []byte(fmt.Sprintf("chat_id=%s&text=%s", chatID, text))

	resp, err := http.Post(url, "application/x-www-form-urlencoded", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status %s", resp.Status)
	}
	return nil
}
