package docker

import "encoding/json"

type DockerServer struct {
	IDAtr         string
	DomainAtr     string
	AddressAtr    string
	WeightAtr     int
	ContainerName string
}

func (d DockerServer) ID() string {
	return d.IDAtr
}

func (d DockerServer) Domain() string {
	return d.DomainAtr
}

func (d DockerServer) Address() string {
	return d.AddressAtr
}

func (d DockerServer) Weight() int {
	return d.WeightAtr
}

func (d DockerServer) String() string {
	bs, _ := json.Marshal(d)
	return string(bs)
}

type label struct {
	Domain   string
	Port     int
	Weight   int
	Internal bool //使用内部地址
}

type labels map[int]label

func (ls *labels) Has() bool {
	return len(*ls) > 0
}
