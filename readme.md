A disk usage utility for Mac
====

```
============================== Mac Disk Usage Utility ==============================
Enter folder # to explore, ".." to go back, "." to refresh, Ctrl+C to quit 
------------------------------------------------------------------------------------
/var
------------------------------------------------------------------------------------
    1 + [log]                                                                80.84 MB
    2 + [protected]                                                          11.38 MB
    3 + [logs]                                                                3.97 MB
    4 + [MobileSoftwareUpdate]                                              390.93 KB
    5 + [db]                                                                 60.97 KB
    6 + [mobile]                                                              7.64 KB
    7 + [run]                                                                   47 B
    8 + [sntpd]                                                                 24 B
    9 + [select]                                                                 9 B
   10 + [at]                                                                     6 B
   11 + [msgs]                                                                   4 B
------------------------------------------------------------------------------------
> COMMAND
```

#### Usage

`% ./disk-usage-mac <directory>`

`% ./disk-usage-mac ~`

#### Arguments

- start directory (optional)

#### Commands

- `<number>` - enter a number of a child folder to explore

- `.` - reload

- `..` - parent dir

- `open`, `finder` - open current dir in Finder

- `reveal` - reveal current dir in Finder

- `q`, `\q`, `quit`, `exit`, Ctrl+C, SIGINT, SIGTERM - quit

#### License

MIT License
