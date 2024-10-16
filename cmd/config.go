package cmd

import (
	"errors"
	"fmt"
	"net"
	"qsend/config"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use: "config",
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := config.DefaultPath(appName)
		if err == nil {
			fmt.Println(p)
		}
		return err
	},
}

func Select[T any](label string, items []T) (int, error) {
	prompt := promptui.Select{
		Label: label,
		Items: items,
	}
	i, _, err := prompt.Run()
	if err != nil {
		return 0, err
	}
	return i, nil
}

func Prompt(label string, validator promptui.ValidateFunc, defaultValue ...string) (string, error) {
	prompt := promptui.Prompt{
		Label:    label,
		Validate: validator,
	}
	if len(defaultValue) > 0 {
		prompt.Default = defaultValue[0]
	}
	return prompt.Run()
}

func selectBind() (string, error) {
	ifaces, err := Interfaces()
	if err != nil {
		return "", err
	}
	i, err := Select("Select network interface", append([]Interface{
		{Name: "all"},
		{Name: "localhost"},
	}, ifaces...))
	if err != nil {
		return "", err
	}
	switch i {
	case 0: // all
		return "", nil
	case 1: // localhost
		return "localhost", nil
	default:
	}

	iface := ifaces[i-2]
	if len(iface.Addrs) == 1 {
		return iface.Addrs[0].String(), nil
	}
	i, err = Select("Select IP address", iface.Addrs)
	if err != nil {
		return "", err
	}

	return iface.Addrs[i].String(), nil
}

type Interface struct {
	Name  string
	Addrs []net.IP
}

func (i Interface) String() string {
	return i.Name
}

var errInvalidInterface = errors.New("invalid interface")
var errNoLAN = errors.New("the interface has no local network addresses")

func newInterface(i net.Interface) (Interface, error) {
	if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
		return Interface{}, errInvalidInterface
	}
	addrs, err := i.Addrs()
	if err != nil {
		return Interface{}, errInvalidInterface
	}
	iface := Interface{
		Name:  i.Name,
		Addrs: make([]net.IP, 0, len(addrs)),
	}
	for _, addr := range addrs {
		network, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		ip := network.IP
		if ip.IsLinkLocalUnicast() ||
			!ip.IsPrivate() || ip.IsLoopback() {
			continue
		}
		iface.Addrs = append(iface.Addrs, ip)
	}
	if len(iface.Addrs) == 0 {
		return Interface{}, errNoLAN
	}
	return iface, nil
}

// Interfaces returns the system network interfaces except loopback and down
// with the ip addresses that are can be accessed from local network.
func Interfaces() ([]Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	list := make([]Interface, 0, len(ifaces))
	for _, i := range ifaces {
		iface, err := newInterface(i)
		if err != nil {
			continue
		}
		list = append(list, iface)
	}
	return list, nil
}
