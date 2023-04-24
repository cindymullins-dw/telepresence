package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/telepresenceio/telepresence/v2/pkg/vif"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	vif.InitLogger(ctx)
	dev, err := vif.NewTunnelingDevice(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer dev.Close(context.Background())
	go func() {
		err := dev.Run(ctx)
		if err != nil {
			panic(err)
		}
	}()
	yesRoutes := []*net.IPNet{}
	noRoutes := []*net.IPNet{}
	whitelist := []*net.IPNet{}
	for _, cidr := range os.Args[1:] {
		var ipnet *net.IPNet
		var err error
		if strings.HasPrefix(cidr, "!") {
			_, ipnet, err = net.ParseCIDR(strings.TrimPrefix(cidr, "!"))
			fmt.Printf("Blacklisting route: %s\n", ipnet)
			noRoutes = append(noRoutes, ipnet)
		} else if strings.HasPrefix(cidr, "+") {
			_, ipnet, err = net.ParseCIDR(strings.TrimPrefix(cidr, "+"))
			fmt.Printf("Whitelisting route: %s\n", ipnet)
			whitelist = append(whitelist, ipnet)
		} else {
			_, ipnet, err = net.ParseCIDR(cidr)
			fmt.Printf("Adding route: %s\n", ipnet)
			yesRoutes = append(yesRoutes, ipnet)
		}
		if err != nil {
			panic(err)
		}
	}
	dev.Router.UpdateWhitelist(whitelist)
	err = dev.Router.UpdateRoutes(ctx, yesRoutes, noRoutes)
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.TrimSpace(line) == "" {
					cancel()
					return
				}
			}
			if err := scanner.Err(); err != nil {
				panic(err)
			}
		}
	}()
	defer fmt.Println("Okay bye!")
	fmt.Printf("Device: %s\n", dev.Device.Name())
	fmt.Println("Press enter (empty line) to end")
	<-ctx.Done()
}
