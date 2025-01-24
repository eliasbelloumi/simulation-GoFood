package agt

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type LineID int32

type Line struct {
	ID                 LineID
	OsmId              interface{}  `json:"osmid"` // Useless for now
	Oneway             bool         `json:"oneway"`
	Name               interface{}  `json:"name"`
	Highway            interface{}  `json:"highway"`
	MaxSpeed           interface{}  `json:"maxspeed"`
	Reversed           interface{}  `json:"reversed"`
	Geometry           LineGeometry `json:"geometry"`
	Length             float64      `json:"length"`
	SpeedKph           float64      `json:"speed_kph"`
	TravelTime         TravelTime   `json:"travel_time"` // on incrémmentera le travel time en fonction du nombre d'usager sur ce tronçont
	Source             NodeId       `json:"source"`
	Target             NodeId       `json:"target"`
	Key                int64        `json:"key"`
	maxVehiculesOnline int32
	vehiculesOnLine    int32
	VehiculesWaiter    chan int32
}

func (l *Line) GetNumberVehiculesOnLine() int32 {
	return l.vehiculesOnLine
}
func (l *Line) RemoveVehiculeFromLine() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// use atomic to ensure that the value is not changed by another goroutine
		freezedNVehicules := atomic.LoadInt32(&l.vehiculesOnLine)
		if freezedNVehicules > 0 {
			if atomic.CompareAndSwapInt32(&l.vehiculesOnLine, freezedNVehicules, freezedNVehicules-1) {
				//fmt.Println("Vehicule removed from line ", l.ID)
			} else {
				//fmt.Println("Failed to remove vehicule from line ", l.ID)
			}
		} else {
			fmt.Println("Line ", l.ID, " is empty")
		}
	}()
	wg.Wait()
}

func (l *Line) AddVehiculeOnLine(t chan *time.Timer, c chan bool, speedCoeff float32) { // revoir ça pu frr
	freezedNVehicules := atomic.LoadInt32(&l.vehiculesOnLine)
	if freezedNVehicules < l.maxVehiculesOnline {
		if atomic.CompareAndSwapInt32(&l.vehiculesOnLine, freezedNVehicules, freezedNVehicules+1) {
			c <- true
			timer := time.NewTimer(time.Duration(l.TravelTime*TravelTime(speedCoeff)) * time.Second)
			t <- timer
		} else {
			fmt.Println("Failed to add vehicule on line ", l.ID)
			c <- false
		}
	} else {
		fmt.Println("Line ", l.ID, " is full")
		c <- false
	}
}

func (l *Line) GetMaxVehiculesOnLine() int32 {
	return l.maxVehiculesOnline
}
