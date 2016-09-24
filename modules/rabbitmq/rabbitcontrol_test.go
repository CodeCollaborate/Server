package rabbitmq

import (
	"fmt"
	"testing"
	"time"
)

func TestNewControl(t *testing.T) {
	control := NewControl()
	doneTesting := make(chan bool, 1)
	received := make(chan bool, 1)
	defer close(doneTesting)
	defer close(received)

	timeout := make(chan bool, 1)
	defer close(timeout)

	subsci := Subscription{
		Channel:     "12345",
		IsSubscribe: true,
	}

	go func() {
		for {
			select {
			case sub := <-control.SubChan:
				if sub == subsci {
					received <- true
				} else {
					t.Fatal("Message was corrupted")
				}
			case <-control.Exit:
				fmt.Println("All jobs completed")
				doneTesting <- true
				return
			}
		}
	}()

	control.SubChan <- subsci

	select {
	case <-received:
	// success
	case <-time.After(time.Second * 5):
		t.Fatal("Control signal timed out")
	}

	control.Exit <- true
	select {
	case <-doneTesting:
	// success
	case <-time.After(time.Second * 5):
		t.Fatal("Control signal timed out")
	}

}
