package agt

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

func (e *Environment) SerializeInit() []byte {
	response := Response{
		MsgType:          "init",
		ActualTime:       e.actualTime.Format(time.Kitchen),
		AccelerationTime: e.accelerationTime,
		CurrentTimestamp: convertTimestampToFloat32(time.Now(), e.startTime),
		InitInfos: InitInfosJSON{
			AccelerationTime: fmt.Sprintf("× %d", int(1/e.accelerationTime)),
			Duration:         fmt.Sprintf("%d h", e.duration),
			NumDeliverer:     e.numDeliverers,
			NumCustomer:      e.numCustomers,
			NumRestau:        len(e.Restaurants),
		},
	}

	initJson, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("Error serializing initJSon: %v", err)
	}

	return initJson
}

func findLatestStatistics(m *sync.Map) *Statistics {
	var latest *Statistics
	var found bool

	fmt.Println("Finding latest statistics")
	m.Range(func(key, value interface{}) bool {
		stats, ok := value.(*Statistics)
		if !ok {
			log.Fatalf("Error casting statistics")
			return true
		}
		fmt.Println("Stats", stats)
		if !found || stats.Time.After(latest.Time) {
			latest = stats
			found = true
		}
		return true // Continue iteration
	})

	return latest
}
func TimeToFloat(t time.Time) float32 {
	hours := t.Hour()
	minutes := t.Minute()
	floatTime := float32(hours) + float32(minutes)/60
	return floatTime
}

func (e *Environment) SerializeStatistics() []byte {
	lastStat := findLatestStatistics(&e.Statistics)

	response := Response{
		MsgType: "statistics",
		Statistics: StatisticsJSON{
			Time:                    TimeToFloat(lastStat.Time),
			RunPriceRatio:           lastStat.RunPriceRatio,
			MoneyMade:               lastStat.MoneyMade,
			NumOrder:                lastStat.NumberOfOrder,
			AverageMoneyMadeByOrder: lastStat.AverageMoneyMadeByOrder,
		}}

	statisticsJSON, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("Error serializing statistcis: %v", err)
	}

	return statisticsJSON
}

func (e *Environment) SerializeTotalStatistics() []byte {
	response := Response{
		MsgType: "totalStatistics",
		Statistics: StatisticsJSON{
			MoneyMade:               e.TotalStatistics.MoneyMade,
			NumOrder:                e.TotalStatistics.NumberOfOrder,
			AverageMoneyMadeByOrder: e.TotalStatistics.AverageMoneyMadeByOrder,
		}}

	statisticsJSON, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("Error serializing total Stat: %v", err)
	}

	return statisticsJSON
}

func (e *Environment) SerializeRestaurants() []byte {
	restaurants := e.Restaurants
	var serializedRestaurants []RestaurantJSON
	for _, r := range restaurants {
		serializedRestaurants = append(serializedRestaurants, RestaurantJSON{
			Name:            r.Name,
			Coordinates:     r.Geometry,
			Amenity:         r.Amenity,
			FoodType:        r.Cuisine,
			Schedule:        r.schedules[0].Format(time.Kitchen) + " - " + r.schedules[1].Format(time.Kitchen),
			Price:           fmt.Sprintf("%.2f", r.price) + " €",
			PreparationTime: r.PreparationTime.String(),
			Stock:           fmt.Sprintf("%d plates left", r.stock),
		})
	}

	response := Response{
		MsgType:     "restaurants",
		Restaurants: serializedRestaurants,
	}

	// Serialize the structure to JSON
	restaurantsJson, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("Error serializing restaurants: %v", err)
	}

	return restaurantsJson
}

func (e *Environment) SerializeCustomers() []byte {
	var serializedCustomers []CustomerJson
	for _, c := range e.Customers {
		serializedCustomers = append(serializedCustomers, CustomerJson{
			Name:            fmt.Sprintf("Customer %d", c.Id),
			Coordinates:     c.Apartment.Geometry,
			HungryLevel:     fmt.Sprintf("%.2f", c.hungryLevel*100) + " %",
			FoodPreferences: c.foodPreferences,
			WantsToOrder:    c.wantsToOrder,
		})
	}

	response := Response{
		MsgType:   "customers",
		Customers: serializedCustomers,
	}

	// Serialize the structure to JSON
	customersJson, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("Error serializing customers: %v", err)
	}

	return customersJson
}

func convertTimestampsToFloat32(timestamps []time.Time, referenceTime time.Time) []float32 {
	floatTimestamps := make([]float32, len(timestamps))
	for i, t := range timestamps {
		floatTimestamps[i] = convertTimestampToFloat32(t, referenceTime)
	}
	return floatTimestamps
}

func convertTimestampToFloat32(timestamp time.Time, referenceTime time.Time) float32 {
	return float32(timestamp.Sub(referenceTime).Milliseconds())
}

func getDelivererState(d *Deliverer) string {
	if !d.isMoving && d.Available {
		return "Waiting for order"
	}
	if d.isMoving && d.Available {
		return "Going to a better position"
	}
	if d.isMoving && !d.Available {
		if d.currentPathType == "toRestau" {
			return "Going to restaurant to retreive order"
		}
		if d.currentPathType == "toClient" {
			return "Delivering order to customer"
		}
	}
	if !d.isMoving && !d.Available {
		if d.currentPathType == "toRestau" {
			return "Waiting for order at restaurant"
		}
		if d.currentPathType == "toClient" {
			return "Waiting for customer to take order"
		}
	}
	return "End of the day"
}

func (e *Environment) SerializeDeliverers() []byte {
	var serializedDeliverers []DelivererJSON
	e.Deliverers.Range(func(key, value any) bool {
		d := value.(*Deliverer)

		delivererJson := DelivererJSON{
			Name:            fmt.Sprintf("Deliverer %d", d.Id),
			IsMoving:        d.isMoving,
			Rating:          fmt.Sprintf("%.1f ⭐", d.GetRating()),
			State:           getDelivererState(d),
			Position:        e.Graph.getNodeById(d.Position).Position,
			DailyGoal:       fmt.Sprintf("%.2f", d.DailyGoal) + " €",
			MoneyMadeToday:  fmt.Sprintf("%.2f", d.MoneyMadeToday) + " €",
			NumOrder:        int32(d.GetNOrders()),
			CurrentPathType: d.currentPathType,
		}

		if d.currentOrder != nil {
			delivererJson.CurrentOrder = OrderJSON{
				NbPlate:             d.currentOrder.nbPlate,
				RestauName:          d.currentOrder.Restau.Name,
				RestauCoordinates:   d.currentOrder.Restau.Geometry,
				CustomerName:        fmt.Sprintf("Customer %d", d.currentOrder.Client.Id),
				CustomerCoordinates: d.currentOrder.Client.Apartment.Geometry,
				Price:               fmt.Sprintf("%.2f", d.currentOrder.price) + " €",
				RunPrice:            fmt.Sprintf("%.2f", d.currentOrder.runPrice) + " €",
			}
		}

		currentPath := d.currentPaths[d.currentPathType]
		if d.currentPathType == "replacement" {
			delivererJson.ReplacementPath = PathJSON{PathNode: currentPath.PathNode[:currentPath.currentIndex], PathTimestamp: convertTimestampsToFloat32(currentPath.PathTimestamp, e.startTime)[:currentPath.currentIndex], Destination: currentPath.Destination}
		} else {
			replacementPath := d.currentPaths["replacement"]
			if replacementPath.currentIndex != 0 {
				delivererJson.ReplacementPath = PathJSON{PathNode: replacementPath.PathNode, PathTimestamp: convertTimestampsToFloat32(replacementPath.PathTimestamp, e.startTime), Destination: replacementPath.Destination}
			}

		}
		if d.currentPathType == "toRestau" {
			delivererJson.ToRestauPath = PathJSON{PathNode: currentPath.PathNode[:currentPath.currentIndex], PathTimestamp: convertTimestampsToFloat32(currentPath.PathTimestamp, e.startTime)[:currentPath.currentIndex], Destination: currentPath.Destination}
		} else {
			toRestauPath := d.currentPaths["toRestau"]
			if toRestauPath.currentIndex != 0 {
				delivererJson.ToRestauPath = PathJSON{PathNode: toRestauPath.PathNode, PathTimestamp: convertTimestampsToFloat32(toRestauPath.PathTimestamp, e.startTime), Destination: toRestauPath.Destination}
			}
		}
		if d.currentPathType == "toClient" {
			delivererJson.ToClientPath = PathJSON{PathNode: currentPath.PathNode[:currentPath.currentIndex], PathTimestamp: convertTimestampsToFloat32(currentPath.PathTimestamp, e.startTime)[:currentPath.currentIndex], Destination: currentPath.Destination}
		} else {
			toClientPath := d.currentPaths["toClient"]
			if toClientPath.currentIndex != 0 {
				delivererJson.ToClientPath = PathJSON{PathNode: toClientPath.PathNode, PathTimestamp: convertTimestampsToFloat32(toClientPath.PathTimestamp, e.startTime), Destination: toClientPath.Destination}
			}
		}

		serializedDeliverers = append(serializedDeliverers, delivererJson)
		return true
	})

	response := Response{
		MsgType:          "deliverers",
		CurrentTimestamp: convertTimestampToFloat32(time.Now(), e.startTime),
		ActualTime:       e.actualTime.Format(time.Kitchen),
		Deliverers:       serializedDeliverers,
	}

	// Serialize the structure to JSON
	deliverersJson, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("Error serializing deliverers: %v", err)
	}

	return deliverersJson
}
