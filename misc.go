package main

import (
	"github.com/google/uuid"
)

const (
	MtChallenge    = "Challenge"
	MtHeartbeat    = "Heartbeat"
	MtConnection   = "Connection"
	MtAuthenticate = "Authenticate"
	MtFreePort     = "FreePort"
	MtHello        = "Hello"
	MtError        = "Error"
)

type ClientMessage struct {
	Type         string    `json:"type"`
	Authenticate string    `json:"authenticate,omitempty"`
	Port         uint16    `json:"port,omitempty"`
	Accept       uuid.UUID `json:"accept,omitempty"`
	ClientId     string    `json:"clientId,omitempty"`
}

type ServerMessage struct {
	Type       string    `json:"type"`
	Challenge  uuid.UUID `json:"challenge,omitempty"`
	Port       uint16    `json:"hello,omitempty"`
	Heartbeat  bool      `json:"heartbeat,omitempty"`
	Connection uuid.UUID `json:"connection,omitempty"`
	Error      string    `json:"error,omitempty"`
}
