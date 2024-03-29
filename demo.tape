# VHS documentation
#
# Output:
#   Output <path>.gif               Create a GIF output at the given <path>
#   Output <path>.mp4               Create an MP4 output at the given <path>
#   Output <path>.webm              Create a WebM output at the given <path>
#
# Settings:
#   Set FontSize <number>           Set the font size of the terminal
#   Set FontFamily <string>         Set the font family of the terminal
#   Set Height <number>             Set the height of the terminal
#   Set Width <number>              Set the width of the terminal
#   Set LetterSpacing <float>       Set the font letter spacing (tracking)
#   Set LineHeight <float>          Set the font line height
#   Set Theme <string>              Set the theme of the terminal (JSON)
#   Set Padding <number>            Set the padding of the terminal
#   Set Framerate <number>          Set the framerate of the recording
#   Set PlaybackSpeed <float>       Set the playback speed of the recording
#
# Sleep:
#   Sleep <time>                    Sleep for a set amount of <time> in seconds
#
# Type:
#   Type[@<time>] "<characters>"    Type <characters> into the terminal with a
#                                   <time> delay between each character
#
# Keys:
#   Backspace[@<time>] [number]     Press the Backspace key
#   Down[@<time>] [number]          Press the Down key
#   Enter[@<time>] [number]         Press the Enter key
#   Space[@<time>] [number]         Press the Space key
#   Tab[@<time>] [number]           Press the Tab key
#   Left[@<time>] [number]          Press the Left Arrow key
#   Right[@<time>] [number]         Press the Right Arrow key
#   Up[@<time>] [number]            Press the Up Arrow key
#   Down[@<time>] [number]          Press the Down Arrow key
#   Ctrl+<key>                      Press the Control key + <key> (e.g. Ctrl+C)
#
# Display:
#   Hide                            Hide the subsequent commands from the output
#   Show                            Show the subsequent commands in the output

Output demo.gif

Set FontSize 22
Set Width 1800
Set Height 1000
Set Theme { "name": "Eastwood", "black": "#424242", "red": "#ff7675", "green": "#badc58", "yellow": "#ffeaa7", "blue": "#74b9ff", "purple": "#cc99cc", "cyan": "#81ecec", "white": "#d3d0c8", "brightBlack": "#c4c2bb", "brightRed": "#ff7675", "brightGreen": "#badc58", "brightYellow": "#ffeaa7", "brightBlue": "#74b9ff", "brightPurple": "#cc99cc", "brightCyan": "#81ecec", "brightWhite": "#d3d0c8", "background": "#424242", "foreground": "#d3d0c8", "selectionBackground": "#74b9ff", "cursorColor": "#d3d0c8" }

Type "hclvet -v"

Enter

Sleep 500ms

Type "hclvet ruleset add github.com/clintjedwards/hclvet-ruleset-terraform-example"

Sleep 1.5s

Enter

Sleep 4s

Type "hclvet ruleset list terraform-example"

Sleep 1.5s

Enter

Sleep 6s

Type "hclvet lint ~/Documents/hclvet/internal/testdata/*"

Sleep 1.5s

Enter

Sleep 20s

