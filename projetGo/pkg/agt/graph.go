package agt

import (
	"fmt"
	"math"
	"sync"

	"golang.org/x/exp/rand"
)

type Graph struct {
	Nodes                 []*Node `json:"nodes"`
	Lines                 []*Line `json:"links"`
	nodesMap              map[NodeId]*Node
	linesMap              map[string]*Line
	distancesBetweenNodes *sync.Map
}

type requestLine struct {
	source NodeId
	target NodeId
}

func (g *Graph) GetLineByNodes(source NodeId, target NodeId) *Line {
	key := fmt.Sprintf("%d-%d", source, target)
	line, ok := g.linesMap[key]
	if ok {
		return line
	}
	return nil
}

func (g *Graph) getNodeById(id NodeId) *Node {
	node, ok := g.nodesMap[id]
	if ok {
		return node
	}
	return nil
}

func (g *Graph) GetDistanceBetweenNodes(source NodeId, target NodeId) float64 {

	key := requestLine{source: source, target: target}
	if dist, ok := g.distancesBetweenNodes.Load(key); ok {
		return dist.(float64)
	}
	fromNode := g.getNodeById(source)
	toNode := g.getNodeById(target)
	// compute euclidean distance
	const R = 6371000.0 // Earth's radius in meters

	// Convert degrees to radians
	toRadians := func(deg Coordinate) float64 {
		return float64(deg * math.Pi / 180.0)
	}
	lat1, lon1, lat2, lon2 := toRadians(fromNode.X), toRadians(fromNode.Y), toRadians(toNode.X), toRadians(toNode.Y)

	// Compute deltas
	dLat := lat2 - lat1
	dLon := lon2 - lon1

	// Haversine formula
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	// Distance in meters
	distance := R * c
	g.distancesBetweenNodes.Store(key, distance)
	return distance
}

func (g *Graph) GetNodeWithMostRestaurantsNear(node NodeId, radius int16) *Node {
	var maxRestaurantsNear int16
	var bestNode = g.getNodeById(node)
	for _, n := range g.Nodes {
		dist := g.GetDistanceBetweenNodes(node, n.Id)
		if dist < float64(radius) && n.NumberRestaurantsNear > maxRestaurantsNear {
			maxRestaurantsNear = n.NumberRestaurantsNear
			bestNode = n
		}
	}
	return bestNode
}

func (g *Graph) GetRandomNode() *Node {
	randomIndex := rand.Intn(len(g.Nodes))
	return g.Nodes[randomIndex]
}
