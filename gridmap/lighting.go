package gridmap

import (
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"strconv"
	"strings"
	"time"
)

type LightSource struct {
	Pos          geometry.Point
	Radius       int
	Color        fxtools.HDRColor
	MaxIntensity float64
}

// LightFalloff switches all lights between flat (full intensity up to the radius)
// and linear falloff with distance. Set from the LightFalloff config option.
var LightFalloff bool

// ColorAt returns the light's contribution at the given distance from its center.
func (s *LightSource) ColorAt(dist float64) fxtools.HDRColor {
	intensity := s.MaxIntensity
	if LightFalloff {
		intensity *= 1 - dist/float64(s.Radius+1)
	}
	return s.Color.MultiplyWithScalar(intensity)
}

func (s LightSource) ToRecord() []recfile.Field {
	return []recfile.Field{
		{Name: "Position", Value: s.Pos.String()},
		{Name: "Radius", Value: strconv.Itoa(s.Radius)},
		{Name: "Color", Value: s.Color.EncodeAsString()},
		{Name: "Max_Intensity", Value: strconv.FormatFloat(s.MaxIntensity, 'f', 2, 64)},
	}
}

func NewLightSourceFromRecord(record []recfile.Field) *LightSource {
	var result LightSource
	for _, field := range record {
		switch strings.ToLower(field.Name) {
		case "position":
			result.Pos, _ = geometry.NewPointFromString(field.Value)
		case "radius":
			result.Radius, _ = strconv.Atoi(field.Value)
		case "color":
			result.Color = fxtools.NewColorFromString(field.Value)
		case "max_intensity":
			result.MaxIntensity, _ = strconv.ParseFloat(field.Value, 64)
		}
	}
	return &result
}

// AddDynamicLightSource adds a light source to the map. Call UpdateDynamicLights afterwards.
func (m *GridMap[ActorType, ItemType, ObjectType]) AddDynamicLightSource(pos geometry.Point, light *LightSource) {
	if m.IsDynamicLightSource(pos) {
		return
	}
	m.DynamicLights[pos] = light
	light.Pos = pos
}

// AddBakedLightSource adds a light source to the map. Call UpdateBakedLights afterwards.
func (m *GridMap[ActorType, ItemType, ObjectType]) AddBakedLightSource(pos geometry.Point, light *LightSource) {
	if m.IsBakedLightSource(pos) {
		return
	}
	m.BakedLights[pos] = light
}
func (m *GridMap[ActorType, ItemType, ObjectType]) SetAmbientLight(color fxtools.HDRColor) {
	m.meta = m.meta.WithAmbientLight(color)
}
func (m *GridMap[ActorType, ItemType, ObjectType]) IsDynamicLightSource(pos geometry.Point) bool {
	_, ok := m.DynamicLights[pos]
	return ok
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsBakedLightSource(pos geometry.Point) bool {
	_, ok := m.BakedLights[pos]
	return ok
}

// MoveLightSource moves (or registers) a light source on this map and recasts the dynamic lights.
// It is a no-op if a light (this one or another) already occupies the target.
func (m *GridMap[ActorType, ItemType, ObjectType]) MoveLightSource(lightSource *LightSource, to geometry.Point) {
	if m.IsDynamicLightSource(to) {
		return
	}
	if m.DynamicLights[lightSource.Pos] == lightSource {
		delete(m.DynamicLights, lightSource.Pos)
	}
	lightSource.Pos = to
	m.DynamicLights[to] = lightSource
	m.UpdateDynamicLights()
}

// RemoveDynamicLightSource unregisters the light from this map and recasts the dynamic lights.
func (m *GridMap[ActorType, ItemType, ObjectType]) RemoveDynamicLightSource(lightSource *LightSource) {
	if m.DynamicLights[lightSource.Pos] != lightSource {
		return
	}
	delete(m.DynamicLights, lightSource.Pos)
	m.UpdateDynamicLights()
}

// LightReach returns the cells a light at origin can reach; walls and actors cast shadows.
func (m *GridMap[ActorType, ItemType, ObjectType]) LightReach(origin geometry.Point, radius int) []geometry.Point {
	return m.lightfov.SSCVisionMap(origin, radius, true, func(p geometry.Point) bool {
		return p == origin || (m.IsTransparent(p) && !m.IsActorAt(p))
	})
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IndoorLightAt(p geometry.Point) fxtools.HDRColor {
	bakedLighting := m.cells[p.X+p.Y*m.mapWidth].BakedLighting
	if m.meta.IndoorAmbientLight.Brightness() > bakedLighting.Brightness() {
		bakedLighting = m.meta.IndoorAmbientLight
	}

	if dynamicLightAt, ok := m.dynamicallyLitCells[p]; ok {
		if dynamicLightAt.Brightness() > bakedLighting.Brightness() {
			return dynamicLightAt
		}
		return bakedLighting
	}
	return bakedLighting
}

func (m *GridMap[ActorType, ItemType, ObjectType]) OutdoorLightAt(p geometry.Point, timeOfDay time.Time) fxtools.HDRColor {
	ambientLight := GetAmbientLightFromDayTime(timeOfDay)
	bakedLighting := m.cells[p.X+p.Y*m.mapWidth].BakedLighting
	if ambientLight.Brightness() > bakedLighting.Brightness() {
		bakedLighting = ambientLight
	}
	if dynamicLightAt, ok := m.dynamicallyLitCells[p]; ok {
		if dynamicLightAt.Brightness() > bakedLighting.Brightness() {
			return dynamicLightAt
		}
		return bakedLighting
	}
	return bakedLighting
}

func (m *GridMap[ActorType, ItemType, ObjectType]) ZoneLightAt(p geometry.Point, timeOfDay time.Time, visible bool, zone ZoneMetadata) fxtools.HDRColor {
	var bakedLighting fxtools.HDRColor
	if m.meta.IsOutdoor {
		ambientLight := GetAmbientLightFromDayTime(timeOfDay)
		bakedLighting = m.cells[p.X+p.Y*m.mapWidth].BakedLighting
		if ambientLight.Brightness() > bakedLighting.Brightness() {
			bakedLighting = ambientLight
		}
	} else {
		ambientLight := m.meta.IndoorAmbientLight
		bakedLighting = m.cells[p.X+p.Y*m.mapWidth].BakedLighting
		if ambientLight.Brightness() > bakedLighting.Brightness() {
			bakedLighting = ambientLight
		}
	}

	if dynamicLightAt, ok := m.dynamicallyLitCells[p]; ok {
		if dynamicLightAt.Brightness() > bakedLighting.Brightness() {
			bakedLighting = dynamicLightAt
		}
	}

	zoneLighting := zone.Lighting

	if zoneLighting.Brightness() > bakedLighting.Brightness() {
		return zoneLighting
	}
	return bakedLighting
}

func GetAmbientLightFromDayTime(timeOfDay time.Time) fxtools.HDRColor {
	morning := fxtools.HDRColor{R: 0.4, G: 0.368, B: 0.3466666666666667, A: 1.0}
	noon := fxtools.HDRColor{R: 1.8, G: 1.7966666666666666, B: 1.7666666666666666, A: 1.0}
	evening := fxtools.HDRColor{R: 1.4177777777777778, G: 1.4177777777777778, B: 1.46666666666666666, A: 1.0}
	night := fxtools.HDRColor{R: 0.2111111111111111, G: 0.2111111111111111, B: 0.33333333333333333, A: 1.0}

	secondsSinceMidnight := timeOfDay.Hour()*3600 + timeOfDay.Minute()*60 + timeOfDay.Second()

	// we need to determine the percentage of the interval between the two times
	// for example, if it's 9:00, we need to know how far we are between 6:00 and 12:00
	// 9:00 is 50% of the way between 6:00 and 12:00
	// BUT: we want to be precise, so we need to know how many seconds are in the interval
	// 6:00 - 12:00 is 6 hours, so 6 * 3600 = 21600 seconds
	// that means we got these intervals:
	// 0:00 - 6:00 = 0 - 21600
	// 6:00 - 12:00 = 21600 - 43200
	// 12:00 - 18:00 = 43200 - 64800
	// 18:00 - 24:00 = 64800 - 86400

	var ambientLightColor fxtools.HDRColor
	startColor := morning
	endColor := noon
	// 6:00 - 12:00
	if secondsSinceMidnight >= 21600 && secondsSinceMidnight < 43200 {
		intervalPercentage := float64(secondsSinceMidnight-21600) / 21600.0
		ambientLightColor = startColor.Lerp(endColor, intervalPercentage)
	} else if secondsSinceMidnight >= 43200 && secondsSinceMidnight < 64800 {
		startColor = noon
		endColor = evening
		intervalPercentage := float64(secondsSinceMidnight-43200) / 21600.0
		ambientLightColor = startColor.Lerp(endColor, intervalPercentage)
	} else if secondsSinceMidnight >= 64800 && secondsSinceMidnight < 86400 {
		startColor = evening
		endColor = night
		intervalPercentage := float64(secondsSinceMidnight-64800) / 21600.0
		ambientLightColor = startColor.Lerp(endColor, intervalPercentage)
	} else if secondsSinceMidnight >= 0 && secondsSinceMidnight < 21600 {
		startColor = night
		endColor = morning
		intervalPercentage := float64(secondsSinceMidnight) / 21600.0
		ambientLightColor = startColor.Lerp(endColor, intervalPercentage)
	}
	return ambientLightColor
}

// we use the value stored in cell.Lighting for lighting the tile later on..
func (m *GridMap[ActorType, ItemType, ObjectType]) UpdateDynamicLights() {
	clear(m.dynamicallyLitCells)
	if len(m.DynamicLights) == 0 {
		return
	}
	setLightAt := func(point geometry.Point, light fxtools.HDRColor) {
		if light.R > 0 || light.G > 0 || light.B > 0 {
			m.dynamicallyLitCells[point] = light
		}
	}
	m.updateLightMap(m.DynamicLights, setLightAt)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) UpdateBakedLights() {
	setLightAt := func(point geometry.Point, light fxtools.HDRColor) {
		m.cells[point.X+point.Y*m.mapWidth].BakedLighting = light
	}
	m.updateLightMap(m.BakedLights, setLightAt)
}
func (m *GridMap[ActorType, ItemType, ObjectType]) updateLightMap(lightSources map[geometry.Point]*LightSource, setLightAt func(p geometry.Point, light fxtools.HDRColor)) {
	if m.lightScratch == nil {
		m.lightScratch = make(map[geometry.Point]fxtools.HDRColor)
	}
	lightAt := m.lightScratch
	clear(lightAt)
	for _, lightSource := range lightSources {
		for _, pos := range m.LightReach(lightSource.Pos, lightSource.Radius) {
			dist := geometry.Distance(lightSource.Pos, pos)
			if dist > float64(lightSource.Radius) {
				continue
			}
			light := lightSource.ColorAt(dist)
			if existing, has := lightAt[pos]; !has || lightReplaces(existing, light) {
				lightAt[pos] = light
			}
		}
	}
	for pos, light := range lightAt {
		setLightAt(pos, light)
	}
}

// lightReplaces decides overlap: colored light beats white, otherwise the brighter one wins.
func lightReplaces(existing, incoming fxtools.HDRColor) bool {
	existingHue, incomingHue := hasHue(existing), hasHue(incoming)
	if existingHue == incomingHue {
		return incoming.Brightness() >= existing.Brightness()
	}
	return incomingHue
}

// hasHue is relative to brightness, so a dimmed (falloff) colored light keeps its hue.
func hasHue(light fxtools.HDRColor) bool {
	hi := max(light.R, light.G, light.B)
	lo := min(light.R, light.G, light.B)
	return hi-lo > 0.1*hi
}
