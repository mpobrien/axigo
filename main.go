package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/chzyer/readline"
)

const (
	stepsPerInch               = 2032
	stepsPerMillimeter         = 80
	defaultSpeedStepsPerSecond = 2032
)

func executeCommands(input string, dvc *Device) error {
	cmds := strings.Split(input, ";")
	for _, cmd := range cmds {
		cmd = strings.TrimSpace(strings.ToLower(cmd))
		cmdParts := strings.Fields(cmd)
		if len(cmdParts) == 0 {
			continue
		}
		switch cmdParts[0] {
		case "on":
			_, err := dvc.Command("EM", "1", "1")
			if err != nil {
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
			_, err := dvc.Command("SP", "1", "0")
			if err != nil {
				return err
			}
			continue
		case "pendown":
			_, err := dvc.Command("SP", "0", "0")
			if err != nil {
				return err
			}
			continue
		case "off":
			_, err := dvc.Command("EM", "0", "0")
			if err != nil {
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

			_, err = dvc.Command("XM", strconv.Itoa(int(duration.Milliseconds())), strconv.Itoa(xMove), strconv.Itoa(yMove))
			if err != nil {
				return err
			}
			continue
		case "raw":
			log.Printf("executing [%s]", strings.Join(cmdParts[1:], " "))
			result, err := dvc.Command(cmdParts[1:]...)
			log.Printf("result: %s", result)
			if err != nil {
				return err
			}
			continue
		default:
			log.Printf("unknown command: %s", cmdParts[0])
		}
	}

	return nil
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

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}
		if err := executeCommands(line, dev); err != nil {
			log.Printf("error: %s", err)
		}
	}
}
