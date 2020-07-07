# rcon
A "run context" utility that can be used to build mock environments for testing software in. Driven
by a configuration file that describes the command to run and defines some attributes about the
context such as environmental variables, timers, network throttling, I/O throttling, file
replacements, and input/output controls.

Requires Linux with a mounted /proc filesystem and cgroup and namespace support
