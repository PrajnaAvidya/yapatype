#!/bin/bash
rm -f /tmp/mykitty.sock

# launch new kitty window for claude
kitty -o allow_remote_control=yes --listen-on unix:/tmp/mykitty.sock 2>/dev/null &

# wait for socket
sleep 0.5

# run voice client in current terminal
/home/rafiq/Code/yapatype/yapatype client --name focus --kitty-socket /tmp/mykitty.sock
