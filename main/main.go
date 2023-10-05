package main

import (
	"context"
	"github.com/kward/go-vnc"
	"log"
	"net"
	"testing"
	"time"

	"github.com/kward/go-vnc/messages"
	"github.com/kward/go-vnc/rfbflags"
)

func TestName(t *testing.T) {

}
func main() {
	// Establish TCP connection to VNC server.
	nc, err := net.Dial("tcp", "10.20.13.17:5900")
	if err != nil {
		log.Fatalf("Error connecting to VNC host. %v", err)
	}

	// Negotiate connection with the server.
	vcc := vnc.NewClientConfig("some_password")
	vc, err := vnc.Connect(context.Background(), nc, vcc)
	if err != nil {
		log.Fatalf("Error negotiating connection to VNC host. %v", err)
	}

	// Periodically request framebuffer updates.
	go func() {
		w, h := vc.FramebufferWidth(), vc.FramebufferHeight()
		for {
			if err := vc.FramebufferUpdateRequest(rfbflags.RFBTrue, 0, 0, w, h); err != nil {
				log.Printf("error requesting framebuffer update: %v", err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// Listen and handle server messages.
	go vc.ListenAndHandle()

	// Process messages coming in on the ServerMessage channel.
	for {
		msg := <-vcc.ServerMessageCh
		switch msg.Type() {
		case messages.FramebufferUpdate:
			log.Println("Received FramebufferUpdate message.")
		default:
			log.Printf("Received message type:%v msg:%v\n", msg.Type(), msg)
		}
	}
}
