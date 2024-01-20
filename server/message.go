package network

type Message struct {
	Sender  *Peer
	Payload []byte
}

func NewMessage(p *Peer, payload []byte) *Message {
	return &Message{
		Sender:  p,
		Payload: payload,
	}
}
