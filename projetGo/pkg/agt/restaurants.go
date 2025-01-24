package agt

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"golang.org/x/exp/rand"
)

type Restaurant struct {
	e                   *Environment
	Osmid               OsmId    `json:"osmid"`
	Name                string   `json:"name"`
	Amenity             string   `json:"amenity"`
	Cuisine             FoodType `json:"cuisine"`
	schedules           [2]time.Time
	price               Price         // prix d'un plat
	Geometry            [2]Coordinate `json:"geometry"` // lieu géographique
	ClosestNode         NodeId        `json:"closest_node"`
	DistanceClosestNode float64       `json:"distance_noeud"` // distance du noeud le plus proche

	NumberRestaurantsNear int16 // nombre de restaurants proches

	prepTimeWaiter chan time.Duration // permet la gestion des attributs en concurrence
	stockWaiter    chan Stock

	PreparationTime  time.Duration
	waitingTime      time.Duration // duree d'attente totale pour une commande
	stock            Stock         // MODIFICATIONS FREQUENTES DE MANIERE CONCURRENTE
	inProgressOrders sync.Map      // gere les modifications concurrentes avec un sync.Map

	nbPlateWaiter     chan Stock
	nbCommandes       int //nombre de commandes qu'il reste à faire
	nbCommandesWaiter chan int
	nbPlates          Stock
	readyOrder        sync.Map // contient l'ensemble des commandes pretes en attente de retrait
	working           bool     // l'agent est actif
	action            int8
	count             int
	score             float64 // calcule le score d'un restaurant (pour gojek)
	NormalizedScore   float64
	stop              bool // indique si le restaurant accepte encore ou non des commandes
	Activated         bool // indique si le restaurant
}

func (r *Restaurant) CalculateScore() {
	prepTime := r.ReturnPreparationTime()
	if prepTime == 0 {
		r.score = 0 // Éviter la division par zéro
	}
	r.score = float64(r.ReturnPrice()) / float64(prepTime.Minutes())
}

func (r *Restaurant) ReturnScore() float64 {
	return r.score
}

func (r *Restaurant) ReturnFirstInProgressOrder() *Order {
	var oldestOrder *Order
	var oldestKey any

	r.inProgressOrders.Range(func(key, value any) bool {
		if order, ok := value.(*Order); ok {
			if oldestOrder == nil || order.BeginningOrdering.Before(oldestOrder.BeginningOrdering) {
				oldestOrder = order
				oldestKey = key
			}
		}
		return true // Continuer à parcourir toutes les commandes
	})

	// Supprimer la commande avec la plus ancienne date si trouvée
	if oldestKey != nil {
		r.inProgressOrders.Delete(oldestKey)
	}

	return oldestOrder
}

func (r *Restaurant) SetOpening() {
	// indique l'heure d'ouverture et de fermeture du restaurant de manière aléatoire
	openingHours := []int{9, 10, 11, 12}
	openingMin := []int{0, 15, 30, 45}
	closingHours := []int{21, 22, 23, 00, 01, 02}

	// on va créer l'heure d'ouverture:
	r.schedules[0] = time.Date(0, time.January, 1, openingHours[rand.Intn(len(openingHours))], openingMin[rand.Intn(len(openingMin))], 0, 0, time.UTC)

	heureFermeture := closingHours[rand.Intn(len(closingHours))]
	if heureFermeture < 21 {
		r.schedules[1] = time.Date(0, time.January, 2, heureFermeture, openingMin[rand.Intn(len(openingMin))], 0, 0, time.UTC)
	} else {
		r.schedules[1] = time.Date(0, time.January, 1, heureFermeture, openingMin[rand.Intn(len(openingMin))], 0, 0, time.UTC)
	}

}

func (r *Restaurant) IsOpen(jour time.Time) bool {
	//vérifie que le restaurant sera ouvert à l'heure de jour
	heureOuverture, minOuverture, _ := r.schedules[0].Clock()
	openingTime := jour.Truncate(24 * time.Hour).Add(time.Duration(heureOuverture) * time.Hour).Add(time.Duration(minOuverture) * time.Minute)

	heureFermeture, minFermeture, _ := r.schedules[1].Clock()

	var closingTime time.Time
	if heureFermeture > 5 {
		closingTime = jour.Truncate(24 * time.Hour).Add(time.Duration(heureFermeture) * time.Hour).Add(time.Duration(minFermeture) * time.Minute)
	} else {
		closingTime = jour.Add(24 * time.Hour)
		closingTime = closingTime.Truncate(24 * time.Hour).Add(time.Duration(heureFermeture) * time.Hour).Add(time.Duration(minFermeture) * time.Minute)
	}
	return jour.After(openingTime) && jour.Before(closingTime)
}

func (r *Restaurant) SetPrice() {
	// met un prix aléatoire sur chaque plate (ajout de 20% pour prendre en compte la taxe Gojek)
	Max := 22.0
	Min := 7.0
	p := Price(rand.Float64()*(Max-Min) + Min)
	r.price = p * 1.2
}

func (r *Restaurant) ReturnID() OsmId {
	return r.Osmid
}

func (r *Restaurant) ReturnFoodType() FoodType {
	return r.Cuisine
}

func (r *Restaurant) ReturnPrice() Price {
	return r.price
}

func (r *Restaurant) ReturnPreparationTime() time.Duration {
	return r.PreparationTime
}

func (r *Restaurant) ReturnSchedules() [2]time.Time {
	return r.schedules
}

func (r *Restaurant) AddWaitingTime() {
	suppTime := <-r.prepTimeWaiter
	r.waitingTime += suppTime
}

func (r *Restaurant) AddPlate() {
	suppTime := <-r.nbPlateWaiter
	r.nbPlates += suppTime
}
func (r *Restaurant) GetPrice() Price {
	return r.price
}

func (r *Restaurant) GetNbPlate() Stock {
	return r.nbPlates
}

func (r *Restaurant) ReturnWaitingTime() time.Duration {
	return r.waitingTime
}

func (r *Restaurant) ChangeStock() {
	val := <-r.stockWaiter
	r.stock += val // marche aussi avec les ints négatifs donc
}

func (r *Restaurant) GetStock() Stock {
	return r.stock
}

func (r *Restaurant) AcceptOrder(ord *Order) bool {
	if !r.stop {
		// dans cette fonction, on décide de récuperer la valeur de stock et waiting a un instant t et de ne pas le lock afin de préserver une idée de réalisme entre ce qui se passe en cuisine et celui qui accepte, qui ne pas toujours connaître l'état des stocks
		// on accepte une order si le stock restant est suffisant pour toutes les commandes et que le waiting ne fait pas dépasser les horaires
		if !r.IsOpen(r.e.ReturnActualTime()) {
			return false
		}

		nbPlates := r.GetNbPlate()

		if r.GetStock() <= nbPlates {
			return false
		}
		prepTime := r.ReturnPreparationTime()
		timePrep := time.Duration(float64(nbPlates+Stock(ord.nbPlate)) * float64(prepTime)) // temps total de preparation

		actualTime := r.e.ReturnActualTime()
		endTime := actualTime.Add(timePrep)
		if !r.IsOpen(endTime) { // si le temps de préparation est après la fermeture
			return false
		}

		// mettre à jour le waiting time du restau et l'id du restau

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.AddWaitingTime()
		}()
		r.prepTimeWaiter <- time.Duration(ord.nbPlate) * prepTime
		wg.Wait()

		r.inProgressOrders.Store(ord.GetId(), ord)

		var wg1 sync.WaitGroup
		wg1.Add(1)
		go func() {
			defer wg1.Done()
			r.AddPlate()
		}()
		r.nbPlateWaiter <- Stock(ord.nbPlate)
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.AddCommandeCounter()
		}()
		r.nbCommandesWaiter <- 1
		wg.Wait()
		ord.SetEstimatedPreparationTime(time.Duration(ord.nbPlate) * prepTime)
		_, estimatedDeliveryTime := r.e.GetShortestPathToNode(r.ClosestNode, ord.Client.Apartment.ClosestNode)
		ord.SetEstimatedDeliveryTime(DeliveryTime(time.Second * time.Duration(estimatedDeliveryTime)))
		//fmt.Println("Estimation de la durée de la livraison", estimatedDeliveryTime)
		ord.SetRestaurant(r)
		r.CalculateScore()
		return true
	} else {
		return false
	}
}

func (r *Restaurant) Start() {
	r.working = true
	for r.working {
		r.Perceive()
		r.Deliberate()
		r.Act()
	}
	fmt.Println(r.Name, " s'arrete")
	r.CancelAllOrder()
	close(r.stockWaiter)
	close(r.nbPlateWaiter)
	close(r.prepTimeWaiter)
}

func (r *Restaurant) Perceive() {
	r.count = r.nbCommandes
}

func (r *Restaurant) Deliberate() {
	if r.action != -2 {
		if r.count == 0 {
			r.action = 0
			time.Sleep(time.Duration(r.e.ReturnAccelerationTime() * float32(time.Minute)))
		} else if r.GetStock() <= 0 {
			r.action = -1
		} else {
			r.action = 1
		}
	} else {
		time.Sleep(time.Duration(r.e.ReturnAccelerationTime() * float32(time.Hour)))
	}
}

func (r *Restaurant) Act() {
	if r.action == 1 {
		order := r.ReturnFirstInProgressOrder()
		nombrePlats := order.GetNbPlate()
		timePrep := time.Duration(float64(nombrePlats) * float64(r.PreparationTime) * float64(r.e.ReturnAccelerationTime()))
		//fmt.Println(r.Name, " commence la préparation de la commande ", order.GetId(), ". Durée estimée:", timePrep)
		g := r.e.ReturnGojeck()
		g.OrderReady.Store(order.GetId(), order)
		timePreparation := time.NewTimer(timePrep)
		<-timePreparation.C // attendre que le timer soit arrivé à expiration

		r.readyOrder.Store(order.GetId(), order)
		//fmt.Println(r.Name, " ", order.GetId(), "ajouté aux commandes pretes")

		var wg sync.WaitGroup

		// mise à jour de l'ensemble des paramètres
		wg.Add(1)
		go func() {
			defer wg.Done()
			// on met à jour le stock. Dans certains cas, on pourra avoir un stock négatif de quelques stocks à la fin de la journée.
			for range nombrePlats - 1 {
				go r.ChangeStock()
				QtStock := rand.Intn(10)
				if QtStock > 7 {
					r.stockWaiter <- -2
				} else {
					r.stockWaiter <- -1
				}
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.AddWaitingTime()
		}()
		r.prepTimeWaiter <- -time.Duration(float64(order.nbPlate) * float64(r.ReturnPreparationTime()))
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.AddCommandeCounter()
		}()
		r.nbCommandesWaiter <- -1
		wg.Wait()

	}
	if r.action == -1 {
		r.CancelAllOrder()
		r.action = -2 // le restaurant ferme
	}
}
func (r *Restaurant) AddCommandeCounter() {
	// permet de compter le nombre de commandes totales
	a := <-r.nbCommandesWaiter
	r.nbCommandes += a
}

func (r *Restaurant) CancelAllOrder() {
	// annule toutes les commandes
	r.inProgressOrders.Range(func(key, value any) bool {
		if val, ok := value.(*Order); ok {
			r.e.ReturnGojeck().CancelOrder(val)
		}
		return true
	})
}
func (r *Restaurant) Populate(a *json.Decoder) error {
	err := a.Decode(r)
	if err != nil {
		fmt.Printf("Erreur lors du décodage : %v\n", err)
		return err
	}
	return nil
}

func (r *Restaurant) GiveOrder(o *Order, d *Deliverer) bool {
	_, val := r.readyOrder.LoadAndDelete(o.GetId())
	if val == false {
		return false // la commande n'existait pas (elle n'était pas sur le desk)
	} else if d.Id == o.Deliverer.Id {
		//fmt.Println(r.Name, "a donnée la commande", o.GetId(), " à ", d.Id)
		return true // retourne le fait que la commande a bien été envoyée (retirée du desk)

	} else {
		return false
	}
}

func (r *Restaurant) SetDuration() {
	//met une valeur de préparation d'un plat de manière aléatoire
	r.PreparationTime = time.Duration(rand.Intn(4)+1)*time.Minute + time.Second*(time.Duration(rand.Intn(59)+1))

}
func (r *Restaurant) SetEnv(e *Environment) {
	r.e = e
}

func (r *Restaurant) Init() {
	r.SetOpening()
	r.SetPrice()
	r.SetDuration()
	r.prepTimeWaiter = make(chan time.Duration)
	r.nbPlateWaiter = make(chan Stock)
	r.stockWaiter = make(chan Stock)
	r.nbCommandesWaiter = make(chan int)
	go r.ChangeStock()
	Min := 10000
	Max := 15000
	randomNumber := rand.Intn(Max-Min+1) + Min
	r.stockWaiter <- Stock(randomNumber)
}

func (r *Restaurant) Stop() {
	r.working = false // indications pour la gui
	r.stop = true

}

func (r *Restaurant) ShowingAttr() {
	fmt.Println("FoodType", r.Cuisine)
	fmt.Println("Prix", r.price)
	fmt.Println("Stock", r.stock)
	fmt.Println("Prep", r.PreparationTime)
	fmt.Println("Ferme", r.schedules[1])
	count := 0
	r.inProgressOrders.Range(func(key, value any) bool {
		count += 1
		return true
	})
	fmt.Println("Nombre de commandes en cours", count)
	fmt.Println("waitingTime", r.ReturnWaitingTime())
}

func (r *Restaurant) CancelOrder(o *Order) {
	r.readyOrder.Delete(o.GetId())
	r.inProgressOrders.Delete(o.GetId())
}
