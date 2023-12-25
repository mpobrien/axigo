package main

import (
	"bufio"
	"errors"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

type Device struct {
	port      serial.Port
	bufReader bufio.Reader
}

func (dvc *Device) Command(params ...string) (string, error) {
	out := strings.Join(params, ",")
	_, err := dvc.port.Write([]byte(out))
	if err != nil {
		return "", err
	}
	_, err = dvc.port.Write([]byte{'\r'})
	if err != nil {
		return "", err
	}
	s, err := dvc.bufReader.ReadString('\n')
	return s, err
}

type Commander interface {
	SteppersOn() error
	SteppersOff() error
	PenUp() error
	PenDown() error
	Move(stepsX, stepsY int, duration time.Duration) error
	Raw(command ...string) (string, error)
}

type deviceCommander struct {
	*Device
}

func (dc *deviceCommander) SteppersOn() error {
	_, err := dc.Command("EM", "1", "1")
	return err
}
func (dc *deviceCommander) SteppersOff() error {
	_, err := dc.Command("EM", "0", "0")
	return err
}

func (dc *deviceCommander) PenUp() error {
	_, err := dc.Command("SP", "1", "0")
	return err
}

func (dc *deviceCommander) PenDown() error {
	_, err := dc.Command("SP", "0", "0")
	return err
}
func (dc *deviceCommander) Move(stepsX, stepsY int, duration time.Duration) error {
	_, err := dc.Command("XM", strconv.Itoa(int(duration.Milliseconds())), strconv.Itoa(stepsX), strconv.Itoa(stepsY))
	return err
}

func (dc *deviceCommander) Raw(command ...string) (string, error) {
	result, err := dc.Command(command...)
	return result, err
}

func findAxiDrawPort() (*enumerator.PortDetails, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, err
	}
	for _, port := range ports {
		if strings.ToUpper(port.VID) == "04D8" && strings.ToUpper(port.PID) == "FD92" {
			return port, nil
		}
	}
	return nil, errors.New("no AxiDraw connection detected.")
}
func OpenDevice() (*Device, error) {
	portInfo, err := findAxiDrawPort()
	if err != nil {
		return nil, err
	}
	port, err := serial.Open(portInfo.Name, &serial.Mode{
		BaudRate: 9600,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	})
	if err != nil {
		return nil, err
	}
	return &Device{port: port, bufReader: *bufio.NewReader(port)}, nil
}
