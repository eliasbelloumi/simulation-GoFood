package agt

import (
	"sync"
	"time"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"

	"gonum.org/v1/gonum/graph/multi"
)

type Environment struct { // un seul agent gojek
	startTime              time.Time // heure de début de la simulation
	duration               int       // durée en heures de la simulation
	numCustomers           int       // nombre de clients
	numDeliverers          int       // nombre de livreurs
	actualTime             time.Time // heure actuelle
	Graph                  *Graph
	GPS                    *multi.WeightedDirectedGraph
	Restaurants            ListRestau // ensemble de restaurant. Type spécial car des méthodes ont été crées
	Apartments             []*Apartment
	Deliverers             *sync.Map // liste des livreurs qui resiste à la concurrence
	Customers              []*Customer
	gojeck                 *Gojek
	accelerationTime       float32
	IncrementalDelivererId DelivererId
	nbClients              int // gestion de l'incrémentation de l'id client

	Statistics      sync.Map // contient les statistiques
	TotalStatistics Statistics
	newStat         bool
}

func (e *Environment) ReturnActualTime() time.Time {
	return e.actualTime
}

func (e *Environment) ReturnCustomerId() int {
	e.nbClients += 1
	return e.nbClients - 1
}
func (e *Environment) ReturnAccelerationTime() float32 {
	return e.accelerationTime
}

func (e *Environment) SetActualTime(newTime time.Time) {
	e.actualTime = newTime
}

func (e *Environment) SetDuration(duration int) {
	e.duration = duration
}

func (e *Environment) GetTimeLeftInWorkingDay() time.Duration {
	// obtention du temps restant dans une journée de travail pour les livreurs
	// Obtenir l'heure actuelle
	now := e.actualTime

	// Créer une heure pour 2 heures du matin aujourd'hui
	twoAM := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())

	// Si l'heure actuelle est après 2h du matin, calculer pour le lendemain
	if now.After(twoAM) {
		twoAM = twoAM.Add(24 * time.Hour)
	}

	// Calculer la différence
	duration := time.Until(twoAM)

	return duration
}

func (e *Environment) Return1Restau() *Restaurant {
	return e.Restaurants[0]
}

func (e *Environment) ReturnGojeck() *Gojek {
	return e.gojeck
}

func (e *Environment) SetNumberRestaurantNearRestau(radius int16) {
	var wg sync.WaitGroup
	for _, r := range e.Restaurants {
		communicate := make(chan bool, 1)
		go func() {
			for {
				// Écoute des messages sur le channel
				_, ok := <-communicate
				if !ok {
					return
				}
				// Traitement du message
				r.NumberRestaurantsNear += 1
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			currentRestaurantPosition := r.ClosestNode
			for _, r2 := range e.Restaurants {
				dist := e.Graph.GetDistanceBetweenNodes(currentRestaurantPosition, r2.ClosestNode)
				if dist < float64(radius) {
					communicate <- true
				}
			}
			//close(communicate)
		}()
	}
	wg.Wait()
}

func (e *Environment) SetNumberRestaurantNearNode(radius int16) {
	var wg sync.WaitGroup

	for _, node := range e.Graph.Nodes {
		communicate := make(chan bool, 1)
		go func() {
			for {
				// Écoute des messages sur le channel
				_, ok := <-communicate
				if !ok {
					return
				}
				// Traitement du message
				node.NumberRestaurantsNear += 1
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, r := range e.Restaurants {
				dist := e.Graph.GetDistanceBetweenNodes(node.Id, r.ClosestNode)
				if dist < float64(radius) {
					communicate <- true
				}
			}
			close(communicate)
		}()
	}
	wg.Wait()
}

func NewEnvironment(graphJson string, apartmentsJson string, retaurantsJson string, nearRestaurantRadius int16, runPriceRatio float32, acceleration float32, maxRunPrice Price, maxRunPriceRatio float32) *Environment {
	gpsChan := make(chan *multi.WeightedDirectedGraph)
	graphChan := make(chan *Graph)
	// Load graph
	go LoadGraphAndGPSfromJSON(graphJson, gpsChan, graphChan)
	GPS := <-gpsChan
	parsedGraph := <-graphChan
	// Load Restaurants
	restaurantsChan := make(chan []*Restaurant)
	go ParseRestaurants(retaurantsJson, restaurantsChan)
	restaurants := <-restaurantsChan
	// Load apartments
	apartmentsChan := make(chan []*Apartment)
	go ParseAppartements(apartmentsJson, apartmentsChan)
	apartments := <-apartmentsChan
	// Create deliverers
	deliverers := new(sync.Map)
	// Create customers
	customers := make([]*Customer, 0)
	//create environment
	newEnv := &Environment{Graph: parsedGraph, GPS: GPS, Restaurants: restaurants, Apartments: apartments, Deliverers: deliverers, accelerationTime: acceleration, startTime: time.Now(), IncrementalDelivererId: 0, nbClients: 0, newStat: false, TotalStatistics: Statistics{}}
	// Create Gojek
	gojeck := CreateGojek(newEnv, GPS, customers, restaurants, runPriceRatio, maxRunPrice, maxRunPriceRatio)
	newEnv.gojeck = gojeck
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		newEnv.SetNumberRestaurantNearRestau(nearRestaurantRadius)
	}()
	wg.Add(1)

	go func() {
		defer wg.Done()
		newEnv.SetNumberRestaurantNearNode(nearRestaurantRadius)
	}()
	wg.Wait()
	return newEnv
}

func (e *Environment) GetShortestPathToNode(fromNode NodeId, toNode NodeId) ([]graph.Node, TravelTime) {
	from := e.GPS.Node(int64(fromNode))
	to := e.GPS.Node(int64(toNode))
	aStarPath, _ := path.AStar(from, to, e.GPS, nil) // Maybe try different algorithms
	itinerary, _ := aStarPath.To(to.ID())
	expectedTravelTime := aStarPath.WeightTo(int64(toNode))
	return itinerary, TravelTime(expectedTravelTime)
}

func (e *Environment) getIncrementalDelivererId() DelivererId {
	val := e.IncrementalDelivererId
	e.IncrementalDelivererId += 1
	return val

}
