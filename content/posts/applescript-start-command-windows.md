---
# WHY: Preserve this archived article's metadata and exact historical URL.
{
  "title": "Applescript script to emulate the behaviour of the `start' command in the windows CMD prompt",
  "date": "2011-07-01T00:00:00+02:00",
  "slug": "applescript-start-command-windows",
  "draft": false,
  "tags": [
    "os-x",
    "applescript"
  ],
  "url": "/2011/07/01/applescript-script-to-emulate-the-behaviour-of-the-start-command-in-the-windows-cmd-prompt/",
  "aliases": [
    "/posts/applescript-start-command-windows/",
    "/2011/07/01/applescript-start-command-windows/"
  ]
}
---

People have asked me how to launch a program through the terminal in the
background but with it's own terminal screen similar to the way the "start"
command in the Windows CMD prompt works.
Well here it is.
Add the following two
files (start.applescript) and start to your path, and make them executable
(`chmod +x {start.applescript,start}`).

One point of notice is to <strong>always</strong> encapsulate strings with
spaces that should be one parameter in quotes.

`gistfile1.bash`

```sh
# /usr/bin/env /bin/sh
start.applescript "`pwd`" $@
```

`start.applescript`

```applescript
#! /usr/bin/env /usr/bin/osascript
on run argv
  
  tell application "Terminal"
    set AppleScript's text item delimiters to " "

    set cwd to first item of argv
    set l to rest of argv

    do script "cd " & cwd & " && " & (l as text)
    activate
  end tell

end run
```
