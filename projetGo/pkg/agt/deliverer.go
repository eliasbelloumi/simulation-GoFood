package agt

import (
	"context"
	"math"
	"sort"
	"sync"
	"time"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/multi"
)

type Deliverer struct {
	Id  DelivererId
	env *Environment

	positionWaiter  chan NodeId
	Position        NodeId
	isMoving        bool
	currentPaths    map[string]Path
	currentPathType string

	GPS   *multi.WeightedDirectedGraph
	Graph *Graph
	Gojek *Gojek

	Available       bool
	availableWaiter chan bool

	DailyGoal Price

	RunPriceThreshold Price

	MoneyMadeToday  Price
	moneyMadeWaiter chan Price

	MinRunPrice Price // doit être une valeur aléatoire, autour du minRunPrice de gojek

	nOrdersWaiter chan int32
	rating        float64
	ratings       []float64

	OrderProposals *sync.Map
	nOrderProposal int8
	currentOrder   *Order

	running       bool
	RecentlyMoved bool
}

func (d *Deliverer) Start() {
	for d.running {
		d.Perceive()
		d.Deliberate()
		d.Act()
		if d.running == false {
			break
		}
	}
	close(d.positionWaiter)
	close(d.availableWaiter)
	close(d.moneyMadeWaiter)
	close(d.nOrdersWaiter)
}

func (d *Deliverer) Perceive() { // j'ai une demande de commande en cours ou pas
	nOrderProposals := d.Gojek.GetNElementsInSyncMap(d.OrderProposals)
	d.nOrderProposal = int8(nOrderProposals)
}

func (d *Deliverer) Deliberate() { // j'accepte cette commande ou pas ? / cette zone me convient ?
	if d.nOrderProposal > 0 {
		timeLeftInWorkingDay := d.env.GetTimeLeftInWorkingDay()
		freezedOrders := d.Gojek.CopySyncMap(d.OrderProposals)
		d.ResetOrderProposals()
		freezedOrders.Range(func(key, value any) bool {
			o := value.(*Order)
			if o.GetState() >= 0 && o.Deliverer == nil {
				if d.AcceptOrder(o, timeLeftInWorkingDay) {
					d.currentOrder = o
				}
			}
			return true
		})
	} else {
		time.Sleep(time.Duration(5 * float32(time.Minute) * d.env.ReturnAccelerationTime()))
	}
}

func (d *Deliverer) Act() {
	if d.currentOrder != nil {
		hasTakenOrder := d.TakeOrder(d.currentOrder)
		if hasTakenOrder {
			d.HandleOrder(d.currentOrder)
		} else {
			d.currentOrder = nil
			d.ResetOrderProposals()
		}
	} else if !d.RecentlyMoved {
		d.MoveToBestArea(1000)
		time.Sleep(time.Duration(4 * float32(time.Minute) * d.env.ReturnAccelerationTime()))
		go func() {
			time.Sleep(time.Duration(5 * float32(time.Minute) * d.env.ReturnAccelerationTime()))
			d.RecentlyMoved = false
		}()

	}
}

func (d *Deliverer) Stop() {
	d.running = false
}

func (d *Deliverer) TakeOrder(o *Order) bool {
	return d.Gojek.AssignOrderToDeliverer(o, d)
}

func (d *Deliverer) HandleOrder(o *Order) {
	d.SetAvailable(false)
	d.currentPathType = "toRestau"
	d.MoveToNode(o.Restau.ClosestNode)
	ctx, cancel := context.WithTimeout(context.Background(), 3*o.Restau.PreparationTime)
	defer cancel()
	success := d.TakeOrderAtRestaurant(o, ctx)
	if !success {
		o.SetState(-2)
		o.Client.OrderInfos <- o
		return
	}
	d.currentPathType = "toClient"
	d.MoveToNode(o.Client.Apartment.ClosestNode)
	d.setActualDeliveryTime(o)
	d.GiveOrder(o)
	d.Gojek.CloseOrder(o) // go routines car bloquant
	d.AddMoneyMadeToday(o.runPrice)
	d.SetAvailable(true)
	d.currentOrder = nil
}
func (d *Deliverer) MoveToBestArea(radius int16) {
	bestNode := d.Graph.GetNodeWithMostRestaurantsNear(d.Position, radius)
	if bestNode.Id != d.Position {
		d.currentPaths["replacement"] = Path{}
		d.currentPaths["toRestau"] = Path{}
		d.currentPaths["toClient"] = Path{}
		d.currentPathType = "replacement"
		d.MoveToNode(bestNode.Id)
	}
}

func (d *Deliverer) ProposeOrder(o *Order) {
	if !d.GetAvailable() {
		return
	}
	d.OrderProposals.Store(o.Id, o)
}

func (d *Deliverer) ResetOrderProposals() {
	d.OrderProposals = new(sync.Map)
}

func (d *Deliverer) GetPosition() NodeId {
	return d.Position
}
func (d *Deliverer) SetPosition(p NodeId) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		pos := <-d.positionWaiter
		d.Position = pos
	}()
	d.positionWaiter <- p
	wg.Wait()
}

func (d *Deliverer) GetMoneyMadeToday() Price {
	return d.MoneyMadeToday
}

func (d *Deliverer) AddMoneyMadeToday(money Price) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		moneyToAdd := <-d.moneyMadeWaiter
		d.MoneyMadeToday += moneyToAdd
	}()
	d.moneyMadeWaiter <- money
	wg.Wait()
}

func (d *Deliverer) RateDeliverer(n float64) {
	d.ratings = append(d.ratings, n)
	var sum float64
	for _, r := range d.ratings {
		sum += r
	}
	d.rating = sum / float64(len(d.ratings))
}

func (d *Deliverer) GetRating() float64 {
	return d.rating
}

func (d *Deliverer) GetDailyGoal() Price {
	return d.DailyGoal
}

func (d *Deliverer) SetDailyGoal(dg Price) {
	d.DailyGoal = dg
}

func (d *Deliverer) GetAvailable() bool {
	return d.Available
}
func (d *Deliverer) SetAvailable(av bool) {
	d.Available = av
}

func CreateDeliverer(env *Environment) *Deliverer {
	d := new(Deliverer)
	d.running = true
	d.Id = env.getIncrementalDelivererId()
	d.Position = env.Graph.GetRandomNode().Id
	d.GPS = env.GPS
	d.Graph = env.Graph
	d.env = env
	d.Available = true
	d.Gojek = env.gojeck
	d.isMoving = false
	d.currentPaths = make(map[string]Path)
	d.currentPaths["replacement"] = Path{}
	d.currentPaths["toRestau"] = Path{}
	d.currentPaths["toClient"] = Path{}
	d.currentPathType = "replacement"
	d.MinRunPrice = 3
	d.MoneyMadeToday = 0
	//random value between 50 and 100$
	d.DailyGoal = Price(50 + rand.Intn(50))
	d.SetAvailable(true)
	// All deliverers start with a rating of 5
	d.rating = 5
	d.ratings = []float64{}
	d.positionWaiter = make(chan NodeId)
	d.availableWaiter = make(chan bool)
	d.moneyMadeWaiter = make(chan Price)
	d.nOrdersWaiter = make(chan int32)
	d.OrderProposals = new(sync.Map)
	d.env.Deliverers.Store(d.Id, d)
	return d
}

func (d *Deliverer) GetHyptotheticalRunPrice(totalDistance float64, nbPlate int8) Price { // Ce que livreur consière comme le prix idéale d'une course
	bonusDistance := totalDistance - 200        // 100 meters are free
	bonus := Price((bonusDistance / 200) * 0.1) // 0.1$ per 500 meters
	return d.MinRunPrice + (bonus)              // proportionnal to the number of plates
}

func (d *Deliverer) SortOrder(orders []*Order) []*Order {
	sort.Slice(orders, func(i, j int) bool {
		return d.ComputeScoreRegardingOrder(orders[i]) > d.ComputeScoreRegardingOrder(orders[j])
	})
	return orders
}

func (d *Deliverer) moneyLeftToDo() Price {
	return d.GetDailyGoal() - d.GetMoneyMadeToday()
}

func (d *Deliverer) GetRunPriceThreshold(timeLeftInDay time.Duration, totalDistance float64, nPlates int8) Price {
	hypotheticalRunPrice := float64(d.GetHyptotheticalRunPrice(totalDistance, nPlates))
	moneyLeftToDo := float64(d.moneyLeftToDo())
	// At the begining of the day the threshold is the hypothetical run price
	// At the end of the day, if there is still lots of money to make, the threshold is the minimum run price
	threshold := hypotheticalRunPrice + math.Log(math.Exp(timeLeftInDay.Minutes()/60)/moneyLeftToDo) // gojek ne connait pas cette formule, sinon il maxera le gain à chaque fois
	threshold = math.Max(threshold, float64(d.MinRunPrice))
	threshold = math.Min(threshold, hypotheticalRunPrice) // A enlever ou pas si on considère que le livreur peut être greedy
	return Price(threshold)
}

func (d *Deliverer) AcceptOrder(o *Order, timeLeftInDay time.Duration) bool {
	distanceToRestaurant := d.Graph.GetDistanceBetweenNodes(d.GetPosition(), o.Restau.ClosestNode)
	customerApartment := o.Client.Apartment
	distanceFromRestaurantToCustomer := d.Graph.GetDistanceBetweenNodes(o.Restau.ClosestNode, customerApartment.ClosestNode)
	totalDistance := distanceToRestaurant + distanceFromRestaurantToCustomer
	threshold := d.GetRunPriceThreshold(timeLeftInDay, totalDistance, o.GetNbPlate())
	if o.GetRunPrice() >= threshold {
		return true
	}
	return false
}

func (d *Deliverer) createCoordinatePath(itinerary []graph.Node) [][2]Coordinate {
	var path [][2]Coordinate
	for _, n := range itinerary {
		node := d.Graph.getNodeById(NodeId(n.ID()))
		path = append(path, node.Position)
	}
	return path
}

func (d *Deliverer) MoveToNode(destination NodeId) bool {
	speedCoeff := d.env.accelerationTime
	itinerary, _ := d.env.GetShortestPathToNode(d.Position, destination)

	if len(itinerary) == 0 {
		return true
	}
	d.isMoving = true
	path := Path{currentIndex: 1, PathNode: d.createCoordinatePath(itinerary), PathTimestamp: make([]time.Time, len(itinerary))} // InitNewPath
	nextTimestamp := time.Now()
	path.PathTimestamp[0] = nextTimestamp
	path.Destination = path.PathNode[len(path.PathNode)-1]
	d.currentPaths[d.currentPathType] = path

	for i, nextNode := range itinerary[1:] { // We start at 1 because the first node is the current node
		t := make(chan *time.Timer)
		c := make(chan bool)
		currentLine := d.Graph.GetLineByNodes(d.Position, NodeId(nextNode.ID()))
		if currentLine == nil {
			return false
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			currentLine.AddVehiculeOnLine(t, c, speedCoeff)
		}()
		isAdded := <-c
		// If the deliverer is not added to the line, it means that the line is full, so we wait for a vehicule to leave
		for !isAdded {
			time.Sleep(time.Duration(5*speedCoeff) * time.Second) // We wait 5 seconds before trying again
			currentLine.AddVehiculeOnLine(t, c, speedCoeff)
			isAdded = <-c
		}
		// timer has started
		travelTime := time.Duration(currentLine.TravelTime*TravelTime(speedCoeff)*1000) * time.Millisecond
		nextTimestamp = nextTimestamp.Add(travelTime)
		path.PathTimestamp[i+1] = nextTimestamp
		path.currentIndex++
		d.currentPaths[d.currentPathType] = path
		timer := <-t
		<-timer.C // We wait for the timer to finish
		wg.Wait()

		currentLine.RemoveVehiculeFromLine()
		d.SetPosition(NodeId(nextNode.ID()))
	}

	d.isMoving = false
	return true
}
func (d *Deliverer) GetNOrders() int {
	return len(d.ratings)
}

func (d *Deliverer) ComputeScoreRegardingOrder(o *Order) float64 {
	eps := 0.001
	// We compute the score of the deliverer regarding an order
	distanceToRestaurant := d.Graph.GetDistanceBetweenNodes(d.GetPosition(), o.Restau.ClosestNode)
	rating := d.GetRating()
	nOrders := max(d.GetNOrders(), 1)                                                                 // Avoid division by 0
	return float64(rating) + 1/math.Pow(math.Log(distanceToRestaurant+eps), 2) - 1/(float64(nOrders)) // can be improved
}

func (d *Deliverer) TakeOrderAtRestaurant(o *Order, ctx context.Context) bool {
	for {
		select {
		case <-ctx.Done():
			// Timeout atteint
			return false
		default:
			if o.Restau.GiveOrder(o, d) {
				// Commande prête
				return true
			}
			// Attendre avant de réessayer
			time.Sleep(time.Duration(2 * d.env.accelerationTime * float32(time.Minute)))
		}
	}
}

func (d *Deliverer) GiveOrder(o *Order) bool {
	customer := o.Client
	if d.Position == customer.Apartment.ClosestNode {
		return customer.PickUpOrder(o)
	} else {
		o.SetState(-2)
		customer.OrderInfos <- o
		return false
	}
}

func (d *Deliverer) setActualDeliveryTime(o *Order) {
	o.ActualDeliveryTime = DeliveryTime(d.env.ReturnActualTime().Sub(o.ReturnBeginningDeliveryTime()).Minutes())
}
