# Journal TUI

A simple terminal UI for managing my journal.

## Intro

My journal is an encrypted [gocryptfs] directory that contains Markdown files.
The app handles mounting the decrypted journal, and provides a UI for browsing,
reading, searching, and creating/editing entries. Apart from convenient (and
cool!), this is also more secure than mounting the journal manually and editing
the files directly. This is because the app runs `gocryptfs` with `-fg`, so when
the app is closed, the journal is unmounted. This eliminates the risk of 
forgetting to unmount the journal when I'm done, as I often do!

## Requirements

- [gocryptfs]
- [tmux]
- [neovim]

## How to use

Install with Go:

```
go install https://github.com/mecha/journal@latest
```

Run with the path to the encrypted directory as argument:

```
journal /path/to/encrypted/dir
```

Follow the [gocryptfs] instructions for how to set up an encrypted directory.

When the app opens, simply enter the password to decrypt the directory. You'll
figure it out from there. Or maybe you won't. But I believe in you.

## TODO

- [x] Replace polling with file watcher
- [ ] Use `$EDITOR` env var instead of assuming Neovim
- [ ] Add flags to make `tmux` dependency optional
- [ ] Add color override support through env vars

## Credits :point_down:

The TUI design is inspired by [lazygit] and built using the [tcell] package.

[gocryptfs]: https://nuetzlich.net/gocryptfs
[tmux]: https://github.com/tmux/tmux
[neovim]: https://github.com/neovim/neovim
[lazygit]: https://github.com/jesseduffield/lazygit
[tcell]: https://github.com/gdamore/tcell
