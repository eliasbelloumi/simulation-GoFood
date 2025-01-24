package agt

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocket handler
func WsHandler(e *Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: 	%v", err)
			return
		}
		defer conn.Close()

		log.Println("Client connected")
		err = conn.WriteMessage(websocket.TextMessage, e.SerializeInit())
		if err != nil {
			log.Printf("Error sending init data: %v", err)
			return
		}

		for {
			if e.newStat {
				err = conn.WriteMessage(websocket.TextMessage, e.SerializeTotalStatistics())
				if err != nil {
					log.Printf("Error sending totalStat data: %v", err)
					return
				}
				err = conn.WriteMessage(websocket.TextMessage, e.SerializeStatistics())
				if err != nil {
					log.Printf("Error sending stats: %v", err)
					return
				}
				e.newStat = false
			}

			restaurantJson := e.SerializeRestaurants()
			err = conn.WriteMessage(websocket.TextMessage, restaurantJson)
			if err != nil {
				log.Printf("Error sending restaurant data: %v", err)
				return
			}
			err = conn.WriteMessage(websocket.TextMessage, e.SerializeCustomers())
			if err != nil {
				log.Printf("Error sending customer data: %v", err)
				return
			}

			deliverersJson := e.SerializeDeliverers()
			err = conn.WriteMessage(websocket.TextMessage, deliverersJson)
			if err != nil {
				log.Printf("Error sending deliverer data: %v", err)
				return
			}
			time.Sleep(2 * time.Second)

		}
	}
}
