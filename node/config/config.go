package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pactus-project/pactus/consensus"
	"github.com/pactus-project/pactus/network"
	"github.com/pactus-project/pactus/state"
	"github.com/pactus-project/pactus/store"
	"github.com/pactus-project/pactus/sync"
	"github.com/pactus-project/pactus/txpool"
	"github.com/pactus-project/pactus/util"
	"github.com/pactus-project/pactus/util/logger"
	"github.com/pactus-project/pactus/www/capnp"
	"github.com/pactus-project/pactus/www/grpc"
	"github.com/pactus-project/pactus/www/http"
	"github.com/pactus-project/pactus/www/nanomsg"
	toml "github.com/pelletier/go-toml"
)

//go:embed example_config.toml
var exampleConfigBytes []byte

type Config struct {
	State     *state.Config     `toml:"state"`
	Store     *store.Config     `toml:"store"`
	Network   *network.Config   `toml:"network"`
	Sync      *sync.Config      `toml:"sync"`
	TxPool    *txpool.Config    `toml:"tx_pool"`
	Consensus *consensus.Config `toml:"consensus"`
	Logger    *logger.Config    `toml:"logger"`
	GRPC      *grpc.Config      `toml:"grpc"`
	Capnp     *capnp.Config     `toml:"capnp"`
	HTTP      *http.Config      `toml:"http"`
	Nanomsg   *nanomsg.Config   `toml:"nanomsg"`
}

func DefaultConfig() *Config {
	conf := &Config{
		State:     state.DefaultConfig(),
		Store:     store.DefaultConfig(),
		Network:   network.DefaultConfig(),
		Sync:      sync.DefaultConfig(),
		TxPool:    txpool.DefaultConfig(),
		Consensus: consensus.DefaultConfig(),
		Logger:    logger.DefaultConfig(),
		GRPC:      grpc.DefaultConfig(),
		Capnp:     capnp.DefaultConfig(),
		HTTP:      http.DefaultConfig(),
		Nanomsg:   nanomsg.DefaultConfig(),
	}

	return conf
}

func SaveMainnetConfig(path, rewardAddr string) error {
	exampleConfig := string(exampleConfigBytes)

	exampleConfig = strings.Replace(exampleConfig, "## reward_address = \"\"",
		fmt.Sprintf("  reward_address = \"%s\"", rewardAddr), 1)

	return util.WriteFile(path, []byte(exampleConfig))
}

func SaveTestnetConfig(path, rewardAddr string) error {
	conf := DefaultConfig()
	conf.Network.Name = "pactus-testnet"
	conf.Network.Listens = []string{"/ip4/0.0.0.0/tcp/21777", "/ip6/::/tcp/21777"}
	conf.Network.Bootstrap.Addresses = []string{"/ip4/172.104.46.145/tcp/21777/p2p/12D3KooWNYD4bB82YZRXv6oNyYPwc5ozabx2epv75ATV3D8VD3Mq"}
	conf.Network.Bootstrap.MinThreshold = 4
	conf.Network.Bootstrap.MaxThreshold = 8
	conf.GRPC.Enable = true
	conf.GRPC.Listen = "[::]:9090"
	conf.GRPC.Gateway.Enable = true
	conf.GRPC.Gateway.Listen = "[::]:80"
	conf.Capnp.Enable = true
	conf.Capnp.Listen = "[::]:37621"
	conf.HTTP.Enable = true
	conf.HTTP.Listen = "[::]:8080"
	conf.State.RewardAddress = rewardAddr
	conf.Nanomsg.Enable = true
	conf.Nanomsg.Listen = "tcp://127.0.0.1:40899"

	return util.WriteFile(path, conf.toTOML())
}

func SaveLocalnetConfig(path, rewardAddr string) error {
	conf := DefaultConfig()
	conf.Network.Name = "pactus-localnet"
	conf.Network.Listens = []string{}
	conf.Network.Bootstrap.Addresses = []string{}
	conf.Network.Bootstrap.MinThreshold = 4
	conf.Network.Bootstrap.MaxThreshold = 8
	conf.GRPC.Enable = true
	conf.GRPC.Listen = "[::]:9090"
	conf.GRPC.Gateway.Enable = true
	conf.GRPC.Gateway.Listen = "[::]:8080"
	conf.Capnp.Enable = true
	conf.Capnp.Listen = "[::]:37621"
	conf.HTTP.Enable = true
	conf.HTTP.Listen = "[::]:8081"
	conf.State.RewardAddress = rewardAddr
	conf.Nanomsg.Enable = true
	conf.Nanomsg.Listen = "tcp://127.0.0.1:40899"

	return util.WriteFile(path, conf.toTOML())
}

func (conf *Config) toTOML() []byte {
	buf := new(bytes.Buffer)
	encoder := toml.NewEncoder(buf)
	encoder.Order(toml.OrderPreserve)
	err := encoder.Encode(conf)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func LoadFromFile(file string) (*Config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	conf := DefaultConfig()
	buf := bytes.NewBuffer(data)
	decoder := toml.NewDecoder(buf)
	decoder.Strict(true)
	if err := decoder.Decode(conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func (conf *Config) SanityCheck() error {
	if err := conf.State.SanityCheck(); err != nil {
		return err
	}
	if err := conf.Store.SanityCheck(); err != nil {
		return err
	}
	if err := conf.TxPool.SanityCheck(); err != nil {
		return err
	}
	if err := conf.Consensus.SanityCheck(); err != nil {
		return err
	}
	if err := conf.Network.SanityCheck(); err != nil {
		return err
	}
	if err := conf.Logger.SanityCheck(); err != nil {
		return err
	}
	if err := conf.Sync.SanityCheck(); err != nil {
		return err
	}
	if err := conf.Capnp.SanityCheck(); err != nil {
		return err
	}
	if err := conf.Nanomsg.SanityCheck(); err != nil {
		return err
	}
	return conf.HTTP.SanityCheck()
}
