package main

import (
	"log"
	"net/http"
	agt "projetGo/pkg/agt"
	"runtime"
	"runtime/debug"
)

func simulation1() {
	// 30 livreurs, 100 clients sur 5 heures , 25 fois plus rapide que la réalité sur compiègne
	e := agt.NewEnvironment(
		"../../../projetPython/graph.json",
		"../../../projetPython/appartements.json",
		"../../../projetPython/restaurants.json",
		1000,
		0.3,
		0.04,
		25,
		0.5)

	go agt.Simulate(30, 100, 5, e)

	http.HandleFunc("/ws", agt.WsHandler(e))

	log.Println("WebSocket server started on ws://localhost:4444/ws")
	if err := http.ListenAndServe(":4444", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func simulation2() {
	//200 fois plus rapide, 200 livreurs, 700 clients à Lille
	e := agt.NewEnvironment(
		"../../../projetPython/graphLille.json",
		"../../../projetPython/appartementsLille.json",
		"../../../projetPython/restaurantsLille.json",
		1000,
		0.3,
		0.005,
		25,
		0.5)

	go agt.Simulate(200, 700, 5, e)

	http.HandleFunc("/ws", agt.WsHandler(e))

	log.Println("WebSocket server started on ws://localhost:4444/ws")
	if err := http.ListenAndServe(":4444", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func simulation3() {
	// 600 livreurs, 500 clients dans Lille
	e := agt.NewEnvironment(
		"../../../projetPython/graphLille.json",
		"../../../projetPython/appartementsLille.json",
		"../../../projetPython/restaurantsLille.json",
		1000,
		0.3,
		0.1,
		25,
		0.5)

	go agt.Simulate(600, 500, 5, e)

	http.HandleFunc("/ws", agt.WsHandler(e))

	log.Println("WebSocket server started on ws://localhost:4444/ws")
	if err := http.ListenAndServe(":4444", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func simulation4() {
	//10000 livreur, 50000 clients sur 5 heures
	e := agt.NewEnvironment(
		"../../../projetPython/graph.json",
		"../../../projetPython/appartements.json",
		"../../../projetPython/restaurants.json",
		1000,
		0.3,
		0.001,
		25,
		0.5)

	go agt.Simulate(10000, 50000, 5, e)

	http.HandleFunc("/ws", agt.WsHandler(e))

	log.Println("WebSocket server started on ws://localhost:4444/ws")
	if err := http.ListenAndServe(":4444", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func main() {
	/* à de commenter pour le profilage et pour obtenir des données
	printMemStats := func() {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		fmt.Printf("Mémoire allouée : %d KB\n", memStats.Alloc/1024)
		fmt.Printf("Mémoire totale allouée : %d KB\n", memStats.TotalAlloc/1024)
		fmt.Printf("Mémoire système obtenue : %d KB\n", memStats.Sys/1024)
		fmt.Printf("Nombre de GC : %d\n", memStats.NumGC)
		fmt.Println()
	}

	*/
	runtime.GOMAXPROCS(4)
	debug.SetMemoryLimit(8 << 30) // 1 Go

	// f, err := os.Create("cpu.prof")
	// if err != nil {
	// 	log.Fatal("Could not create CPU profile: ", err)
	// }

	// fmem, errmem := os.Create("mem.prof")
	// if errmem != nil {
	// 	log.Fatal("Could not create memory profile: ", err)
	// }

	// Démarrer le profilage CPU
	// if err := pprof.StartCPUProfile(f); err != nil {
	// 	log.Fatal("Could not start CPU profile: ", err)
	// }
	// go agt.Simulate(20, 2, 10, e)

	// pprof.StopCPUProfile()
	// if err := pprof.WriteHeapProfile(fmem); err != nil {
	// 	log.Fatal("Could not write heap profile: ", err)
	// }
	// f.Close()
	// fmem.Close()

	// http.HandleFunc("/ws", agt.WsHandler(e))

	// log.Println("WebSocket server started on ws://localhost:4444/ws")
	// if err := http.ListenAndServe(":4444", nil); err != nil {
	// 	log.Fatalf("Server failed: %v", err)
	// }

	// slowSimu()
	simulation1()
	// simulation2()
	// simulation3()
	// simulation4()
}
