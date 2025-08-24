package serial

import (
	"bufio"
	"context"
	"io"
	"log"
	"strings"
	"time"

	"go.bug.st/serial"
)

type PortListener struct {
	port     serial.Port
	baudRate int
	portName string
	reader   *bufio.Reader
}

func NewPortListener(portName string, baudRate int) *PortListener {
	return &PortListener{
		portName: portName,
		baudRate: baudRate,
	}
}

func (pl *PortListener) Open() error {
	mode := &serial.Mode{
		BaudRate: pl.baudRate,
	}

	port, err := serial.Open(pl.portName, mode)
	if err != nil {
		return err
	}

	pl.port = port
	pl.reader = bufio.NewReader(port)
	return nil
}

func (pl *PortListener) Close() error {
	if pl.port != nil {
		return pl.port.Close()
	}
	return nil
}

func (pl *PortListener) Listen(ctx context.Context, dataChan chan<- string, errorChan chan<- error) {
	defer pl.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			line, err := pl.reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					// Port closed, try to reconnect
					log.Println("Port disconnected, attempting to reconnect...")
					pl.reconnect(ctx)
					continue
				}
				errorChan <- err
				return
			}

			// Trim any trailing whitespace or line endings
			line = strings.TrimSpace(line)
			if line != "" {
				dataChan <- line
			}
		}
	}
}

func (pl *PortListener) reconnect(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := pl.Open()
			if err == nil {
				log.Printf("Successfully reconnected to port %s\n", pl.portName)
				return
			}
			log.Printf("Reconnection attempt failed: %v\n", err)
		}
	}
}

func (pl *PortListener) Name() string {
	return pl.portName
}

func (pl *PortListener) BaudRate() int {
	return pl.baudRate
}
