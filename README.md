# Telegraf Input Plugin for Lotus

A [Telegraf](https://github.com/influxdata/telegraf) external (`execd`) input plugin for streaming metrics from Filecoin `lotus` and `lotus-miner` nodes.

## Building from source
Requirements:
- **Golang**: tested on version `go1.17.1 darwin/amd64` and `go1.17.1 linux/amd64`

Steps:
- Run `make build`
- The binary should be available inside `bin/telegraf-input-lotus`

## Running the plugin
Requirements:
- Available `lotus` and `lotus-miner` nodes. See the [Filecoin docs](https://docs.filecoin.io/).

Example plugin config (`plugin.conf`):
```toml
daemonAddr = "127.0.0.1:1234"
daemonToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.ioNehvGKVQ6_aqxq77X6WK5dsESkbKEE5QW_NTuczME"
minerAddr = "127.0.0.1:2345"
minerToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.3VvvGysY9ZwUG6e3Bap0g2TVL-2tcjAJAb5Aw2kPcsk"
```
If you want only metrics from daemon, set the miner address and token to `""` and vice-versa.


Example Telegraf config section (`/etc/telegraf/telegraf.conf`):
```toml
... other config ...

[[inputs.execd]]
  command = ["./bin/telegraf-input-lotus", "-config", "plugin.conf"]

... other config ...
```

## Contributing
There is a Vagrant/Ansible development environment setup inside the `env` folder. It provisions a local ubuntu VM with the whole TICK stack and the necessary lotus binaries.

Requirements:
- [Ansible](https://www.ansible.com/): tested with version `ansible [core 2.11.5]`
- [Vagrant](https://www.vagrantup.com/): tested with version `Vagrant 2.2.18`

Running the development environment:
- `cd env`
- `vagrant up`
- `vagrant ssh`

Building & running the code (inside the virtual machine):
- `make rundev`

By default `make rundev` streams the plugin output metrics to `stdout`.

To check if it's actually streaming metrics to influx you can take a look at the following:
- If you built a new version of the code you need to restart telegraf so it can start the new binary `sudo service telegraf restart`
- To check if there are any obvious errors with the plugin, run `sudo service telegraf status`
- Check if lotus measurements are persisted inside InfluxDB:
```bash
$ influx
> use telegraf
> select * from lotus
```

The `select` query should return some results.