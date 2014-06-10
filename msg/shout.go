package msg

import (
	zmq "github.com/pebbe/zmq4"

	"bytes"
	"encoding/binary"
	"errors"
)

// Send a multi-part message to a group
type Shout struct {
	address  string
	sequence uint16
	Group    string
	Content  []byte
}

// New creates new Shout message.
func NewShout() *Shout {
	shout := &Shout{}
	return shout
}

// String returns print friendly name.
func (s *Shout) String() string {
	return "SHOUT"
}

// Marshal serializes the message.
func (s *Shout) Marshal() ([]byte, error) {
	// Calculate size of serialized data
	bufferSize := 2 + 1 // Signature and message ID

	// Sequence is a 2-byte integer
	bufferSize += 2

	// Group is a string with 1-byte length
	bufferSize++ // Size is one byte
	bufferSize += len(s.Group)

	// Now serialize the message
	b := make([]byte, bufferSize)
	b = b[:0]
	buffer := bytes.NewBuffer(b)
	binary.Write(buffer, binary.BigEndian, Signature)
	binary.Write(buffer, binary.BigEndian, ShoutId)

	// Sequence
	binary.Write(buffer, binary.BigEndian, s.Sequence())

	// Group
	putString(buffer, s.Group)

	return buffer.Bytes(), nil
}

// Unmarshals the message.
func (s *Shout) Unmarshal(frames ...[]byte) error {
	frame := frames[0]
	frames = frames[1:]

	buffer := bytes.NewBuffer(frame)

	// Check the signature
	var signature uint16
	binary.Read(buffer, binary.BigEndian, &signature)
	if signature != Signature {
		return errors.New("invalid signature")
	}

	var id uint8
	binary.Read(buffer, binary.BigEndian, &id)
	if id != ShoutId {
		return errors.New("malformed Shout message")
	}

	// Sequence
	binary.Read(buffer, binary.BigEndian, &s.sequence)

	// Group
	s.Group = getString(buffer)

	// Content
	if 0 <= len(frames)-1 {
		s.Content = frames[0]
	}

	return nil
}

// Sends marshaled data through 0mq socket.
func (s *Shout) Send(socket *zmq.Socket) (err error) {
	frame, err := s.Marshal()
	if err != nil {
		return err
	}

	socType, err := socket.GetType()
	if err != nil {
		return err
	}

	// If we're sending to a ROUTER, we send the address first
	if socType == zmq.ROUTER {
		_, err = socket.Send(s.address, zmq.SNDMORE)
		if err != nil {
			return err
		}
	}

	// Now send the data frame
	_, err = socket.SendBytes(frame, zmq.SNDMORE)
	if err != nil {
		return err
	}
	// Now send any frame fields, in order
	_, err = socket.SendBytes(s.Content, 0)

	return err
}

// Address returns the address for this message, address should is set
// whenever talking to a ROUTER.
func (s *Shout) Address() string {
	return s.address
}

// SetAddress sets the address for this message, address should be set
// whenever talking to a ROUTER.
func (s *Shout) SetAddress(address string) {
	s.address = address
}

// SetSequence sets the sequence.
func (s *Shout) SetSequence(sequence uint16) {
	s.sequence = sequence
}

// Sequence returns the sequence.
func (s *Shout) Sequence() uint16 {
	return s.sequence
}