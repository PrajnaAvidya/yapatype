#!/bin/bash
SOCK="$HOME/.local/share/yapatype/kitty.sock"
mkdir -p "$(dirname "$SOCK")"
rm -f "$SOCK"

# launch new kitty window for claude
kitty -o allow_remote_control=yes --listen-on "unix:$SOCK" 2>/dev/null &

# wait for socket
sleep 0.5

# run voice client in current terminal
"$HOME/Code/yapatype/yapatype" client --name focus --kitty-socket "$SOCK"
