# Records the SSH Tamagotchi TUI.
# Requires the server to be running on localhost:2222, e.g. `make run`.
#
#   vhs demo.tape
#
Output demo.gif

Require ssh

Set Shell "bash"
Set FontFamily "JetBrains Mono"
Set FontSize 18
Set Width 1200
Set Height 720
Set Padding 24
Set TypingSpeed 50ms
Set Theme "Catppuccin Mocha"

Hide
Type "mkdir -p .ssh; test -f .ssh/demo_ed25519 || ssh-keygen -q -t ed25519 -N '' -f .ssh/demo_ed25519"
Enter
Type "clear"
Enter
Show

Type "ssh -F demo_ssh_config pet"
Enter
Sleep 3s

Type "f"
Sleep 1s

Type "p"
Sleep 1s

Type "s"
Sleep 1s

Type "s"
Sleep 1s

Type "q"
Sleep 1s
