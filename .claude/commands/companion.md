---
description: Switch the xscapes companion (cat, crab)
---

Run `xscapes companion $ARGUMENTS` and report what it printed.

If no argument was given, the command prints the current companion and the
available ones — relay that. If an argument was given, it is saved to
`~/.config/xscapes/companion` and a running scape picks it up on its next
frame, so tell the user it has changed and that they do not need to restart.

If it exits non-zero, the name was not one of the available companions; say
which ones are.
