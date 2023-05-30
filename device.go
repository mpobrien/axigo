package main

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

type Device struct {
	port      serial.Port
	bufReader bufio.Reader
}

func (dvc *Device) Command(params ...string) (string, error) {
	out := strings.Join(params, ",")
	fmt.Println("sending command: [", out, "]")
	_, err := dvc.port.Write([]byte(out))
	if err != nil {
		return "", err
	}
	_, err = dvc.port.Write([]byte{'\r'})
	if err != nil {
		return "", err
	}
	fmt.Println("waiting for reply")
	s, err := dvc.bufReader.ReadString('\n')
	fmt.Println("got reply", s)
	return s, err
}

func findAxiDrawPort() (*enumerator.PortDetails, error) {
	//VID_PID = '04D8:FD92'
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, err
	}
	for _, port := range ports {
		if strings.ToUpper(port.VID) == "04D8" && strings.ToUpper(port.PID) == "FD92" {
			return port, nil
		}
	}
	return nil, errors.New("No AxiDraw connection detected.")
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
