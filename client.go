package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/briandowns/spinner"
	"io"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

const networkTimeout = 2 * time.Minute

// Client is a type that represents a client in a client-server communication system.
//
// Fields:
// - sp uint16: the local port that the client listens on for incoming connections.
// - cc *Codec: the control connection to the server.
// - da string: the destination address of the server.
// - lh string: the local host that is forwarded.
// - lp uint16: the local port that is forwarded.
// - rp uint16: the port that is publicly available on the remote server.
// - auth *Authenticator: an optional secret used to authenticate clients.
// - cid string: the client ID.
//
// Usage example:
//
//	// Create a new client
//	client, err := NewClient(sp, lh, lp, da, cid, s)
//
//	if err != nil {
//	  log.Fatal(err)
//	}
//
//	// Get the remote port
//	rp := client.RemotePort()
//
//	// Listen for server messages
//	err := client.Listen()
//
//	if err != nil {
//	  log.Fatal(err)
//	}
type Client struct {
	sp   uint16
	cc   *Codec         // Control connection to the server.
	da   string         // Destination address of the server.
	lh   string         // Local host that is forwarded.
	lp   uint16         // Local port that is forwarded.
	rp   uint16         // Port that is publicly available on the remote.
	auth *Authenticator // Optional secret used to authenticate clients.
	cid  string
}

// NewClient creates a new instance of the Client struct and initializes it with the provided parameters.
// It establishes a connection with the server at the specified destination address and port
// and performs a client handshake to authenticate with the server.
// If the handshake is successful, it sends a hello message to the server.
// It then receives and processes the initial server message, which includes the remote port that
// is publicly available on the remote server.
// If all steps are successful, it returns a pointer to the newly created Client instance.
// Otherwise, it returns an error.
func NewClient(sp uint16, lh string, lp uint16, da, cid, s string) (*Client, error) {
	conn, err := establishConnectionWithTimeout(da, sp)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", da, err)
	}

	cc := NewCodec(conn)
	auth := NewAuthenticator(s)

	destPort, err := auth.PerformClientHandshake(cc, cid)
	if err != nil {
		return nil, fmt.Errorf("client handshake failed: %w", err)
	}

	if err := cc.Send(ClientMessage{Type: MtHello, Port: destPort}); err != nil {
		return nil, fmt.Errorf("failed to send hello message: %w", err)
	}

	var msg ServerMessage
	ctx, cancel := context.WithTimeout(context.Background(), NetworkTimeout)
	defer cancel()

	err = cc.Recv(ctx, &msg)
	if err != nil {
		return nil, fmt.Errorf("failed to receive server message: %w", err)
	}

	rp, err := processInitialServerMessage(msg)
	if err != nil {
		return nil, err
	}

	log.Printf("Connected to server at %s:%d\n", da, rp)
	log.Printf("Listening for connection to redirect\n\n")

	return &Client{
		sp:   sp,
		cc:   cc,
		da:   da,
		lh:   lh,
		lp:   lp,
		rp:   rp,
		auth: auth,
		cid:  cid,
	}, nil
}

// RemotePort returns the port that is publicly available on the remote server.
func (c *Client) RemotePort() uint16 {
	return c.rp
}

// Listen listens for server messages and processes them accordingly.
// It continuously receives messages from the server using the connection's Recv method.
// If there is an error receiving a message, it returns an error message.
// If there is an error processing a server message, it returns the error.
// The method runs indefinitely until there is an error or the connection is closed.
// The method uses the processServerMessage method to handle the different types of server messages.
// If there is an error receiving a message or processing a server message, the method exits and returns the error.
// The method returns nil if the connection is closed gracefully.
func (c *Client) Listen() error {
	for {
		s := spinner.New(spinner.CharSets[39], 100*time.Millisecond)
		s.Start()
		var msg ServerMessage
		if err := c.cc.Recv(context.Background(), &msg); err != nil {
			return fmt.Errorf("failed to receive server message: %w", err)
		}

		if err := c.processServerMessage(msg); err != nil {
			return err
		}
		s.Stop()
	}
}

// processServerMessage processes a server message received by the client.
// Depending on the message type, it performs different actions:
//
//   - MtHello: Prints an unexpected hello message.
//   - MtChallenge: Prints an unexpected challenge message.
//   - MtHeartbeat: Does nothing.
//   - MtConnection: Establishes a connection with the server in a separate goroutine using the received connection ID.
//     If the connection is established successfully, it prints "Connection closed gracefully" when it's closed.
//     If there is an error, it prints "Connection exited with error: <error>".
//   - MtError: Returns an error with the server error message.
//   - Default: Returns an error with the unexpected message type.
//
// It returns nil if the message is processed successfully.
func (c *Client) processServerMessage(msg ServerMessage) error {
	switch msg.Type {
	case MtHello:
		log.Println("Received an unexpected hello message")
	case MtChallenge:
		log.Println("Received an unexpected challenge message")
	case MtHeartbeat:
		// Do nothing
	case MtConnection:
		id := msg.Connection
		go func() {
			if err := c.establishConnectionRoutine(id); err != nil {
				log.Printf("Connection exited with error: %v\n", err)
			} else {
				log.Println("Connection closed gracefully")
			}
		}()
	case MtError:
		return fmt.Errorf("server error: %s", msg.Error)
	default:
		return fmt.Errorf("received unexpected message type: %s", msg.Type)
	}
	return nil
}

// establishConnectionRoutine establishes a connection with the server and performs
// the necessary handshakes for authentication. It then sends an "Accept" message
// with the provided ID to the server. It also establishes a connection with the
// local host and sets up bidirectional data transfer between the server and the
// local host. This function returns an error if any step in the process fails.
func (c *Client) establishConnectionRoutine(id uuid.UUID) error {
	conn, err := establishConnectionWithTimeout(c.da, c.sp)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", c.da, err)
	}
	defer conn.Close()

	rc := NewCodec(conn)
	if c.auth != nil {
		if _, err := c.auth.PerformClientHandshake(rc, c.cid); err != nil {
			return fmt.Errorf("client handshake failed: %w", err)
		}
	}

	if err := rc.Send(ClientMessage{Type: "Accept", Accept: id}); err != nil {
		return fmt.Errorf("failed to send accept message: %w", err)
	}

	lconn, err := establishConnectionWithTimeout(c.lh, c.lp)
	if err != nil {
		return fmt.Errorf("failed to connect to local host %s:%d: %w", c.lh, c.lp, err)
	}
	defer lconn.Close()

	eg := new(errgroup.Group)
	eg.Go(func() error {
		_, err := io.Copy(lconn, rc.conn)
		return err
	})
	eg.Go(func() error {
		_, err := io.Copy(rc.conn, lconn)
		return err
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("data transfer failed: %w", err)
	}
	return nil
}

// establishConnectionWithTimeout establishes a TCP connection to the specified address (host:port) with a timeout of 30 seconds.
// It returns a net.Conn object representing the established connection and an error if connection establishment fails.
func establishConnectionWithTimeout(host string, port uint16) (net.Conn, error) {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, networkTimeout)
	if err != nil {
		return nil, fmt.Errorf("could not connect to %s: %w", address, err)
	}
	return conn, nil
}

// processInitialServerMessage processes the initial server message and handles different message types.
// It takes a ServerMessage as input and returns the remote port if the message type is MtHello.
// If the message type is MtError, it returns an error message with the server error.
// If the message type is MtChallenge, it returns an error indicating that the server requires authentication but no client secret was provided.
// For any other message type, it returns an error message with the unexpected message type.
// The function returns both the remote port and an error, if any.
func processInitialServerMessage(msg ServerMessage) (uint16, error) {
	var rp uint16
	switch msg.Type {
	case MtHello:
		rp = msg.Port
	case MtError:
		return 0, fmt.Errorf("server error: %s", msg.Error)
	case MtChallenge:
		return 0, errors.New("server requires authentication, but no client secret was provided")
	default:
		return 0, fmt.Errorf("unexpected initial non-hello message of type: %s", msg.Type)
	}
	return rp, nil
}
