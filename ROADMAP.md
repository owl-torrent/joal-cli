# ROADMAP to v3.0

- [ ] implement the new torrent struct
- [ ] link torrent with orchestrator
- [ ] link torrent with dispatcher
- [ ] link torrent with emulatedClient
- [x] change logging library from logrus to [uber-go/zap](https://github.com/uber-go/zap)
- [x] shuffle tier (or tracker i don't remember which) list when reading torrent file
- [ ] implement a replacement for seedmanager.seed-manager
- [ ] make listening port customizable
- [ ] Udp support
- [ ] allow proxy integration (via http client)
- [ ] run some real life tests on public trackers
- [ ] add a fake tracker for integration tests
- [ ] write some integrations tests
- [x] add multi tracker support (that mimic real clients)
- [x] add multi tier support (that mimic real clients)
- [ ] Add peer listener that "choke" everybody (emulatedclient.listener)

# Future improvements

- [ ] WebTorrent support (look at https://github.com/anacrolix/torrent/commit/77cbbec926f1bea68f2136917499b5e1acd3876f)
