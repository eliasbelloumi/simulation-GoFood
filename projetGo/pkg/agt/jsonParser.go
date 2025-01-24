package agt

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sync"

	"gonum.org/v1/gonum/graph/multi"
)

func ParseRestaurants(path string, c chan []*Restaurant) {
	// OPEN FILE
	file, fileErr := os.Open(path)
	if fileErr != nil {
		log.Fatalf("Failed to open JSON file: %v", fileErr)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	// on va lire le tableau
	_, err := decoder.Token()
	if err != nil {
		fmt.Println("Erreur lors de la lecture du fichier JSON:", err)
		c <- nil
	}

	var restaurants = []*Restaurant{}
	for decoder.More() {
		var restaurant Restaurant
		err := restaurant.Populate(decoder)
		if err != nil {
			fmt.Println("Erreur lors du décodage d'un restaurant:", err)
			c <- nil
		}
		restaurant.Init()
		restaurants = append(restaurants, &restaurant)

	}

	_, err = decoder.Token() // Lire le dernier token (`]`)
	if err != nil {
		fmt.Println("Erreur lors de la fermeture du tableau JSON:", err)
	}

	fmt.Println("Restaurants have been loaded in Go")
	c <- restaurants
}

func ParseAppartements(path string, c chan []*Apartment) {
	// OPEN FILE
	file, fileErr := os.Open(path)
	if fileErr != nil {
		log.Fatalf("Failed to open JSON file: %v", fileErr)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	// on va lire le tableau
	_, err := decoder.Token()
	if err != nil {
		fmt.Println("Erreur lors de la lecture du fichier JSON:", err)
		c <- nil
	}

	var restaurants = []*Apartment{}
	for decoder.More() {
		var restaurant Apartment
		err := decoder.Decode(&restaurant)
		if err != nil {
			fmt.Println("Erreur lors du décodage d'un restaurant:", err)
			c <- nil
		}

		restaurants = append(restaurants, &restaurant)

	}

	_, err = decoder.Token() // Lire le dernier token (`]`)
	if err != nil {
		fmt.Println("Erreur lors de la fermeture du tableau JSON:", err)
	}

	fmt.Println("Appartments have been loaded in Go")
	c <- restaurants
}

func LoadGraphAndGPSfromJSON(path string, m chan *multi.WeightedDirectedGraph, g chan *Graph) {
	// OPEN FILE
	file, fileErr := os.Open(path)
	if fileErr != nil {
		log.Fatalf("Failed to open JSON file: %v", fileErr)
	}
	defer file.Close()
	// READ FILE CONTENT
	byteValue, readErr := ioutil.ReadAll(file)
	if readErr != nil {
		log.Fatalf("Failed to read JSON file: %v", readErr)
	}
	//PARSE JSON FILE
	var graphStruct = Graph{}
	parsingErr := json.Unmarshal(byteValue, &graphStruct)
	if parsingErr != nil {
		log.Fatalf("Failed to parse JSON: %v", parsingErr)
	}
	graphStruct.linesMap = make(map[string]*Line)
	graphStruct.nodesMap = make(map[NodeId]*Node)
	graphStruct.distancesBetweenNodes = new(sync.Map)
	fmt.Println("New parsed Graph has been loaded in Go")
	//CREATE NEW GRAPH
	newWeightedGraph := multi.NewWeightedDirectedGraph()
	//ADD ALL THE NODES
	for _, newNodeRequest := range graphStruct.Nodes {
		newNode, newNodeCheck := newWeightedGraph.NodeWithID(int64(newNodeRequest.Id))
		if newNodeCheck != true {
			fmt.Println("Failed to load Node n°", newNodeRequest.Id)
		} else {
			newNodeRequest.Position[0] = newNodeRequest.X
			newNodeRequest.Position[1] = newNodeRequest.Y
			newWeightedGraph.AddNode(newNode)
			graphStruct.nodesMap[newNodeRequest.Id] = newNodeRequest
		}
	}
	//ADD ALL THE LINES
	for i, newLineRequest := range graphStruct.Lines {
		//add in graphStruct with key being the start and end node
		key := fmt.Sprintf("%d-%d", newLineRequest.Source, newLineRequest.Target)
		graphStruct.linesMap[key] = newLineRequest
		//Set non-parsed attributes
		newLineRequest.maxVehiculesOnline = max(int32(math.Round(newLineRequest.Length/3)), 1) //We consider 3 meters is the average size of a vehicule, althought there is always enough space for 1 vehicule
		newLineRequest.ID = LineID(i)
		newLineRequest.VehiculesWaiter = make(chan int32)
		// Add new line to the Weighted Graph
		fromNode := newWeightedGraph.Node(int64(newLineRequest.Source))
		toNode := newWeightedGraph.Node(int64(newLineRequest.Target))
		newLine := newWeightedGraph.NewWeightedLine(fromNode, toNode, float64(newLineRequest.TravelTime))
		newWeightedGraph.SetWeightedLine(newLine)

		if newLineRequest.Oneway == false {
			reversedNewLine := newWeightedGraph.NewWeightedLine(toNode, fromNode, float64(newLineRequest.TravelTime))
			newWeightedGraph.SetWeightedLine(reversedNewLine)
		}
	}
	//TEST
	fmt.Println("Directed grap has been loaded in Go")
	m <- newWeightedGraph
	g <- &graphStruct
}

func main() {
	fmt.Println("All JSON files have been loaded in Go")
}
