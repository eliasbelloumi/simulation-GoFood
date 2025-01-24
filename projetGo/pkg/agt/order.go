package agt

import (
	"time"

	"golang.org/x/exp/rand"
)

type Order struct {
	Id                       OrderId
	state                    OrderState    // Etat de la commande, de 1 à 7. ACCES CONCURRENT
	temperature              float32       // temperature de la commande UTILE ???
	nbPlate                  int8          // nombre de plats commandés (max 256)
	estimatedPreparationTime time.Duration //temps de preparation estimé
	BeginningDeliveryTime    time.Time     // moment où la livraison commence
	EstimatedDeliveryTime    DeliveryTime  // moment où la livraison est censée se finir
	ActualDeliveryTime       DeliveryTime
	gap                      time.Duration // gap entre le moment de début et de fin de commande
	Restau                   *Restaurant   // restaurant en charge de la commande
	Client                   *Customer     // client en charge de la commande
	Deliverer                *Deliverer    // livreur en charge de la commande
	price                    Price         //prix des plat
	runPrice                 Price         // prix de la course
	DelivererWaiter          chan *Deliverer
	Ttl                      int       // nombre de fois où gojek doit tenter de trouver un livreur avant d'annuler la commande
	BeginningOrdering        time.Time // heure où la commande est crée par gojek

}

func (o *Order) Close() {
	close(o.DelivererWaiter)
}

func NewOrderForTest() *Order {
	o := Order{Id: 2, state: 1, nbPlate: int8(rand.Intn(2) + 2)}
	return &o
}

func (o *Order) SetDeliverer(d *Deliverer) bool {
	if o.Deliverer != nil {
		return false
	} else {
		o.Deliverer = d
		return true
	}
}

func (o *Order) ReturnBeginningDeliveryTime() time.Time {
	return o.BeginningDeliveryTime
}
func (o *Order) ReturnDeliverer() *Deliverer {
	return o.Deliverer
}

func (o *Order) ReturnCustomer() *Customer {
	return o.Client
}

func (o *Order) SetGap(param time.Duration) {
	o.gap = param
}

// GetId retourne l'ID de la commande
func (o *Order) GetId() OrderId {
	return o.Id
}

// SetId modifie l'ID de la commande
func (o *Order) SetId(id OrderId) {
	o.Id = id
}

func (o *Order) GetState() OrderState {
	return o.state
}

func (o *Order) SetState(state OrderState) {
	o.state = state
}

func (o *Order) SetRestaurant(r *Restaurant) {
	o.Restau = r
}

// GetTemperature retourne la température
func (o *Order) GetTemperature() float32 {
	return o.temperature
}

// SetTemperature modifie la température
func (o *Order) SetTemperature(temp float32) {
	o.temperature = temp
}

// GetNbPlate retourne le nombre de plats
func (o *Order) GetNbPlate() int8 {
	return o.nbPlate
}

// SetNbPlate modifie le nombre de plats
func (o *Order) SetNbPlate(nb int8) {
	o.nbPlate = nb
}

// GetEstimatedPreparationTime retourne le temps estimé de préparation
func (o *Order) GetEstimatedPreparationTime() time.Duration {
	return o.estimatedPreparationTime
}

// SetEstimatedPreparationTime modifie le temps estimé de préparation
func (o *Order) SetEstimatedPreparationTime(duration time.Duration) {
	o.estimatedPreparationTime = duration
}

// GetEstimatedDeliveryTime retourne l'heure estimée de livraison
func (o *Order) GetEstimatedDeliveryTime() DeliveryTime {
	return o.EstimatedDeliveryTime
}

// SetEstimatedDeliveryTime modifie l'heure estimée de livraison
func (o *Order) SetEstimatedDeliveryTime(time DeliveryTime) {
	o.EstimatedDeliveryTime = time
}

// GetPrice retourne le prix du plat
func (o *Order) GetPrice() Price {
	return o.price
}

// SetPrice modifie le prix du plat
func (o *Order) SetPrice(price Price) {
	o.price = price
}

// GetRunPrice retourne le prix de la course
func (o *Order) GetRunPrice() Price {
	return o.runPrice
}

// SetRunPrice modifie le prix de la course
func (o *Order) SetRunPrice(price Price) {
	o.runPrice = price
}

// GetTotalPrice retourne le prix total
func (o *Order) GetTotalPrice() Price {
	return o.runPrice + o.price
}

func (o *Order) GetGojekCommission() Price {
	return o.GetTotalPrice() * 0.1
}

func (o *Order) ReturnOrderState() OrderState {
	return o.state
}

func (o *Order) GetDeliveryDelay() DeliveryTime {
	return o.ActualDeliveryTime - o.EstimatedDeliveryTime
}

func (o *Order) SetBeginningDeliveryTime(t time.Time) {
	o.BeginningDeliveryTime = t
}

func (o *Order) GetActualDeliveryTime() DeliveryTime {
	return o.ActualDeliveryTime
}
