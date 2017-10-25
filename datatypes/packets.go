package datatypes

type PacketId int

func (packetId PacketId) Abs() int {
	return int(packetId)
}
