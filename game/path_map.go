package game

import (
	"contractor/foundation"
	"contractor/gridmap"
	"github.com/memmaker/go/geometry"
	"slices"
)

type MapPosition struct {
	MapName  string
	Position geometry.Point
	// LocationName is optional and may be empty, will only be set if the position is a named location
	LocationName string
}

type Pathfinder struct {
	getMap         func(string) *gridmap.GridMap[*Actor, foundation.Item, Object]
	neighborsCache map[*Actor]map[MapPosition][]MapPosition
}

func NewPathfinder(getMap func(string) *gridmap.GridMap[*Actor, foundation.Item, Object]) *Pathfinder {
	return &Pathfinder{
		getMap:         getMap,
		neighborsCache: make(map[*Actor]map[MapPosition][]MapPosition),
	}
}

func (p *Pathfinder) getNeighbors(actor *Actor, from MapPosition, to MapPosition) []MapPosition {
	neighbors, cacheExists := p.neighborsCache[actor][from]

	if !cacheExists {
		gMap := p.getMap(from.MapName)
		// The name is authoritative (named locations can move); unnamed targets like spawn points only have a position.
		location := from.Position
		if pos, named := gMap.TryGetNamedLocation(from.LocationName); named {
			location = pos
		}

		if transitionTo, exists := gMap.GetTransitionAt(location); exists {
			if targetMap := p.getMap(transitionTo.TargetMap); targetMap != nil { // dangling transition to a deleted map
				targetLocation := targetMap.GetNamedLocation(transitionTo.TargetLocation)
				neighbors = append(neighbors, MapPosition{MapName: transitionTo.TargetMap, LocationName: transitionTo.TargetLocation, Position: targetLocation})
			}
		}

		for locationOfTransition, _ := range gMap.Transitions() {
			if locationOfTransition == location {
				continue
			}
			pathToOtherTransition := gMap.GetJPSPath(location, locationOfTransition, func(point geometry.Point) bool {
				return gMap.IsWalkableFor(point, actor)
			})

			if len(pathToOtherTransition) > 0 {
				nameOfLocation := gMap.GetNamedLocationByPos(locationOfTransition)
				neighbors = append(neighbors, MapPosition{MapName: from.MapName, LocationName: nameOfLocation, Position: locationOfTransition})
			}
		}

		if to != from && to.MapName == from.MapName {
			pathToTarget := gMap.GetJPSPath(location, to.Position, func(point geometry.Point) bool {
				return gMap.IsWalkableFor(point, actor)
			})
			if len(pathToTarget) > 0 {
				neighbors = append(neighbors, to)
			}
		}

		if p.neighborsCache[actor] == nil {
			p.neighborsCache[actor] = make(map[MapPosition][]MapPosition)
		}
		p.neighborsCache[actor][from] = neighbors
	}

	return neighbors
}

// FindPath returns a path of transitions from the current map to the target map.
// It expects the actor to have a DijkstraMap that has been updated with the current map.
func (p *Pathfinder) FindPath(actor *Actor, currentMap string, to MapPosition) []MapPosition {
	p.neighborsCache[actor] = make(map[MapPosition][]MapPosition)

	reachable := actor.GetDijkstraMap()
	gMap := p.getMap(currentMap)
	reachableTransitions := make(map[MapPosition]int)
	openTransitions := make(map[MapPosition]int)
	for pos, _ := range reachable {
		if _, exists := gMap.GetTransitionAt(pos); exists {

			locationName := gMap.GetNamedLocationByPos(pos)

			transition := MapPosition{MapName: currentMap, LocationName: locationName, Position: pos}

			openTransitions[transition] = 1
			reachableTransitions[transition] = 1
		}
	}
	closedTransitions := make(map[MapPosition]int)
	parents := make(map[MapPosition]MapPosition) // first-discovered predecessor; start nodes have none

	for {
		if len(openTransitions) == 0 {
			break
		}
		for transition, dist := range openTransitions {
			closedTransitions[transition] = dist
			delete(openTransitions, transition)

			neighbors := p.getNeighbors(actor, transition, to)
			for _, neighbor := range neighbors {
				if _, exists := closedTransitions[neighbor]; exists {
					continue
				}
				if _, exists := openTransitions[neighbor]; exists {
					continue
				}
				openTransitions[neighbor] = dist + 1
				parents[neighbor] = transition
			}
		}
	}

	if _, exists := closedTransitions[to]; !exists {
		return nil
	}

	path := make([]MapPosition, 0)

	// Walk the recorded parents back; transitions are one-way, so re-deriving
	// predecessors from forward neighbors can dead-end and spin forever.
	current := to
	path = append(path, current)
	for reachableTransitions[current] != 1 {
		current = parents[current]
		if current.MapName != path[len(path)-1].MapName {
			path = append(path, current)
		}
	}

	slices.Reverse(path)

	return path
}
