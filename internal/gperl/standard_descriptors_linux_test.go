package gperl_test

import "syscall"

// dupOnto makes descriptor to a duplicate of from (dup2).
func dupOnto(from, to int) error { return syscall.Dup3(from, to, 0) }
