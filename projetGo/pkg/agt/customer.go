package agt

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"golang.org/x/exp/rand"
)

type Customer struct {
	Id                 int             // identifiant du client
	Environment        *Environment    // pointeur sur son environnement
	Apartment          *Apartment      // localisation
	hungryLevel        float32         // niveau de faim
	foodPreferences    FoodPreferences // préférences de nourriture
	orderHistoric      sync.Map        // historique des commandes
	wantsToOrder       bool            // booleen permettant au client de commander ou non
	nMealToOrder       int8            // nombre de plats que le client souhaite commander
	previousTimeStamp  time.Time       // derniere heure de perceive
	PossibleRestaurant chan ListRestau // propositions de restaurants par gojek
	OrderInfos         chan *Order     // communication de la commande par gojek
	looking            bool            // informe la gui si le client est  actif
	Sleeping           bool            // indique à gojek si le client a encore ses channels ouverts pour communication
	endOfDigestionTime time.Time
	isDigesting        bool
}

var hungryLevelByHour = map[int]float64{ // probabilité de commander en fonction de l'heure
	7:  0.0,
	8:  0.1,
	9:  0.1,
	10: 0.1,
	11: 0.2,
	12: 0.7,
	13: 0.7,
	14: 0.3,
	15: 0.2,
	16: 0.2,
	17: 0.1,
	18: 0.3,
	19: 0.4,
	20: 0.8,
	21: 0.7,
	22: 0.6,
	23: 0.6,
	0:  0.3,
	1:  0.2,
	2:  0.1,
	3:  0.0,
	4:  0.0,
	5:  0.0,
	6:  0.0,
}

// méthode pour incrémenter le niveau de faim
func (c *Customer) incrementHungryLevel(i float32) {
	c.hungryLevel += i
}

func shuffleFoodPref(slice FoodPreferences) FoodPreferences {
	// Allouer un nouveau slice avec la même longueur que le slice d'entrée
	a := make(FoodPreferences, len(slice))
	copy(a, slice) // Copier le contenu de slice dans a

	// Mélanger les éléments dans le slice a
	for i := len(a) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
	return a
}

// fonction qui génére les préférences de l'agent, utilisé pour les petites villes où il y a peu de restaurants
func (c *Customer) CreatePrefs(listing []FoodType) []FoodType {
	/*
		var selected [5]FoodType
		cmpt := 0
		for cmpt < 5 {
			// Génère un indice aléatoire
			index := rand.Intn(len(listing))

			// Ajoute l'élément à la slice sélectionnée
			selected[cmpt] = listing[index]
			cmpt++

			// supprime l'élément du slice pour éviter les répétitions
			listing = append(listing[:index], listing[index+1:]...)
		}

	*/

	c.foodPreferences = FoodPreferences(listing)
	return listing
}

// fonction qui génére les préférences de l'agent à partir de ce que Gojek propose
func (c *Customer) CreatePrefsInBigCity(listing []FoodType) []FoodType {
	selected := make(FoodPreferences, 5)
	cmpt := 0
	for cmpt < 5 {
		// Génère un indice aléatoire
		index := rand.Intn(len(listing))

		// Ajoute l'élément à la slice sélectionnée
		selected[cmpt] = listing[index]
		cmpt++

		// supprime l'élément du slice pour éviter les répétitions
		listing = append(listing[:index], listing[index+1:]...)
	}

	c.foodPreferences = selected
	return listing
}

func (c *Customer) ReturnFoodTypeRanking(f FoodType) int {
	// retourne l'index d'un food type dans les préférences de l'agent
	for i, elem := range c.foodPreferences {
		if elem == f {
			return i
		}
	}
	return -1
}

// fonction qui emet les préferences à gojek dans l'optique de commander
func (c *Customer) EmitPref(proposals ListRestau, nPlates int8) ListRestau {
	var listing restaurantToSelect
	for _, restaurants := range proposals {
		// itére sur l'ensemble des propositions de gojek
		score := c.ComputeRestaurantScore(restaurants, nPlates)
		elem := restaurantToOrder{
			value: float32(score),
			id:    restaurants,
		}
		listing = append(listing, elem)
	}
	sort.Sort(listing) // ordonnancement de la liste
	var filteredListing restaurantToSelect
	for _, r := range listing {
		hungerLevel := c.hungryLevel
		probRegardingHour := hungryLevelByHour[c.Environment.actualTime.Hour()]
		tolerance := 1 - (1*hungerLevel + 1.5*float32(probRegardingHour))
		if r.value > tolerance { // max hunger level is 14.4 -> now between 0 and 1
			// remove all restaurants with a score lower than the hungry level
			filteredListing = append(filteredListing, r)
		}
	}
	choices := filteredListing.extractNames() // extrait les pointeurs vers les restaurants
	return choices
}

func CreateCustomer(env *Environment) *Customer {
	custo := Customer{Id: env.ReturnCustomerId(), previousTimeStamp: env.ReturnActualTime()} // client sans commande en cours
	custo.Environment = env
	custo.hungryLevel = 0                                                          // initalisation à 0 du niveau de faim
	custo.Apartment = env.ReturnApartment()[rand.Intn(len(env.ReturnApartment()))] // affectation à un appartemment aléatoire
	foodTypes := env.ReturnGojeck().ReturnFoodType()
	copyfoodTypes := append([]FoodType{}, foodTypes...)
	custo.CreatePrefs(copyfoodTypes) // génération des food types

	// création des différents channels
	custo.PossibleRestaurant = make(chan ListRestau, 1)
	custo.OrderInfos = make(chan *Order, 1)
	return &custo
}

func (c *Customer) Perceive() {
	datetime := c.Environment.ReturnActualTime()
	delta := datetime.Sub(c.previousTimeStamp)
	coef := delta.Minutes()

	/*a := rand.Intn(100 + 1)
	for range int(coef + 1) {
		if a > 90 {
			c.incrementHungryLevel(0.25)
			return
		} else if a > 80 {
			c.incrementHungryLevel(0.16)
			return
		} else if a > 40 {
			c.incrementHungryLevel(0.13)
			return
		} else if a > 30 {
			c.incrementHungryLevel(0.05)
			return
		} else if a > 20 {
			c.incrementHungryLevel(0.09)
			return
		} else {
			c.incrementHungryLevel(0.15)
			return
		}
	}*/

	// On recommence à avoir faim 2h après avoir mangé. Ici, on percoit l'heure actuelle et met à jour la croyance liée au niveau de faim
	if c.endOfDigestionTime.Sub(c.Environment.ReturnActualTime()) < 2*time.Hour {
		for range int(coef) + 1 {
			// we are supposed to eat every 8 ~ 10 hours
			increment := 1 / 60 * (10 - rand.Intn(2)) // sum up to 1 every 8 ~ 10 hours
			c.incrementHungryLevel(float32(increment))
			c.previousTimeStamp = datetime
		}
	}

}

func (c *Customer) Deliberate() {
	//get current hour
	probRegardingHour := hungryLevelByHour[c.Environment.actualTime.Hour()-1]

	hungerLevel := float64(c.hungryLevel)

	// Combine hunger and time probabilities (weighted sum)
	weightedProbability := hungerLevel + probRegardingHour // à voir si c'est pas trop faible

	randomValue := rand.Float64() + 0.05

	// Decision: order if the weighted probability is greater than the random number
	c.wantsToOrder = hungerLevel+probRegardingHour > randomValue
	//Convert weighted probability to nbPlate : 0.0 - 0.2 : 1 Plate ; 0.2 - 0.4 : 2 Plates ; 0.4 - 0.6 : 3 Plates ; 0.6 - 0.8 : 4 Plates ; 0.8 - 1.0 : 5 Plates
	c.nMealToOrder = int8(math.Floor(weightedProbability*4.0) + 1.0)
}

func (c *Customer) ReturnFoodPrefs() FoodPreferences {
	return c.foodPreferences
}

func (c *Customer) Order() {
	goj := c.Environment.ReturnGojeck()

	goj.OrderRequest.Store(c, c) // envoi d'une demande de commande à Gojek

	// attente d'une réponse de Gojek
	proposedRestaurants := <-c.PossibleRestaurant

	// emmisions de préférences
	nMealToOrder := c.nMealToOrder
	a := c.EmitPref(proposedRestaurants, nMealToOrder)
	if len(a) == 0 { // si aucun restaurant ne convient
		a = nil
	}

	// on envoie à gojek une demande de commande
	goj.OrderTreatment.Store(c, OrderStructRequest{Preferences: a, NbPlat: nMealToOrder, Customer: c})
	if a != nil {
		//ch := make(chan bool)

		// Create a context with a timeout of 5 seconds
		//ctxTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Minute*time.Duration(c.Environment.ReturnAccelerationTime()))
		//defer cancel()

		// Start the doSomething function
		//go func() {

		valOrd, ok := <-c.OrderInfos // attente d'une réponse de gojek
		if !ok {
			fmt.Println("channel fermé")
			return
		}
		if valOrd.GetState() == -1 {
			// si aucun restaurant n'a été trouvé
			fmt.Println("Ma commande a été cancel")
			c.wantsToOrder = false
			c.nMealToOrder = 0
			return
		}

		// sinon on continue
		valOrd = <-c.OrderInfos
		state := valOrd.GetState()
		if state == 6 { // si la commande a bien été réceptionnée
			c.wantsToOrder = false
			c.nMealToOrder = 0
			c.hungryLevel = 0.0
			fmt.Println("Le client", c.Id, "a bien recu sa commande. Elle est arrivée en ", c.Environment.ReturnActualTime().Sub(valOrd.BeginningOrdering)) // affichage du temps nécessaire pour obtenir la commande
			// cool down de digestion
			c.isDigesting = true
			c.endOfDigestionTime = c.Environment.ReturnActualTime().Add(4 * time.Hour)
			c.previousTimeStamp = c.Environment.ReturnActualTime()
			return
		} else if state == -2 { // si aucun livreur n'était disponible, la commande est annulé
			fmt.Println("Ma commande n'a pas trouvée de livreur")
			c.wantsToOrder = false
			c.nMealToOrder = 0
			return
		} else {
			return
		}
		/*}
		select {
		case <-ctxTimeout.Done():
			close(c.OrderInfos)
			return
			//fmt.Printf("Context cancelled pour obtenir la commande: %v\n", ctxTimeout.Err())
		case result := <-ch:
			if result == true {
			}
		}
		*/
		//close(c.OrderInfos)

	} else {
		//y := 1.0 * c.Environment.ReturnAccelerationTime()
		c.wantsToOrder = false
		c.nMealToOrder = 0
		//a := time.Duration(y * float32(time.Minute))
		//time.Sleep(a)
		return

	}
}

func (c *Customer) Act() {
	if c.wantsToOrder && !c.isDigesting && c.nMealToOrder > 0 {
		c.Order()
	} else if c.isDigesting && c.endOfDigestionTime.Before(c.Environment.ReturnActualTime()) {
		c.isDigesting = false
	}
}
func (c *Customer) Start() {
	c.looking = true
	for c.looking {
		c.Perceive()
		c.Deliberate()
		c.Act()
		time.Sleep(time.Duration(2 * float32(time.Minute) * c.Environment.ReturnAccelerationTime()))
	}

	// une fois que le client a fini ses actions en cours, il va fermer ses channels

	//close(c.PossibleRestaurant)
	close(c.OrderInfos)
	c.Sleeping = true
	fmt.Println(c.Id, " supprimé et ressources nettoyées")

}

func (c *Customer) PickUpOrder(o *Order) bool {
	// récupération d'une commande par le client
	_, itemLoaded := c.orderHistoric.Load(o.GetId())
	if itemLoaded == false { // si la commande n'a pas encore été récupérée
		c.orderHistoric.Store(o.GetId(), o)
		c.RateOrder(o)
		return true
	} else {
		return false // si la commande a déjà été recupérée
	}
}

func (c *Customer) RateOrder(o *Order) {
	delay := o.GetDeliveryDelay()
	var lateFactor float64
	if delay <= 0 {
		lateFactor = 0
	} else {
		lateFactor = float64(delay) * 2 / float64(o.EstimatedDeliveryTime)
	}
	rating := 5 - lateFactor
	o.Deliverer.RateDeliverer(rating)
}

func (c *Customer) ComputeRestaurantScore(r *Restaurant, nPlate int8) float64 {
	//calculer le score des restaurants, fonction des préférences du client, du prix et du temps d'attente
	foodPrefLength := len(c.foodPreferences)
	foodTypeRank := c.ReturnFoodTypeRanking(r.ReturnFoodType())
	price := r.ReturnPrice() * Price(nPlate)
	waitingTime := r.ReturnPreparationTime().Minutes()
	//standardisation des valeurs
	standardizedPrice := float64(price / 22)
	standardizedFoodTypeRank := float64(foodPrefLength-foodTypeRank) / 5
	standardizedWaitingTime := float64(waitingTime) / 12
	standerdizedDistance := c.Environment.Graph.GetDistanceBetweenNodes(c.Apartment.ClosestNode, r.ClosestNode) / 10000 // 10 km max
	//calcul du score
	var score float64
	if foodTypeRank == -1 {
		score = 0
	} else {
		score = standardizedFoodTypeRank - standardizedPrice*standardizedWaitingTime*standerdizedDistance
	}
	return math.Max(score, 0)
}

func (c *Customer) Stop() {
	// Fermer les canaux pour éviter des fuites de ressources
	c.looking = false
	fmt.Println("Dès que la tâche en cours est finie,", c.Id, "s'arrête")

	// Afficher un message de suppression

}
