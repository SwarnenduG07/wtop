//go:build !unix

package main

func setupResizeCh() <-chan struct{} {
	return make(chan struct{})
}
