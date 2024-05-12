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
}

type Devices struct {
	Devices []Device `mapstructure:"devices"`
}

func (d Device) ConnectViaJump(j *ssh.Client) (*ssh.Client, error) {
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
	return sClient, err
}

func (d Device) Connect() (*ssh.Client, error) {
	addr := d.Connections.Cli.Ip + ":" + d.Connections.Cli.Port
	jumpClient, err := createSSHClient(addr, d.Credentials.Default.Username, d.Credentials.Default.Password)
	return jumpClient, err
}

// func ConnectViaJump(d []Device, j Device) {
// 	// Create an SSH client to the jump box
// 	jaddr := j.Connections.Cli.Ip + ":" + j.Connections.Cli.Port
// 	jumpClient, err := createSSHClient(jaddr, j.Credentials.Default.Username, j.Credentials.Default.Password)
// 	if err != nil {
// 		log.Fatalf("Failed to connect to jump box: %v", err)
// 	}
// 	defer jumpClient.Close()

// 	// Loop through each switch and execute show run command concurrently
// 	for _, host := range d {

// 		haddr := host.Connections.Cli.Ip + ":" + host.Connections.Cli.Port

// 		conn, err := jumpClient.Dial("tcp", haddr)
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		config := &ssh.ClientConfig{
// 			User: host.Credentials.Default.Username,
// 			Auth: []ssh.AuthMethod{
// 				ssh.Password(host.Credentials.Default.Password),
// 			},
// 			HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Skip host key verification
// 		}

// 		ncc, chans, reqs, err := ssh.NewClientConn(conn, haddr, config)
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		sClient := ssh.NewClient(ncc, chans, reqs)

// 		// _, err = executeCommand(sClient, "terminal width 0")
// 		// if err != nil {
// 		// 	log.Printf("Failed to execute terminal width 0 on switch %s: %v", host, err)
// 		// 	return
// 		// }
// 		_, err = executeCommand(sClient, "terminal length 0")
// 		if err != nil {
// 			log.Printf("Failed to execute terminal length 0 on switch %s: %v", host, err)
// 			return
// 		}
// 		// Execute the show run command on the switch
// 		output, err := executeCommand(sClient, "show run")
// 		if err != nil {
// 			log.Printf("Failed to execute command on switch %s: %v", host, err)
// 			return
// 		}

// 		// Print the output of the command
// 		fmt.Printf("Output from switch %s:\n%s\n", host, output)

// 	}

// }

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
func ExecuteCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
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
