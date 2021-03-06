package main

import (
	"flag"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"goim-demo/internal/comet"
	"goim-demo/internal/comet/conf"
	"goim-demo/internal/comet/grpc"
	"goim-demo/pkg/etcdv3"

	log "github.com/golang/glog" //日志默认放在/tmp 目录
)

const (
	ver = "2.0.0"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())

	log.Infof("goim-comet [version: %s conf: %+v] start", ver, conf.Conf)
	// discovery
	dis := etcdv3.New(conf.Conf.Discovery.Nodes)
	dis.ResolverEtcd()

	// new comet server
	srv := comet.NewServer(conf.Conf)
	if err := comet.InitWhitelist(conf.Conf.Whitelist); err != nil {
		panic(err)
	}

	if err := comet.InitTCP(srv, conf.Conf.TCP.Bind, runtime.NumCPU()); err != nil {
		panic(err)
	}

	if err := comet.InitWebsocket(srv, conf.Conf.Websocket.Bind, runtime.NumCPU()); err != nil {
		panic(err)
	}
	/*
		if conf.Conf.Websocket.TLSOpen {
			if err := comet.InitWebsocketWithTLS(srv, conf.Conf.Websocket.TLSBind, conf.Conf.Websocket.CertFile, conf.Conf.Websocket.PrivateFile, runtime.NumCPU()); err != nil {
				panic(err)
			}
		}
	*/

	rpcSrv := grpc.New(conf.Conf.RPCServer, srv)
	register(dis, conf.Conf)

	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("goim-comet get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			rpcSrv.GracefulStop()
			srv.Close()
			dis.Deregister() // 移除 etcd 中的节点
			log.Infof("goim-comet [version: %s] exit", ver)
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

// 服务注册
func register(dis *etcdv3.Registry, c *conf.Config) error {
	// 当前grpc 服务的 外网ip 端口
	_, port, _ := net.SplitHostPort(c.RPCServer.Addr)
	ip := c.Env.Host
	region := c.Env.Region
	zone := c.Env.Zone
	env := c.Env.DeployEnv
	appid := "goim.comet"
	// 服务注册至ETCD
	return dis.Register(5, env, appid, region, zone, ip, port)
}
