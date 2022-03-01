# ROADMAP

## Beta availability
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
- [x] add an auto default configuration setup if config does not exists yet
- [x] wire-up the bandwidth dispatcher with the manager
- [x] wire-up the torrent stats updating on regular interval
- [ ] finish bandwidth dispatcher implementation
- [ ] implement a replacement for seedmanager.seed-manager
- [ ] publish messages from core to plugins
- [ ] review all the map[]: `delete` from map does not free any memory, if a map is getting a lot of delete it need to be rebuilt once in a while (iterate old with for and append values to a new one). https://stackoverflow.com/a/23231539/2275818
- [ ] review all the map[]: I've used torrent.InfoHash as the key at some place, it's fucked up since slice are ref, they are succeptible to change over time, use `torrent.InfoHash.AsHexString()` as the key instead
- [ ] run some real life tests on public trackers
- [ ] write some integrations tests
- [x] add multi tracker support (that mimic real clients)
- [x] add multi tier support (that mimic real clients)

## 1.0.0 GA
- [ ] Udp support
- [ ] make listening port customizable
- [ ] add some benchmarks
- [ ] add a fake tracker for integration tests

## V2.0.0
- [ ] Add peer listener that "choke" everybody (emulatedclient.listener)
- [ ] Replace current proxy usage with a SOCK5 proxy impl (work for both udp and TCP)

# Future improvements

- [ ] WebTorrent support (look at https://github.com/anacrolix/torrent/commit/77cbbec926f1bea68f2136917499b5e1acd3876f)
