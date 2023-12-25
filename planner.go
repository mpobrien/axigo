package main

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type Vec2d struct {
	x, y float64
}

func (p Vec2d) String() string {
	return fmt.Sprintf("Point(x=%.2f, y=%.2f)", p.x, p.y)
}

func (p Vec2d) Dot(other Vec2d) float64 {
	return p.x*other.x + p.y*p.y
}

func (p Vec2d) Distance(other Vec2d) float64 {
	return math.Hypot(other.x-p.x, other.y-p.y)
}

func (p Vec2d) DistanceSquared(other Vec2d) float64 {
	x := other.x - p.x
	y := other.y - p.y
	return (x * x) + (y * y)
}

func (p Vec2d) Magnitude() float64 {
	return math.Hypot(p.x, p.y)
}

func (p Vec2d) Normalize() Vec2d {
	mag := p.Magnitude()
	return Vec2d{
		x: p.x / mag,
		y: p.y / mag,
	}
}

func (p Vec2d) LinearInterpolate(to Vec2d, magnitude float64) Vec2d {
	v := to.Subtract(p).Normalize()
	return p.Add(v.Multiply(magnitude))
}

func (p Vec2d) SegmentDistance(v, w Vec2d) float64 {
	l2 := v.DistanceSquared(w)
	if l2 == 0 {
		return p.Distance(v)
	}
	t := ((p.x-v.x)*(w.x-v.x) + (p.y-v.y)*(w.y-v.y)) / l2
	if t > 1 {
		t = 1
	}
	if t < 0 {
		t = 0
	}
	x := v.x + t*(w.x-v.x)
	y := v.y + t*(w.y-v.y)
	return p.Distance(Vec2d{x, y})
}

func (p Vec2d) Add(other Vec2d) Vec2d {
	return Vec2d{x: p.x + other.x, y: p.y + other.y}
}

func (p Vec2d) Multiply(scalar float64) Vec2d {
	return Vec2d{x: p.x * scalar, y: p.y * scalar}
}

func (p Vec2d) Subtract(other Vec2d) Vec2d {
	return Vec2d{
		x: p.x - other.x,
		y: p.y - other.y,
	}
}

func makePlan(points []Vec2d, accel, cornerFactor float64, debug bool) Plan {
	vmax := float64(4)
	thr := throttler{.02, points, .001, vmax, nil}
	thr.init()

	maxVelocities := thr.computeMaxVelocities()
	segments := []Segment{}
	// Make a Segment for each consecutive pair of points
	for i := 1; i < len(points); i++ {
		segments = append(segments, Segment{p1: points[i-1], p2: points[i]})
	}

	// Compute a max entry velocity for each segment
	for i := 0; i < len(segments)-1; i++ {
		in, out := segments[i], segments[i+1]
		if in.maxEntryVelocity > maxVelocities[i] {
			in.maxEntryVelocity = maxVelocities[i]
		}
		out.maxEntryVelocity = cornerVelocity(in, out, vmax, accel, cornerFactor)
	}

	// add a dummy segment at the end to force a final velocity of zero
	segments = append(
		segments,
		Segment{
			p1: points[len(points)-1],
			p2: points[len(points)-1],
		},
	)

	var i int
	for i < len(segments)-1 {
		segment, nextSegment := &segments[i], &segments[i+1]

		vExit := nextSegment.maxEntryVelocity
		segmentLength := segment.p1.Distance(segment.p2)
		profile := newTriangle(
			segmentLength,
			segment.entryVelocity,
			vExit,
			accel,
			segment.p1,
			segment.p2,
		)
		if profile.s1 < -1*EPS {
			if debug {
				fmt.Println("here1")
			}

			// too fast, update max entry vel and backtrack
			segment.maxEntryVelocity = math.Sqrt(vExit*vExit + 2*accel*segmentLength)
			i -= 1
			continue
		}
		if profile.s2 < 0 {
			if debug {
				fmt.Println("here2")
			}
			vf := math.Sqrt(segment.entryVelocity*segment.entryVelocity + 2*accel*segmentLength)
			t := (vf - segment.entryVelocity) / accel
			segment.blocks = []Block{
				{
					accel,
					t,
					segment.entryVelocity,
					segment.p1,
					segment.p2,
				},
			}
			nextSegment.entryVelocity = vf
			i++
			continue
		}
		if profile.vmax > vmax {
			if debug {
				fmt.Println("here3")
			}
			// accelerate, cruise, decelerate
			z := newTrapezoid(segmentLength, segment.entryVelocity, vmax, vExit, accel, segment.p1, segment.p2)
			segment.blocks = []Block{
				{accel, z.t1, segment.entryVelocity, z.p1, z.p2},
				{0, z.t2, vmax, z.p2, z.p3},
				{-accel, z.t3, vmax, z.p3, z.p4},
			}
			nextSegment.entryVelocity = vExit
			i++
			continue
		}
		if debug {
			fmt.Println("here4")
		}
		segment.blocks = []Block{
			{accel, profile.t1, segment.entryVelocity, profile.p1, profile.p2},
			{-1 * accel, profile.t2, profile.vmax, profile.p2, profile.p3},
		}
		nextSegment.entryVelocity = vExit
		i++
	}
	var blocks []Block
	var totalTime float64
	for _, s := range segments {
		blocks = append(blocks, s.blocks...)
		for _, b := range s.blocks {
			//fmt.Printf("%d, %d: %#v\n", i, j, b)
			totalTime += b.t
		}
	}
	return Plan{
		blocks:    blocks,
		totalTime: totalTime,
	}
}

type throttler struct {
	deltaTime   float64
	points      []Vec2d
	threshold   float64
	maxVelocity float64

	distances []float64
}

func (th *throttler) computeMaxVelocities() []float64 {
	out := make([]float64, len(th.points))
	for i := 0; i < len(th.points); i++ {
		out[i] = th.computeMaxVelocity(i)
	}
	return out
}

// isFeasible returns true if moving between points (from index to index+1)
// can be done at the given velocity without overshooting the next point.
func (th throttler) isFeasible(index int, velocity float64) bool {
	incrementalDistance := velocity * th.deltaTime
	dist0 := th.distances[index]
	proposedDistance := dist0 + incrementalDistance
	indexAtTarget := th.lookup(proposedDistance)
	if indexAtTarget == index {
		return true
	}
	p0, p1 := th.points[index], th.points[indexAtTarget]
	var p11 Vec2d
	if indexAtTarget+1 >= len(th.points) {
		p11 = p1
	} else {
		p11 = th.points[indexAtTarget+1]
	}

	s := proposedDistance - th.distances[indexAtTarget]
	interpolatedPoint := p1.LinearInterpolate(p11, s)
	nextIndex := index + 1
	for nextIndex < indexAtTarget {
		p := th.points[nextIndex]
		if p.SegmentDistance(p0, interpolatedPoint) > th.threshold {
			return false
		}
		nextIndex++
	}
	return true
}

func (th *throttler) lookup(distance float64) int {
	// find the first index in distances that is smaller than distance
	return sort.SearchFloat64s(th.distances, distance) - 1
}
func (th *throttler) computeMaxVelocity(index int) float64 {
	if th.isFeasible(index, th.maxVelocity) {
		return th.maxVelocity
	}
	low, high := float64(0), th.maxVelocity
	var velocity float64
	for i := 0; i < 16; i++ {
		velocity = (low + high) / 2
		if th.isFeasible(index, velocity) {
			low = velocity
		} else {
			high = velocity
		}
	}
	return low
}

func (th *throttler) init() {
	if len(th.distances) > 0 {
		return
	}
	th.distances = make([]float64, 0, len(th.points))
	totalDistance := float64(0)
	prevPoint := th.points[0]
	for i := 0; i < len(th.points); i++ {
		totalDistance += th.points[i].Distance(prevPoint)
		th.distances = append(th.distances, totalDistance)
		prevPoint = th.points[i]
	}
}

type Plan struct {
	blocks      []Block
	totalTime   float64
	totalLength float64
}

func newPlan(blocks []Block) Plan {
	totalTime := time.Duration(0)
	totalLength := float64(0)
	//for _, b := range blocks {
	//totalTime += b.t

	//}
	_ = totalLength
	_ = totalTime

	return Plan{}
}

type Block struct {
	accel      float64
	t          float64
	velocity   float64
	start, end Vec2d
}

type Segment struct {
	p1, p2           Vec2d
	maxEntryVelocity float64
	entryVelocity    float64
	blocks           []Block
}

func (s Segment) Vec() Vec2d {
	return s.p2.Subtract(s.p1).Normalize()
}

const EPS = 1e-9

func cornerVelocity(s1, s2 Segment, vmax, accel, cornerFactor float64) float64 {
	cosine := -1 * s1.Vec().Dot(s2.Vec())
	if cosine < EPS {
		return 0
	}
	sine := math.Sqrt((1 - cosine) / 2)
	if math.Abs(sine-1) < EPS {
		return vmax
	}
	v := math.Sqrt((accel * cornerFactor * sine) / (1 - sine))
	if v > vmax {
		return vmax
	}
	return v
}

type Triangle struct {
	s1, s2, t1, t2, vmax float64
	p1, p2, p3           Vec2d
}

func newTriangle(s, vi, vf, a float64, p1, p3 Vec2d) Triangle {
	s1 := (2*a*s + vf*vf - vi*vi) / (4 * a)
	s2 := s - s1

	vmax := math.Pow(vi*vi+2*a*s1, 0.5)
	t1 := (vmax - vi) / a
	t2 := (vf - vmax) / (-1 * a)
	p2 := p1.LinearInterpolate(p3, s1)
	return Triangle{s1, s2, t1, t2, vmax, p1, p2, p3}
}

type Trapezoid struct {
	s1, s2, s3, t1, t2, t3 float64
	p1, p2, p3, p4         Vec2d
}

func newTrapezoid(s, vi, vmax, vf, a float64, p1, p4 Vec2d) Trapezoid {
	// compute a trapezoidal profile: accelerating, cruising, decelerating
	t1 := (vmax - vi) / a
	s1 := (vmax + vi) / 2 * t1
	t3 := (vf - vmax) / -a
	s3 := (vf + vmax) / 2 * t3
	s2 := s - s1 - s3
	t2 := s2 / vmax
	p2 := p1.LinearInterpolate(p4, s1)
	p3 := p1.LinearInterpolate(p4, s-s3)
	return Trapezoid{s1, s2, s3, t1, t2, t3, p1, p2, p3, p4}
}
