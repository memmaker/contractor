package gridmap

import (
	"testing"

	"github.com/memmaker/go/fxtools"
)

func TestLightBlendAndFalloff(t *testing.T) {
	white := fxtools.HDRColor{R: 1, G: 1, B: 1, A: 1}
	red := fxtools.HDRColor{R: 1, G: 0.2, B: 0.2, A: 1}
	if !lightReplaces(white, red.MultiplyWithScalar(0.1)) {
		t.Error("dim colored light must keep its hue and beat white")
	}
	if lightReplaces(white, white.MultiplyWithScalar(0.5)) {
		t.Error("dimmer white must not replace brighter white")
	}

	l := &LightSource{Radius: 5, Color: white, MaxIntensity: 1}
	defer func(old bool) { LightFalloff = old }(LightFalloff)
	LightFalloff = false
	if l.ColorAt(5).R != 1 {
		t.Error("flat light must be full intensity at the edge")
	}
	LightFalloff = true
	if c := l.ColorAt(5).R; c <= 0 || c >= l.ColorAt(1).R {
		t.Errorf("falloff must dim with distance but stay lit at the edge, got %v", c)
	}
}
