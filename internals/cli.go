package internals

import "fmt"

var helpText = `
AI assitant for AIXEdge products.
	Arguments:
	aixedge-chat										Chat with the AI assitant
	aixedge <query>       	 							Queries adressed to AI Assistant
	aixedge <show command> @ <query to AI assistant> 				AI Assistant helps with command's output
	aixedge-pacp <file_name> <query to AI assistant>		AI assistant helps with PCAP interpretation
	aixedge-optics <query>							        Queries AI Assistant regarding optics & compatibility
	aixedge-feature <query>							         Queries AI Assistant regarding optics & compatibility 
	aixedge-help  	     								Presents options to run AI assistant
	aixedge-upgrade      	 							Upgrades the AI Assistant to the latest version
	aixedge-cfg <API_KEY>  								Initial config of the script; Adds the API key;
	aixedge-init									Initialization of AI assistant
	aixedge-uninstall								Uninstall the AI assistant
	aixedge-version                                                                 Shows installed version
	aixedge-history										See what configs have been applied from copilot

For more information visit: https://github.com/Cisco-AIXEdge/Cisco-AIXEdge
	`

func (*Client) Help() {
	fmt.Println(helpText)
}
