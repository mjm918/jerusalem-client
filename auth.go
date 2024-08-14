package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
)

// Authenticator represents an object responsible for handling client authentication and generating and validating answers.
//
// Fields:
// - k []byte: the secret key used for generating answers and validating answers.
// - pr *RangeInclusive: the range of ports used for finding free ports during authentication.
// - db *TcpClientRepository: the repository for accessing TCP client data.
//
// Methods:
// - GenerateAnswer(ch uuid.UUID) string: generates an answer for a challenge.
// - ValidateAnswer(ch uuid.UUID, ans string) bool: validates an answer for a challenge.
// - PerformServerHandshake(stream *Codec) error: performs the server handshake with the client.
// - handleClientAuth(stream *Codec, id string) error: handles the client authentication process.
// - findFreePort() (uint16, error): finds a free port within the specified range.
type Authenticator struct {
	k []byte
}

// NewAuthenticator creates a new instance of the Authenticator struct and initializes it with the provided parameters.
// The secret parameter is used to generate the authentication key, which is a SHA-256 hash of the secret.
// The db parameter is a reference to a TcpClientRepository, which is used to interact with the client database.
// The pr parameter is a reference to a RangeInclusive struct, which represents the range of ports that can be used.
// The function returns a pointer to the newly created Authenticator instance.
func NewAuthenticator(secret string) *Authenticator {
	h := sha256.Sum256([]byte(secret))
	return &Authenticator{k: h[:]}
}

// GenerateAnswer generates an answer using the HMAC-SHA256 algorithm.
// It takes a uuid.UUID as a challenge, appends it to the key provided during
// Authenticator initialization, and computes the HMAC-SHA256 hash. The result
// is then encoded to a hexadecimal string and returned.
func (a *Authenticator) GenerateAnswer(ch uuid.UUID) string {
	m := hmac.New(sha256.New, a.k)
	m.Write(ch[:])
	return hex.EncodeToString(m.Sum(nil))
}

// ValidateAnswer validates the answer provided by the client for a challenge.
// It decodes the answer from a hex-string to bytes and computes the HMAC of
// the challenge using the provided key. Then it checks if the computed HMAC
// is equal to the decoded answer. Returns true if the answer is valid, false
// otherwise.
func (a *Authenticator) ValidateAnswer(ch uuid.UUID, ans string) bool {
	b, err := hex.DecodeString(ans)
	if err != nil {
		return false
	}
	m := hmac.New(sha256.New, a.k)
	m.Write(ch[:])
	em := m.Sum(nil)
	return hmac.Equal(em, b)
}

// PerformClientHandshake answers a challenge to attempt to authenticate with the server.
func (a *Authenticator) PerformClientHandshake(stream *Codec, clientId string) (uint16, error) {
	var msg ServerMessage
	ctx, cancel := context.WithTimeout(context.Background(), NetworkTimeout)
	defer cancel()

	if err := stream.Recv(ctx, &msg); err != nil {
		return 0, err
	}

	if msg.Type != MtChallenge {
		return 0, fmt.Errorf("no secret provided / invalid secret key")
	}

	answer := a.GenerateAnswer(msg.Challenge)
	if err := stream.Send(ClientMessage{Type: MtAuthenticate, Authenticate: answer, ClientId: clientId}); err != nil {
		return 0, err
	}

	if err := stream.Recv(ctx, &msg); err != nil {
		return 0, err
	}

	if msg.Type != MtFreePort {
		return 0, fmt.Errorf("rejection response from server")
	}

	return msg.Port, nil
}
