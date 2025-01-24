package agt

import (
	"time"
)

type Agent interface {
	Perceive()
	Deliberate()
	Act()
	Start()
	Stop()
}

type Path struct {
	currentIndex  int
	Destination   [2]Coordinate
	PathNode      [][2]Coordinate
	PathTimestamp []time.Time
}

type OrderId int
type PreparationTime int

type DelivererId int32

type FoodType string

type OsmId float64

type Price float32

type FoodPreferences []FoodType

type OrderState int64

type RestaurantInfosForOrderChoice struct {
	// utilisé pour envoyer les informations sur un restaurant au client au moment de faire son choix
	id          OsmId
	genre       FoodType
	waitingTime time.Duration
	price       Price
}

type OrderStructRequest struct {
	Preferences ListRestau
	NbPlat      int8
	Customer    *Customer
	DownBit     bool
}

type Schedule time.Time

type DeliveryTime time.Duration

type RestaurantChoice []RestaurantInfosForOrderChoice

type Stock int // stock de nourriture d'un restaurant

type AcceptOrder bool

type Kitchen []Order

type DelivererPreferences []Deliverer

type NodeId int64
type Coordinate float32

type Node struct {
	Y                     Coordinate `json:"y"`
	X                     Coordinate `json:"x"`
	Position              [2]Coordinate
	Ref                   string `json:"ref"`
	Highway               string `json:"highway"`
	StreetCount           int64  `json:"street_count"`
	Id                    NodeId `json:"id"`
	NumberRestaurantsNear int16
}

type LineGeometry struct {
	Type        string          `json:"type"`
	Coordinates [][2]Coordinate `json:"coordinates"`
}
type TravelTime float32

type Apartment struct {
	Osmid               OsmId         `json:"osmid"`
	AddrHousenumber     string        `json:"addr:housenumber"`
	AddrStreet          string        `json:"addr:street"`
	Geometry            [2]Coordinate `json:"geometry"`
	ClosestNode         NodeId        `json:"closest_node"`
	DistanceClosestNode float64       `json:"distance_noeud"`
}

type restaurantToOrder struct { // strucutre permettant de noter les restaurants proposés par gojek
	value float32
	id    *Restaurant
}

type restaurantToSelect []restaurantToOrder // liste de restaurants possibles notés

func (r *restaurantToSelect) extractNames() ListRestau {
	var ids ListRestau
	for _, r := range *r {
		ids = append(ids, r.id)
	}
	return ids
}

// implementation d'une interface permettant d'utiliser la bibliotheque sort
func (r restaurantToSelect) Less(i, j int) bool {
	return r[i].value > r[j].value
}

func (r restaurantToSelect) Len() int {
	return len(r)
}

func (r restaurantToSelect) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
