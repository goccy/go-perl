// Package cpan is the link set of a project's CPAN dependencies: the
// place where their module trees and native XS modules enter a binary.
// The go-perl package imports it, so every instance sees whatever it
// carries.
//
// This copy carries nothing. gperl generates a module with the same
// import path from a project's cpanfile.snapshot and puts it in a Go
// workspace (go.work), where it takes this package's place.
package cpan
