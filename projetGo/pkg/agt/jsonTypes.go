package agt

type Response struct {
	MsgType          string           `json:"msgType"`
	CurrentTimestamp float32          `json:"currentTimestamp,omitempty"`
	ActualTime       string           `json:"actualTime,omitempty"`
	AccelerationTime float32          `json:"accelerationTime,omitempty"`
	Restaurants      []RestaurantJSON `json:"restaurants,omitempty"`
	Customers        []CustomerJson   `json:"customers,omitempty"`
	Deliverers       []DelivererJSON  `json:"deliverers,omitempty"`
	InitInfos        InitInfosJSON    `json:"initInfos,omitempty"`
	Statistics       StatisticsJSON   `json:"statistics,omitempty"`
}

type RestaurantJSON struct {
	Name            string        `json:"name"`
	Coordinates     [2]Coordinate `json:"coordinates"`
	Amenity         string        `json:"amenity,omitempty"`
	FoodType        FoodType      `json:"foodType,omitempty"`
	Schedule        string        `json:"schedule"`
	Price           string        `json:"price"`
	PreparationTime string        `json:"preparationTime"`
	Stock           string        `json:"stock"`
}

type CustomerJson struct {
	Name            string          `json:"name"`
	Coordinates     [2]Coordinate   `json:"coordinates"`
	HungryLevel     string          `json:"hungryLevel,omitempty"`
	FoodPreferences FoodPreferences `json:"foodPreferences"`
	WantsToOrder    bool            `json:"wantsToOrder"`
}

type DelivererJSON struct {
	Name            string        `json:"name"`
	Position        [2]Coordinate `json:"position"`
	Rating          string        `json:"rating"`
	State           string        `json:"state"`
	IsMoving        bool          `json:"isMoving"`
	DailyGoal       string        `json:"dailyGoal"`
	MoneyMadeToday  string        `json:"moneyMadeToday"`
	NumOrder        int32         `json:"numOrder"`
	CurrentPathType string        `json:"currentPathType"`
	ReplacementPath PathJSON      `json:"replacementPath,omitempty"`
	ToRestauPath    PathJSON      `json:"toRestauPath,omitempty"`
	ToClientPath    PathJSON      `json:"toClientPath,omitempty"`
	CurrentOrder    OrderJSON     `json:"currentOrder,omitempty"`
}

type PathJSON struct {
	PathNode      [][2]Coordinate `json:"pathNode,omitempty"`
	PathTimestamp []float32       `json:"pathTimestamp,omitempty"`
	Destination   [2]Coordinate   `json:"destination,omitempty"`
}

type OrderJSON struct {
	NbPlate             int8          `json:"nbPlate"`
	RestauName          string        `json:"restauName"`
	RestauCoordinates   [2]Coordinate `json:"restauCoordinates"`
	CustomerName        string        `json:"customerName"`
	CustomerCoordinates [2]Coordinate `json:"customerCoordinates"`
	Price               string        `json:"price"`
	RunPrice            string        `json:"runPrice"`
	TotalPrice          string        `json:"totalPrice,omitempty"`
}

type InitInfosJSON struct {
	AccelerationTime string `json:"accelerationTime,omitempty"`
	Duration         string `json:"duration,omitempty"`
	NumDeliverer     int    `json:"numDeliverer,omitempty"`
	NumRestau        int    `json:"numRestau,omitempty"`
	NumCustomer      int    `json:"numCustomer,omitempty"`
}

type StatisticsJSON struct {
	Time                    float32 `json:"time"`
	RunPriceRatio           float32 `json:"runPriceRatio"`
	MoneyMade               Price   `json:"moneyMade"`
	AverageMoneyMadeByOrder Price   `json:"averageMoneyMadeByOrder"`
	NumOrder                int     `json:"numOrder"`
}
