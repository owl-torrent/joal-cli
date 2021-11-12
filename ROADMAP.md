# ROADMAP to v3.0

- [x] remove dependency to mockgen. It introduce too much CI clutering
- [x] add github actions
- [x] implement the new torrent struct
- [x] link torrent with orchestrator
- [x] link torrent with dispatcher
- [x] link torrent with emulatedClient
- [x] test the new torrent structure
- [x] change logging library from logrus to [uber-go/zap](https://github.com/uber-go/zap)
- [x] shuffle tier (or tracker i don't remember which) list when reading torrent file
- [x] rewrite dispatcher with performance in mind. This is the bottleneck of the whole project.
- [x] add application configuration
- [ ] add an auto default configuration setup if config does not exists yet
- [ ] implement a replacement for seedmanager.seed-manager
- [ ] make listening port customizable
- [x] review all the map[]: `delete` from map does not free any memory, if a map is getting a lot of delete it need to be rebuilt once in a while (iterate old with for and append values to a new one). https://stackoverflow.com/a/23231539/2275818
- [ ] Udp support
- [ ] Replace current proxy usage with a SOCK5 proxy impl (work for both udp and TCP)
- [ ] run some real life tests on public trackers
- [ ] add a fake tracker for integration tests
- [ ] add some benchmarks
- [ ] write some integrations tests
- [x] add multi tracker support (that mimic real clients)
- [x] add multi tier support (that mimic real clients)
- [ ] Add peer listener that "choke" everybody (emulatedclient.listener)

# Future improvements

- [ ] WebTorrent support (look at https://github.com/anacrolix/torrent/commit/77cbbec926f1bea68f2136917499b5e1acd3876f)
