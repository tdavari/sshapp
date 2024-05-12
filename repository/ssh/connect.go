package ssh

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

type Connection struct {
	Protocol           string   `mapstructure:"protocol"`
	Ip                 string   `mapstructure:"ip"`
	Port               string   `mapstructure:"port"`
	InitConfigCommands []string `mapstructure:"init_config_commands"`
	Proxy              string   `mapstructure:"proxy"`
}

type Credentials struct {
	Default struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"default"`
	Enable struct {
		Password string `mapstructure:"password"`
	} `mapstructure:"enable"`
}

type Device struct {
	Name        string `mapstructure:"name"`
	Connections struct {
		Cli Connection `mapstructure:"cli"`
	} `mapstructure:"connections"`
	Credentials Credentials `mapstructure:"credentials"`
	Os          string      `mapstructure:"os"`
	Type        string      `mapstructure:"type"`
	Client      *ssh.Client
}

type Devices struct {
	Devices []Device `mapstructure:"devices"`
}

func (d *Device) ConnectViaJump(j *ssh.Client) error {
	println("hi there")
	addr := d.Connections.Cli.Ip + ":" + d.Connections.Cli.Port
	conn, err := j.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	config := &ssh.ClientConfig{
		User: d.Credentials.Default.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(d.Credentials.Default.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Skip host key verification
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		log.Fatal(err)
	}

	sClient := ssh.NewClient(ncc, chans, reqs)
	d.Client = sClient

	return err
}

func (d *Device) Connect() error {
	addr := d.Connections.Cli.Ip + ":" + d.Connections.Cli.Port
	jumpClient, err := createSSHClient(addr, d.Credentials.Default.Username, d.Credentials.Default.Password)

	d.Client = jumpClient

	return err
}

func (d *Device) IsConnected() bool {
	// Check if the client is not nil
	if d.Client == nil {
		return false
	}

	// Attempt to create a new SSH session
	_, err := d.Client.NewSession()
	if err != nil {
		return false // If session creation fails, return false
	}

	// If no error occurs during session creation, the SSH connection is still open
	return true
}

// createSSHClient establishes an SSH connection to the given host using username and password authentication
func createSSHClient(host, user, password string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Skip host key verification
	}
	return ssh.Dial("tcp", host, config)
}

// executeCommand executes the given command on the SSH client and returns the output
func (d *Device) ExecuteCommand(command string) (string, error) {
	session, err := d.Client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func Load(path string) ([]Device, error) {

	var devices Devices
	// Using Viper to read the YAML configuration file
	viper.SetConfigName(path) // Name of your YAML file (without the extension)
	viper.AddConfigPath(".")  // Look for the file in the current directory
	err := viper.ReadInConfig()
	if err != nil {
		return devices.Devices, fmt.Errorf("fatal error config file: %s", err)
	}

	// Unmarshal the YAML data into a struct
	err = viper.Unmarshal(&devices)
	if err != nil {
		return devices.Devices, fmt.Errorf("failed to unmarshal devices: %s", err)
	}

	return devices.Devices, nil
}
