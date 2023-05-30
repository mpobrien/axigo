package main

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
}type Vec2d struct {
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