package internals

import "fmt"

var helpText = `
AI assitant for Cisco IOS-XE products.
	Arguments:
	copilot-chat										Chat with the AI assitant
	copilot <query>       	 							Queries adressed to AI Assistant
	copilot <show command> @ <query to AI assistant> 				AI Assistant helps with command's output
	copilot-optics <query>							        Queries AI Assistant regarding optics & compatibility
	copilot-feature <query>							         Queries AI Assistant regarding optics & compatibility 
	copilot-help  	     								Presents options to run AI assistant
	copilot-upgrade      	 							Upgrades the AI Assistant to the latest version
	copilot-cfg <API_KEY>  								Initial config of the script; Adds the API key;
	copilot-init									Initialization of AI assistant
	copilot-uninstall								Uninstall the AI assistant
	copilot-version                                                                 Shows installed version
	copilot-history										See what configs have been applied from copilot

For more information visit: https://docs.yosemite.iosxe.net/
	`

func (*Client) Help() {
	fmt.Println(helpText)
}
