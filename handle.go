package main

/*
func handleDecodedMsg(m *msg.Msg) (*msg.Msg, error) {
	switch m.Type {
	case "ACK":
		event.Handle(m.InResponseTo)

	case "PING":
		return &msg.Msg{
			Type: "ACK",
		}, nil

	case "SET":
		kv := &store.KV{}
		err := cbor.Unmarshal(m.Body, kv)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal SET msg: %w", err)
		}
		fmt.Println("set: ", kv)
		store.Set(kv.Key, kv.Value)
		return &msg.Msg{
			Type: "ACK",
			Body: []byte("set"),
		}, nil

	case "GET":
		v := store.Get(string(m.Body))
		return &msg.Msg{
			Type: "VALUE",
			Body: v,
		}, nil
	}
	return &msg.Msg{
		Type: "ERROR",
		Body: []byte("unknown message type"),
	}, nil
}
*/
