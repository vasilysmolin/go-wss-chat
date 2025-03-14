package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	centrifuge "github.com/centrifugal/centrifuge-go"
)

func main() {
	// Получаем URL Centrifugo из переменных окружения
	centrifugoURL := os.Getenv("CENTRIFUGO_URL")
	if centrifugoURL == "" {
		centrifugoURL = "ws://localhost:8000/connection/websocket"
	}

	// Создаём новый клиент Centrifuge с использованием JSON-сериализации
	client, err := centrifuge.NewJsonClient(centrifugoURL, centrifuge.Config{
		// Здесь можно настроить опции клиента, например, авто-переподключение
	})
	if err != nil {
		log.Fatalf("Ошибка создания клиента: %v", err)
	}

	// Обработчик успешного подключения
	client.OnConnect(func(ctx context.Context, e centrifuge.ConnectEvent) (centrifuge.ConnectResult, error) {
		fmt.Println("Подключились к Centrifugo!")
		return centrifuge.ConnectResult{}, nil
	})

	// Обработчик отключения
	client.OnDisconnect(func(e centrifuge.DisconnectEvent) {
		fmt.Printf("Отключились: %v\n", e.Reason)
	})

	// Подключаемся к Centrifugo
	if err := client.Connect(); err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}

	// Создаём подписку на канал "chat"
	sub, err := client.NewSubscription("chat")
	if err != nil {
		log.Fatalf("Ошибка создания подписки: %v", err)
	}

	// Обработчик входящих сообщений
	sub.OnPublish(func(event centrifuge.PublicationEvent) {
		fmt.Printf("Получено сообщение: %s\n", event.Data)
	})

	// Подписываемся на канал
	if err := sub.Subscribe(); err != nil {
		log.Fatalf("Ошибка подписки: %v", err)
	}

	// Публикуем сообщение каждые 5 секунд (для демонстрации)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			<-ticker.C
			data := map[string]interface{}{
				"text": fmt.Sprintf("Привет! Текущее время: %s", time.Now().Format(time.RFC3339)),
			}
			// Публикация сообщения на канал.
			// Заметьте: по умолчанию клиентская публикация может быть запрещена – для теста можно включить publish в конфигурации Centrifugo.
			if err := sub.Publish(data); err != nil {
				log.Printf("Ошибка публикации: %v", err)
			} else {
				fmt.Println("Сообщение опубликовано")
			}
		}
	}()

	// Ожидаем сигнала прерывания для корректного завершения
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	fmt.Println("Завершаем работу...")
	sub.Unsubscribe()
	client.Close()
}
