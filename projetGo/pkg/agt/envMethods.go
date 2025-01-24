package agt

import (
	"fmt"
	"sync"
	"time"
)

type Statistics struct {
	Time                    time.Time
	RunPriceRatio           float32
	MoneyMade               Price
	NumberOfOrder           int
	AverageMoneyMadeByOrder Price
}

// a faire tourner en go routine
func (e *Environment) AddTime(duration int) {
	// permet de gérer l'horloge interne de la simulation
	var wg sync.WaitGroup
	Max := 60 * 60 * duration // durée max en seconde de la simulation
	for i := range Max {
		actualTime := e.ReturnActualTime()
		e.SetActualTime(actualTime.Add(1 * time.Second))
		time.Sleep(time.Duration(e.accelerationTime * float32(1*time.Second))) // la fonction dort 1 seconde
		statPrecision := 60 * 10
		if i%statPrecision == 0 {
			wg.Add(1)
			go func() { // génération des stats en go routine (dans le sync.Map) pour éviter des soucis de déréglement de l'horloge (l'obtention de statistique prend du temps)
				defer wg.Done()
				e.GenerateStat(statPrecision)
			}()
		}
		if (actualTime.Minute() == 30 || actualTime.Minute() == 0) && actualTime.Second() == 0 {
			// affiche l'heure toute les 30 minutes
			fmt.Println("-------- Il est ", actualTime)
		}
	}
	wg.Wait()
}

// Fonction pour obtenir l'argent total généré
func (e *Environment) GetTotalMoneyMade() Price {
	var totalMoneyMade Price
	e.Statistics.Range(func(key, value interface{}) bool {
		stat, ok := value.(*Statistics) // Assurez-vous que la valeur est du bon type
		if ok {
			totalMoneyMade += stat.MoneyMade
		}
		return true // continue l'itération
	})
	return totalMoneyMade
}

// Fonction pour obtenir l'argent total généré par heure
func (e *Environment) GetTotalMoneyMadeByHour() map[int]Price {
	totalMoneyMadeByHour := make(map[int]Price)
	e.Statistics.Range(func(key, value interface{}) bool {
		stat, ok := value.(*Statistics)
		if ok {
			hour := stat.Time.Hour()
			totalMoneyMadeByHour[hour] += stat.MoneyMade
		}
		return true
	})
	return totalMoneyMadeByHour
}

// Fonction pour obtenir le nombre de commandes passées par heure
func (e *Environment) GetOrderMadeByHour() map[int]int {
	totalOrderMadeByHour := make(map[int]int)
	e.Statistics.Range(func(key, value interface{}) bool {
		stat, ok := value.(*Statistics)
		if ok {
			hour := stat.Time.Hour()
			totalOrderMadeByHour[hour] += stat.NumberOfOrder
		}
		return true
	})
	return totalOrderMadeByHour
}

func (e *Environment) GetAveragedMoneyMadeByOrderByHour() map[int]Price {
	var averagedMoneyMadeByOrderByHour = map[int]Price{}
	totalMoneyMadeByHour := e.GetTotalMoneyMadeByHour()
	totalOrderMadeByHour := e.GetOrderMadeByHour()
	for hour, totalMoney := range totalMoneyMadeByHour {
		averagedMoneyMadeByOrderByHour[hour] = totalMoney / Price(totalOrderMadeByHour[hour])
	}
	return averagedMoneyMadeByOrderByHour
}

func (e *Environment) GenerateStat(statPrecision int) {
	actTime := e.actualTime
	timedelta := time.Duration(statPrecision) * time.Second
	beginningTime := actTime.Add(-timedelta)
	var totalMoneyMade Price
	var totalOrderMade int
	var averageMoneyMadeByOrder Price
	var runPriceRatio float32
	e.ReturnGojeck().ReturnOrderHistoric().Range(func(key, value any) bool {
		order := value.(*Order)
		if order.BeginningDeliveryTime.After(beginningTime) {
			totalMoneyMade += order.GetGojekCommission()
			totalOrderMade += 1
		}
		return true
	})
	if totalOrderMade == 0 {
		averageMoneyMadeByOrder = 0
	} else {
		averageMoneyMadeByOrder = totalMoneyMade / Price(totalOrderMade)
	}
	runPriceRatio = e.ReturnGojeck().RunPriceRatio
	newStat := &Statistics{
		Time:                    beginningTime,
		RunPriceRatio:           runPriceRatio,
		MoneyMade:               totalMoneyMade,
		NumberOfOrder:           totalOrderMade,
		AverageMoneyMadeByOrder: averageMoneyMadeByOrder,
	}
	// Updtate total statistics
	e.TotalStatistics.MoneyMade += totalMoneyMade
	e.TotalStatistics.NumberOfOrder += totalOrderMade
	if e.TotalStatistics.NumberOfOrder == 0 {
		e.TotalStatistics.AverageMoneyMadeByOrder = 0
	} else {
		e.TotalStatistics.AverageMoneyMadeByOrder = e.TotalStatistics.MoneyMade / Price(e.TotalStatistics.NumberOfOrder)
	}
	e.Statistics.Store(actTime, newStat)
	e.newStat = true
}

func Simulate(nbLivreur int, nbCustomer int, duration int, e *Environment) {
	// fonction de simulation
	currentDate := time.Now()
	e.SetDuration(duration)
	e.numCustomers = nbCustomer
	e.numDeliverers = nbLivreur
	e.SetActualTime(time.Date(
		currentDate.Year(), currentDate.Month(), currentDate.Day(), // Année, Mois, Jour
		11, 0, 0, 0, // Heure: 15h, Minutes, Secondes, Nanosecondes
		currentDate.Location(), // Fuseau horaire actuel
	))

	// création du Gojek
	goj := e.ReturnGojeck()
	fmt.Println("Starting gojek")
	go goj.Start()

	for range nbCustomer {
		newCustomer := CreateCustomer(e)
		e.Customers = append(e.Customers, newCustomer)
		go func() {
			time.Sleep(1 * time.Second) // pour que les clients soient tous initialisés avant de commencer
			newCustomer.Start()

		}()
	}

	for range nbLivreur {
		newDeliverer := CreateDeliverer(e) // bon apres uniformisation
		go func() {
			time.Sleep(1 * time.Second) // pour que les livreurs soient tous initialisés avant de commencer
			newDeliverer.Start()
		}()
	}
	for _, restaurant := range e.Restaurants {
		restaurant.SetEnv(e)
		restaurant.Activated = true
		restaurant.Init()
		go func() {
			time.Sleep(1 * time.Second) // pour que les restaurants soient tous initialisés avant de commencer
			restaurant.Start()
		}()
	}

	endSimu := make(chan bool)

	go func() {
		e.AddTime(duration) // fonction qui lance l'horloge interne
		endSimu <- true
	}()

	value := <-endSimu
	fmt.Println("Fin de la simulation", value)
	e.EndSimulation()
	fmt.Println(e.GetOrderMadeByHour()) // montre le nombre de commandes faites par heure
	fmt.Println("Heure de fin de la simulation", e.ReturnActualTime())
}

func (e *Environment) ReturnDeliverers() *sync.Map {
	return e.Deliverers
}
func (e *Environment) EndSimulation() {
	/*
		Cette fonction va mettre les bollens working de chacun des agents a false (arreter tout les agents). Alors, les agents vont finir leur action en cours puis s'arreter. Il est possible que certains agents aient des problemes de channel(notamment les clients qui attendent des infos)
	*/
	var wg sync.WaitGroup
	for _, r := range e.Restaurants {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Stop()
		}()
	}
	wg.Wait()

	e.Deliverers.Range(func(key, value any) bool {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value.(*Deliverer).Stop()

		}()
		return true
	})
	wg.Wait()

	time.Sleep(3 * time.Second) // pour que tous les livreurs soient bien arretés

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, elem := range e.Customers {
			go elem.Stop()
		}
	}()
	wg.Wait()

	e.ReturnGojeck().Stop()
}

func (e *Environment) ReturnApartment() []*Apartment {
	return e.Apartments
}
