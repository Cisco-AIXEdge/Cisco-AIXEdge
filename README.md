# Cisco-AIXEdge

BUILD COMMAND GUESTSHELL

env GOOS=linux GOARCH=386 go build ./copilot.go


![ChatGPT Image Apr 10, 2025 at 02_42_48 PM](https://github.com/user-attachments/assets/0e3440bc-28d5-4f00-9c72-f1171627d67a)

Release Notes:

0.0.5   

        added CFN integration
        
        added Service Register (string hashing for all the features enabled on the device)

        Multi-LLM options for all existing functionalities
        

0.0.4   

        added EULA
        
        optics compatibility feature added in aws lambda
        
        change the way we catch SN/PID

0.0.3   

        added error handling with telemetry (we see what commands don't work in aws)
        
        created always on sandbox - https://demo.yosemite.iosxe.net/

0.0.2 

        added copilot-version and copilot-upgrade as new commands
        
        works in STATIC/DHCP environment
        
        works behind a proxy

0.0.1 
        
        interprets output
        
        answers general questions
        
        working only with direct internet connection
        
