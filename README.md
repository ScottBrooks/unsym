# unsym
Tries to turn a stacktrace from eu-stack into something real with unreals .sym files

```shell
$ apt-get update && apt-get install elfutils # Install eu-stack
$ ps ax | grep Server # usually 31 or 32 or something.
$ cat /proc/<PID_OF_SERVER>/maps | head -n 1 # the start of the output is the load address of this process.  Something like 00200000-02820000.  We want the 00200000, which we pass below as 0x200000
$ eu-stack -p <PID_OF_SERVER> -r | unsym <PATH_TO_SERVER.sym> 0x200000 # all threads
$ eu-stack -p <PID_OF_SERVER> -r -1 | unsym <PATH_TO_SERVER.sym> 0x200000 # just a single thread
```
