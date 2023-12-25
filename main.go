package main

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/fogleman/gg"
)

const (
	stepsPerInch               = 2032
	stepsPerMillimeter         = 80
	defaultSpeedStepsPerSecond = 2032
)

func readEvalPrint(input string, cmdr Commander) error {
	cmds := strings.Split(input, ";")
	for _, cmd := range cmds {
		cmd = strings.TrimSpace(strings.ToLower(cmd))
		cmdParts := strings.Fields(cmd)
		if len(cmdParts) == 0 {
			continue
		}
		switch cmdParts[0] {
		case "on":
			if err := cmdr.SteppersOn(); err != nil {
				return err
			}
			continue
		case "sleep":
			if len(cmdParts[1:]) != 1 {
				return fmt.Errorf("incorrect param count to 'sleep'")
			}
			sleepTime, err := time.ParseDuration(cmdParts[1])
			if err != nil {
				return fmt.Errorf("invalid sleep duration: %w", err)
			}
			time.Sleep(sleepTime)
			continue
		case "penup":
			if err := cmdr.PenUp(); err != nil {
				return err
			}
			continue
		case "pendown":
			if err := cmdr.PenDown(); err != nil {
				return err
			}
			continue
		case "off":
			if err := cmdr.SteppersOff(); err != nil {
				return err
			}
			continue
		case "move":
			if len(cmdParts[1:]) != 2 {
				return fmt.Errorf("incorrect param count to 'move'")
			}
			xMove, err := strconv.Atoi(cmdParts[1])
			if err != nil {
				return fmt.Errorf("invalid param to 'move': %s", err)
			}
			yMove, err := strconv.Atoi(cmdParts[2])
			if err != nil {
				return fmt.Errorf("invalid param to 'move': %s", err)
			}

			distance := math.Sqrt(float64(xMove*xMove + yMove*yMove))
			fmt.Println("distance is", distance)
			duration := time.Duration(int(float64(distance/defaultSpeedStepsPerSecond) * float64(time.Second)))
			fmt.Println("duration is", duration)

			if err := cmdr.Move(xMove, yMove, duration); err != nil {
				return err
			}
			continue
		case "raw":
			log.Printf("executing [%s]", strings.Join(cmdParts[1:], " "))
			result, err := cmdr.Raw(cmdParts[1:]...)
			log.Printf("result: %s", result)
			if err != nil {
				return err
			}
			continue
		case "text":
			if len(cmdParts[1:]) != 1 {
				return fmt.Errorf("incorrect param count to 'text'")
			}
			textPaths := text(cmdParts[1], FontAstrology)
			d := newDrawing(textPaths)

			for i, path := range d.paths {
				plan := makePlan([]Vec2d(path.Path), float64(16), 0.001, i == 3)
				for j, block := range plan.blocks {
					fmt.Printf(
						"%d, %d: a=%.2f, t=%.2f, vi=%.2f, p1=%s, p2=%s\n",
						i,
						j,
						block.accel,
						block.t,
						block.velocity,
						block.start,
						block.end,
					)
					//print(f"{pathindex}, {blockindex}: a={b.a}, t={b.t}, vi={b.vi}, p1={b.p1}, p2={b.p2}, s={b.s}")

				}
			}
			stats := d.Stats()
			fmt.Printf("stats: %#v\n", stats)
			encoded, err := d.Render()
			if err != nil {
				return err
			}

			f, err := os.CreateTemp("", "*.png")
			if err != nil {
				return err
			}
			if _, err := f.Write(encoded.Bytes()); err != nil {
				return err
			}
			fmt.Println(f.Name())
			f.Close()

			fmt.Println("png is", len(encoded.Bytes()))
			if err := imgcat(bytes.NewReader(encoded.Bytes()), os.Stdout); err != nil {
				return err
			}

			continue
		default:
			log.Printf("unknown command: %s", cmdParts[0])
		}
	}

	return nil
}

func PlotDrawing(d Drawing) {
	timesliceMs := 10
	stepMs := timesliceMs
	_ = stepMs
	//stepSec := float64(stepMs) / 1000

	for _, path := range d.paths {
		plan := makePlan(
			path.Path,
			float64(16), //accel
			0.001,       //corner factor
			false,
		)
		t := 0
		for t < int(plan.totalTime) {
			break
		}
	}
	/*
	   step_ms = TIMESLICE_MS
	   step_s = step_ms / 1000
	   t = 0
	   while t < plan.t:
	       i1 = plan.instant(t)
	       i2 = plan.instant(t + step_s)
	       d = i2.p.sub(i1.p)
	       ex, ey = self.error
	       ex, sx = modf(d.x * self.steps_per_unit + ex)
	       ey, sy = modf(d.y * self.steps_per_unit + ey)
	       self.error = ex, ey
	       self.stepper_move(step_ms, int(sx), int(sy))
	       t += step_s
	   # self.wait()
	*/
}

func newDrawing(paths []Path) Drawing {
	out := Drawing{}
	prevPosition := Vec2d{0, 0} // start at origin
	for _, path := range paths {
		if len(path) == 0 {
			continue
		}
		out.paths = append(
			out.paths,
			PenPath{
				Path{
					prevPosition,
					path[0],
				},
				true,
			},
		)
		out.paths = append(out.paths, PenPath{path, false})
		prevPosition = path[len(path)-1]
	}

	return out
}

func main() {
	dev, err := OpenDevice()
	if err != nil {
		log.Fatalf("failed to open device: %s", err)
	}
	_ = dev

	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt)
	// go func() {
	// 	for _ = range c {
	// 		fmt.Println("got ctrl c!")
	// 		//if _, err := dev.Command("EM", "0", "0"); err != nil {
	// 		//log.Printf("failed to disable motors on shutdown: %s", err)
	// 		//}
	// 		//os.Exit(0)
	// 	}
	// }()

	rl, err := readline.New("> ")
	if err != nil {
		log.Fatalf("failed to open readline: %s", err)
	}
	defer rl.Close()
	readline.CaptureExitSignal(func() {
		fmt.Println("captured exit signal!")
	})

	commander := &deviceCommander{dev}

	for {
		line, err := rl.Readline()
		if err != nil {
			commander.SteppersOff()
			break
		}
		if err := readEvalPrint(line, commander); err != nil {
			log.Printf("error: %s", err)
		}
	}
}

func text(input string, font Font) []Path {
	const spacing = 0
	var out []Path
	x := 0
	for _, ch := range input {
		index := int(ch) - 32
		if index < 0 || index > len(font) {
			x += spacing
			continue
		}
		glyph := font[index]
		for _, path := range glyph.paths {
			var newPath = make(Path, 0, len(path))
			for _, point := range path {
				newPath = append(
					newPath,
					Vec2d{
						x: float64(x) + point.x - float64(glyph.left),
						y: point.y,
					})
			}
			out = append(out, newPath)
		}
		x += glyph.right - glyph.left + spacing

	}
	return out
}

type Drawing struct {
	paths []PenPath
}

type DrawingStats struct {
	UpLength   float64
	DownLength float64
}

func (d Drawing) Stats() DrawingStats {
	var upLength, downLength float64
	for i, path := range d.paths {
		if len(path.Path) <= 1 {
			continue
		}
		var length float64
		for i := 1; i < len(path.Path); i++ {
			p0, p1 := path.Path[i-1], path.Path[i]
			length += math.Hypot(p1.x-p0.x, p1.y-p0.y)
		}
		if path.penUp {
			if i == 0 {
				continue
			}
			upLength += length
			continue
		}
		downLength += length
	}
	return DrawingStats{
		UpLength:   upLength,
		DownLength: downLength,
	}
}

func (d Drawing) Render() (*bytes.Buffer, error) {
	margin := float64(10) // TODO
	topLeft, bottomRight := Bounds(d.paths)
	dc := gg.NewContext(int(bottomRight.x-topLeft.x+2*margin), int(bottomRight.y-topLeft.y+2*margin))
	translation := Vec2d{-1 * topLeft.x, -1 * topLeft.y}
	dc.Clear()
	dc.SetColor(color.White)
	dc.MoveTo(0, 0)
	dc.DrawRectangle(0, 0, float64(dc.Width()), float64(dc.Height()))
	dc.Fill()

	dc.SetColor(color.Black)
	dc.SetLineWidth(1.5)
	dc.SetLineCap(gg.LineCapRound)

	for _, p := range d.paths {
		for _, point := range p.Path {
			if p.penUp {
				dc.MoveTo(
					point.x+translation.x+margin,
					point.y+translation.y+margin,
				)
			} else {
				dc.LineTo(
					point.x+translation.x+margin,
					point.y+translation.y+margin,
				)
			}
		}
		dc.Stroke()
	}
	out := &bytes.Buffer{}
	if err := dc.EncodePNG(out); err != nil {
		return nil, err
	}
	return out, nil
}

type PenPath struct {
	Path
	penUp bool
}

// Bounds returns points representing the upper left and lower right corner
func Bounds(paths []PenPath) (Vec2d, Vec2d) {
	var xMin, xMax, yMin, yMax float64
	for _, path := range paths {
		for _, point := range path.Path {
			if point.x < xMin {
				xMin = point.x
			} else if point.x > xMax {
				xMax = point.x
			}
			if point.y < yMin {
				yMin = point.y
			} else if point.y > yMax {
				yMax = point.y
			}
		}
	}
	return Vec2d{xMin, yMin}, Vec2d{xMax, yMax}
}
