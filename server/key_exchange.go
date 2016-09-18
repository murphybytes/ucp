package main

type keyManager interface {
	doPublicKeyExchange() error
}
