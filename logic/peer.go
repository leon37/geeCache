package logic

import pb "geeCache/logic/geecachepb"

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	Get(req *pb.Request, rsp *pb.Response) error
}
