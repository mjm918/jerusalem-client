package main

import (
	"context"
	"encoding/json"
	"net"
	"time"
)

const NetworkTimeout = 10 * time.Second

// Codec is a type that represents a codec for encoding and decoding data to and from a network connection using JSON format.
// It contains fields for the underlying JSON encoder and decoder, as well as the network connection itself.
// The Codec type provides methods for receiving, sending, and closing the connection.
//
// Usage example:
//
//	conn, _ := net.Dial("tcp", "localhost:8080")
//	codec := NewCodec(conn)
//	defer codec.Close()
//
//	err := codec.Send(data)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	err = codec.RecvTimeout(&result)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	err = codec.Close()
//	if err != nil {
//		log.Fatal(err)
//	}
type Codec struct {
	decoder *json.Decoder
	encoder *json.Encoder
	conn    net.Conn
}

// NewCodec creates a new instance of the Codec struct using the provided net.Conn connection.
// The Codec is responsible for encoding and decoding messages between the client and server.
// It sets the decoder and encoder to use the JSON format and assigns the connection to the codec.
// The returned Codec pointer can be used to send and receive messages.
func NewCodec(conn net.Conn) *Codec {
	return &Codec{
		decoder: json.NewDecoder(conn),
		encoder: json.NewEncoder(conn),
		conn:    conn,
	}
}

// Recv reads a message from the codec's decoder and assigns it to the provided variable.
// It uses a separate goroutine to decode the message, so it can be cancelled using the provided context.
// If the context is cancelled, Recv returns the context error.
// If decoding the message fails, Recv returns the decoding error.
// Usage example: st.Recv(ctx, &msg)
func (d *Codec) Recv(ctx context.Context, v interface{}) error {
	errChan := make(chan error, 1)
	go func() {
		errChan <- d.decoder.Decode(v)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

func (d *Codec) RecvTimeout(v interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), NetworkTimeout)
	defer cancel()
	return d.Recv(ctx, v)
}

// Send sends the given value to the remote connection using the encoder of the Codec.
// It returns an error if the encoding process fails.
func (d *Codec) Send(v interface{}) error {
	return d.encoder.Encode(v)
}

// Close closes the underlying network connection of the Codec and releases any resources associated with it.
// It returns an error if there was a problem closing the connection.
func (d *Codec) Close() error {
	return d.conn.Close()
}
