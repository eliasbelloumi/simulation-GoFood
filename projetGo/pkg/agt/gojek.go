package agt

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/graph/multi"
)

// add context with timeOut pour les commandes qui échouent

type Gojek struct {
	environment          *Environment
	GPS                  *multi.WeightedDirectedGraph
	orderHistoric        sync.Map    //[]Order
	onGoingOrder         sync.Map    // commandes en cours (non finie ou annulées)
	customers            []*Customer // liste des clients
	restaurants          ListRestau  // liste des restaurants
	RunPriceRatio        float32     // prix minimum pour une course, il ne peut pas être inférieur à ce prix
	MaxRunPriceRatio     float32
	maxRunPrice          Price
	DenialRatio          float32
	foodTypeInCity       []FoodType              // type de nourriture dans une villes
	restaurantByFoodType map[FoodType]ListRestau // contient les restaurants pour chaque type de nourriture
	idIncrement          OrderId                 // suivi de l'id pour garantir l'unicité
	IdIncrementChan      chan bool
	count                int
	PreparedOrders       []*Order // commandes pretes en attente de livreurs
	allOrder             sync.Map

	customerDenialRatio float64 // nombre de clients qui refusent de commander

	nOrderWithDeliverer           int
	nOrderWithoutDeliverer        int
	delivererDenialRatio          float32
	customerRatio                 float64 // ratio à l'instant t utilisé pour adapter
	currentOrdersWithoutDeliverer *sync.Map

	OrderRequest             *sync.Map // demande d'un client pour obtenir des restaurants
	OrderTreatment           *sync.Map // demande de commande d'un client (preferences, nombre de plats, id)
	OrderReady               *sync.Map // commandes qui commencent à etre préparée par le restaurant
	OrderAcceptedByDeliverer *sync.Map
	idNumb                   chan OrderId
	normalizeCount           int

	running bool // indique si gojek est actif
}

func (g *Gojek) AddValIdIncrement() {
	for range g.IdIncrementChan {
		g.idNumb <- g.idIncrement
		g.idIncrement += 1

	}
}

func (g *Gojek) ReturnOnGoingOrders() *sync.Map {
	return &g.onGoingOrder
}

func (g *Gojek) ReturnOrderHistoric() *sync.Map {
	return &g.orderHistoric
}
func (g *Gojek) ReturnFoodType() []FoodType {
	return g.foodTypeInCity
}

func CreateGojek(environment *Environment, directedGraph *multi.WeightedDirectedGraph, customers []*Customer, restaurants []*Restaurant, RunPriceRatio float32, maxRunPrice Price, maxRunPriceRatio float32) *Gojek {
	var elem *Gojek
	foodTypeInCity := make([]FoodType, 0)
	listRestau := ListRestau(restaurants)
	elem = &Gojek{RunPriceRatio: RunPriceRatio, environment: environment, GPS: directedGraph, customers: customers, restaurants: listRestau, foodTypeInCity: foodTypeInCity, maxRunPrice: maxRunPrice, MaxRunPriceRatio: maxRunPriceRatio, customerDenialRatio: 1.0}
	// les sync map seront crées vides automatiquement
	elem.OrderTreatment = new(sync.Map)
	elem.OrderRequest = new(sync.Map)
	elem.OrderReady = new(sync.Map)
	elem.OrderAcceptedByDeliverer = new(sync.Map)
	elem.createFoodTypeInCity()
	elem.IdIncrementChan = make(chan bool)
	elem.running = true
	go elem.AddValIdIncrement()
	return elem
}

func (g *Gojek) createFoodTypeInCity() {

	// Utiliser une map pour garantir l'unicité
	uniqueTypes := make(map[FoodType]struct{})
	restaurantByFoodType := make(map[FoodType]ListRestau)

	// Parcourir les Restaurants
	for _, r := range g.restaurants {
		if r != nil { // Vérifier que le pointeur n'est pas nil
			ftype := r.ReturnFoodType()
			restaurantByFoodType[ftype] = append(restaurantByFoodType[ftype], r)
			uniqueTypes[r.ReturnFoodType()] = struct{}{}
		}
	}

	// Convertir la map en une liste
	result := make([]FoodType, 0, len(uniqueTypes))
	for ft := range uniqueTypes {
		result = append(result, ft)
	}

	g.foodTypeInCity = result
	g.restaurantByFoodType = restaurantByFoodType
}

type ListRestau []*Restaurant

// Structure pour contenir les résultats
type Result struct {
	Restaurant *Restaurant
	Score      float64 // price / waiting_time
}

func SortResults(results []Result) {
	sort.Slice(results, func(i, j int) bool {
		// Trier par score de manière décroissante
		return results[i].Score > results[j].Score
	})
}

func (restaurants *ListRestau) Shuffle() {
	r := *restaurants
	for i := len(*restaurants) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		r[i], r[j] = r[j], r[i]
	}
}

// Fonction principale avec goroutines
func (restaurants *ListRestau) filterAndRankRestaurants(cuisineType FoodType, seuil float64, g *Gojek) ListRestau {

	var filteredResults []Result
	// Lancer une goroutine pour chaque restaurant

	for _, restaurant := range g.restaurantByFoodType[cuisineType] {
		score := restaurant.NormalizedScore
		//if score > seuil {

		filteredResults = append(filteredResults, Result{Restaurant: restaurant, Score: score})
		if len(filteredResults) > 50 {

			break
		}
		SortResults(filteredResults)
		//}
	}

	// Collecter les resultats
	SortResults(filteredResults)

	if len(filteredResults) > 30 {
		if seuil > 1 {
			seuil = 1.0
		}
		filteredResults = filteredResults[:int(seuil*float64(len(filteredResults)))]
	}

	var restaurantsObtained ListRestau
	for _, elem := range filteredResults {
		restaurantsObtained = append(restaurantsObtained, elem.Restaurant)
	}

	return restaurantsObtained
}

func (g *Gojek) PrintRestau() ListRestau {
	return g.restaurants
}

// Normaliser les scores pour qu'ils soient dans un intervalle [0, 1] en utilisant des goroutines
func (restaurants ListRestau) NormalizeScores() {
	if len(restaurants) == 0 {
		return
	}

	// Calculer min/max des scores en utilisant des goroutines et un canal
	var minScore, maxScore float64
	minScore, maxScore = restaurants[0].score, restaurants[0].score

	var wg sync.WaitGroup
	// Canal pour envoyer les scores et mettre à jour min/max
	scoreUpdates := make(chan float64)

	// Diviser la tâche de calcul des scores en goroutines
	for i := 0; i < len(restaurants); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Calculer le score pour ce restaurant

			// Envoyer le score pour mise à jour de min/max
			scoreUpdates <- restaurants[i].score
		}(i)
	}

	// Goroutine pour mettre à jour min/max en fonction des scores envoyés
	go func() {
		for score := range scoreUpdates {
			if score < minScore {
				minScore = score
			}
			if score > maxScore {
				maxScore = score
			}
		}
	}()

	// Attendre que toutes les goroutines aient terminé
	wg.Wait()

	// Fermer le canal une fois les mises à jour terminées
	close(scoreUpdates)

	var wg1 sync.WaitGroup
	// Normaliser les scores après avoir calculé min/max
	for _, restau := range restaurants {
		wg1.Add(1)
		go func() {
			defer wg1.Done()
			if maxScore == minScore {
				restau.score = 1 // Si tous les scores sont égaux, définissez tous à 1
			} else {
				restau.NormalizedScore = (restau.score - minScore) / (maxScore - minScore)
			}
		}()
	}
	wg1.Wait()
}

func (g *Gojek) ReturnRestaurants(prefs FoodPreferences) ListRestau {
	results := make(chan ListRestau, len(prefs))
	var wg sync.WaitGroup
	for _, pref := range prefs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resto := g.restaurants
			a := resto.filterAndRankRestaurants(pref, g.customerDenialRatio, g)

			results <- a
		}()

	}
	wg.Wait()
	close(results)
	var allRestau ListRestau
	for element := range results {
		allRestau = append(allRestau, element...)

	}
	return allRestau
}

func (g *Gojek) sortDeliverers(deliverers []*Deliverer, o *Order) []*Deliverer {
	sort.Slice(deliverers, func(i, j int) bool {
		return deliverers[i].ComputeScoreRegardingOrder(o) > deliverers[j].ComputeScoreRegardingOrder(o)
	})
	return deliverers
}

func (g *Gojek) FindDeliverer(ord *Order) {
	var selectedDeliverers []*Deliverer
	g.environment.Deliverers.Range(func(_, value interface{}) bool {
		if v, ok := value.(*Deliverer); ok {
			if v.GetAvailable() {
				selectedDeliverers = append(selectedDeliverers, v)
			}
		}
		return true
	})
	if len(selectedDeliverers) == 0 {
		return
	}
	sortedDeliverers := g.sortDeliverers(selectedDeliverers, ord)
	for _, currentDeliverer := range sortedDeliverers {
		if currentDeliverer.GetAvailable() {
			currentDeliverer.ProposeOrder(ord)
		}
	}
	//fmt.Println("il est ", g.environment.actualTime, "je suis à la sortie de Deliverer")
	return
}
func (g *Gojek) FindDeliverers() {

	var wg sync.WaitGroup

	// Vérification initiale
	if g.GetNElementsInSyncMap(g.OrderReady) == 0 {
		return
	}

	// Parcourir les commandes prêtes
	g.OrderReady.Range(func(k, value interface{}) bool {
		order := value.(*Order)
		order.runPrice = min(g.GetRunPrice(order.GetPrice()), g.maxRunPrice) // Limite à g.maxRunPrice

		// Vérifier l'état et le livreur de la commande
		if order.Deliverer == nil && order.GetState() >= 0 {
			order.Ttl += -1 // Réduire le TTL

			if order.Ttl < 0 && false { // Cas où la commande expire
				order.SetState(-2)
				g.OrderReady.Delete(order.GetId())
				order.Restau.CancelOrder(order)

				fmt.Println("juste devant le channel (bloquant ?????)")
				order.Client.OrderInfos <- order
			} else { // Trouver un livreur
				wg.Add(1)
				go func(order *Order) {
					defer wg.Done()
					g.FindDeliverer(order)
				}(order)
			}
		}
		return true
	})

	// Attendre la fin de toutes les goroutines
	wg.Wait()
}

func (g *Gojek) HandleOrderRequest() {
	var n int8
	n += 1
	var wg sync.WaitGroup
	for _, element := range g.restaurantByFoodType {
		element.Shuffle()
	}
	g.OrderRequest.Range(func(key, value any) bool {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if v, ok := value.(*Customer); ok {
				if v != nil {
					v.PossibleRestaurant <- g.ReturnRestaurants(v.ReturnFoodPrefs()) // retourne les restaurants possibles au client
				}
			}
		}()
		g.OrderRequest.Delete(key) // Supprimer la clé courante
		if n < 100 {
			return true // Continuer l'itération
		} else {
			return false
		}
	})

	wg.Wait()
}

func (g *Gojek) HandleOrderTreatment() {
	// gestion des demandes des clients une fois leur préférences emises
	g.idNumb = make(chan OrderId)
	nbEchec := 0
	n := 0

	g.OrderTreatment.Range(func(key, value any) bool {
		// pour chaque demande
		n += 1
		if v, ok := value.(OrderStructRequest); ok {
			indic := false
			if len(v.Preferences) > 0 { // si le client a indiqué des préférences
				ord := Order{Id: g.idIncrement, state: 0, nbPlate: v.NbPlat, Client: v.Customer, Ttl: 11, BeginningOrdering: g.environment.ReturnActualTime()}
				g.idIncrement += 1
				g.allOrder.Store(ord.GetId(), ord)
				for _, restaurant := range v.Preferences {
					a := restaurant.AcceptOrder(&ord) // on demande à chaque restaurant si il accepte la commande
					//fmt.Println(a)
					if a {
						ord.SetPrice(Price(v.NbPlat) * restaurant.ReturnPrice()) // met à jour le prix de la commande
						ord.SetState(1)
						g.onGoingOrder.Store(ord.GetId(), ord)
						indic = true
						break
					}

				}
				if !indic {
					ord.SetState(-1)
					nbEchec += 1
				}

				if !v.Customer.Sleeping { // le client ne s'est pas arrété, on lui envoie les informations
					v.Customer.OrderInfos <- &ord
				}
			} else {
				nbEchec += 1
			}
		}
		g.OrderTreatment.Delete(key) // Supprimer la clé courante
		if n < 80 {
			return true
		} else {
			return false
		}
	})
	if n != 0 {
		ratio := float64(nbEchec) / float64(n)
		fmt.Println("traitement de ", n, " commandes,", nbEchec, " Echecs")
		g.customerRatio = ratio
	} else {
		g.customerRatio = -1.0
	}
}

func (g *Gojek) AssignOrderToDeliverer(o *Order, d *Deliverer) bool {
	var wasSet bool
	_, wasTaken := g.OrderReady.LoadAndDelete(o.GetId())
	if wasTaken {
	}
	if wasTaken { // si la commande a bien été attribuée
		wasSet = o.SetDeliverer(d)
		if wasSet {
			d.currentOrder = o
			g.OrderAcceptedByDeliverer.Store(o.GetId(), o)
			o.SetBeginningDeliveryTime(g.environment.actualTime)
		}
	}
	return wasSet
}

func (g *Gojek) CloseOrder(o *Order) {
	o.SetState(6)
	g.OrderAcceptedByDeliverer.Delete(o.Id)
	g.onGoingOrder.Delete(o.Id)
	g.orderHistoric.Store(o.Id, o)
	if !o.Client.looking {
		o.Client.OrderInfos <- o
	}
}

func (g *Gojek) GetNElementsInSyncMap(sm *sync.Map) int {
	elements := 0
	sm.Range(func(_, _ interface{}) bool {
		elements++
		return true // Continue iteration
	})
	return elements
}

func (g *Gojek) GetNOrdersWithDeliverer(previousOrderReady *sync.Map) int {
	nOrdersWithDeliverer := 0
	g.OrderAcceptedByDeliverer.Range(func(k, _ interface{}) bool {
		_, isIn := previousOrderReady.Load(k.(OrderId))
		if isIn { // on ne compte que les commandes qui sont dans la liste des commandes concernées (celle sans livreurs au début de l'itération)
			nOrdersWithDeliverer++
		}
		return true // Continue iteration
	})
	return nOrdersWithDeliverer
}

func (g *Gojek) PerceiveOrderDelivery() {
	g.currentOrdersWithoutDeliverer = g.CopySyncMap(g.OrderReady)
	g.nOrderWithoutDeliverer = g.GetNElementsInSyncMap(g.currentOrdersWithoutDeliverer)
}

func (g *Gojek) DeliberateOrderDelivery() { // pas grave si elle est vide
}

func (g *Gojek) ActOrderDelivery() {
	g.FindDeliverers()
	// on attend un peu pour laisser le temps aux livreurs de répondre, sinon gojek augmente son seuil beaucoup trop vite
	// c'est pas grave si elle n'ont pas toutes été analysées :  ça veut dire qu'il n'y a pas assez de livreur
	// -> plus de demande que d'offre -> ça reste logique d'augmenter le prix
	g.nOrderWithDeliverer = g.GetNOrdersWithDeliverer(g.currentOrdersWithoutDeliverer)
	g.UpdateRunPriceRatio(g.GetDelivererDenialRatio())

}

func (g *Gojek) GetDelivererDenialRatio() float32 {
	if g.nOrderWithoutDeliverer == 0 {
		return -1
	} else {
		return 1 - (float32(g.nOrderWithDeliverer) / float32(g.nOrderWithoutDeliverer))
	}
}

func (g *Gojek) UpdateRunPriceRatio(denialRatio float32) {
	if denialRatio == -1 { // code pour ne pas changer le ratio, appliqué lors de la première itération, ou quand il n'y a pas de commande
		return
	} else if denialRatio == 0 { // si il n'y a pas eu de refus
		g.RunPriceRatio = g.RunPriceRatio * 0.95                           // on baisse le prix de 5 %
		g.RunPriceRatio = float32(math.Max(0.1, float64(g.RunPriceRatio))) // on ne peut pas aller en dessous de 0.1 fois le prix de la commande
	} else {
		g.RunPriceRatio = g.RunPriceRatio * float32(math.Sqrt(float64(1+denialRatio)))             // on augmente le prix par le ratio de refus
		g.RunPriceRatio = float32(math.Min(float64(g.MaxRunPriceRatio), float64(g.RunPriceRatio))) // on ne peut pas aller au dessus de d'un seuil
	}
}

func (g *Gojek) GetRunPrice(price Price) Price {
	return price * Price(g.RunPriceRatio) // l'idée ici est que gojeck essaye d'approcher une fonction complexe (celle du livreur)
	// par une fonction très simple celle ci, ce qui l'ammène à tatonner
}

func (g *Gojek) DeliberateForCustomer() {
	// traitement des commandes
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		g.HandleOrderTreatment()
	}()
	go func() {
		defer wg.Done()
		g.HandleOrderRequest()
	}()
	wg.Wait()
}

func (g *Gojek) updateCustomerDenialRatio() {
	defer func() {
		if g.customerDenialRatio < 0.5 {
			g.customerDenialRatio = 0.5
		}
	}()
	if g.customerRatio >= 0 {
		if g.customerRatio == 0 {
			g.customerDenialRatio -= 0.08
			return
		} else if g.customerRatio < 0.2 {
			g.customerDenialRatio += 0.05
			return
		} else if g.customerRatio < 0.3 {
			g.customerDenialRatio += 0.2
			return
		} else if g.customerRatio > 0.5 {
			g.customerDenialRatio += 0.4
			return
		}
	}
}

func (g *Gojek) CopySyncMap(original *sync.Map) *sync.Map {
	copied := &sync.Map{}
	original.Range(func(key, value any) bool {
		copied.Store(key, value)
		return true // Continue iteration
	})
	return copied
}

func (g *Gojek) Start() {
	for g.running {
		g.Perceive()
		//fmt.Println("Perceive fini")
		g.Deliberate()
		//fmt.Println("Deliberate fini")
		g.Act()
		time.Sleep(time.Duration(float32(time.Minute) * g.environment.ReturnAccelerationTime()))
	}
	close(g.IdIncrementChan)
	close(g.idNumb)
}
func (g *Gojek) Perceive() {
	//fmt.Println("Perceive pour livreur")
	g.PerceiveOrderDelivery()
}

func (g *Gojek) Act() {
	g.ActOrderDelivery()
	g.updateCustomerDenialRatio()
}
func (g *Gojek) Deliberate() {
	if g.normalizeCount == 10 {
		g.restaurants.NormalizeScores()
		g.normalizeCount = 0
		fmt.Println("Ratio de refus des customer", g.customerDenialRatio)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.DeliberateOrderDelivery()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.DeliberateForCustomer()
	}()
	wg.Wait()
	g.normalizeCount += 1

}

func (g *Gojek) Stop() {
	g.running = false
}

func (e *Gojek) CancelOrder(o *Order) {
	// annule une commande. La commande est passée à -2 et on informe le client
	o.SetState(-2)
	e.orderHistoric.Store(o.GetId(), o)
	if !o.Client.looking {
		o.Client.OrderInfos <- o
	}
}
