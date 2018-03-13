## WIP

## Goal
Automagic bash script dependency management for [https://github.com/RyanJarv/coderun](https://github.com/RyanJarv/coderun)

## Example
At the moment this just parses a given bash script and queries command-not-found db for missing dependencies

```
$ go run -v main.go ./test.sh 
command-line-arguments
✓ echo
✓ command
✓ /bin/true
✓ /bin/false
✓ cat
✓ ls
✓ cat
✓ cat
✗ iptables
```
