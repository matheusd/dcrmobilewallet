# Decred Mobile Wallet

:iphone: :computer: :tada: :tada: :tada: :tada:

:exclamation: :exclamation: :exclamation: **This is currently a Proof of Concept. Use at your own risk!**

Please note that since dcrwallet does **not** support SPV mode yet, there is no privacy in the dcrd <-> dcrwallet message exchanges. In other words, the node the mobile wallet connects to is able to know **everything** about a given wallet (except its private keys). So only connect to a node you trust.

## Architecture

This is probably **not** how the final official decred mobile wallet will work. This is meant as a simple POC of how it *could* work, and also demonstrates a way that dcrwallet could be used as a library on a desktop client (instead of being used as a daemon as is currently the case).

General architecture is the following:

- dcrwallet is compiled as an Android library using gomobile;
- A new class (LibWallet) serves as interface between the go library and a mobile App;
- A native Android app instantiates the library and uses its methods to perform common wallet functions.

## Requirements

In order to run this POC, you need:

- An Android Phone/Tablet/VM
- A running and synced dcrd node, acessible by the phone

You need to make sure the dcrd node is reachable by the phone/vm. So, configure any firewalls, routing, NATs or whatever is needed to get the phone to ping/connect to the node. Easiest way to do that IME is to get a cheap VPS (eg, from AWS or linode), disable firewall and run a dcrd node with `rpclisten=`. Note that this POC is currently configured for **testnet only**.

Let the node sync to the current best block before connecting. Grab the rpcuser/rpcpassword/rpccert from the node. Try to connect from a desktop dcrwallet instance first, so you are sure you can remotely connect and authenticate to it.

After that, copy the file `user-consts.go-sample` (in the `pkg/mobilewallet` dir) to `user-consts.go` and edit it, filling the appropriate constants (mainly ip/port of the node and credentials). 

Then you can proceed to building and running the mobile wallet.


## Development Environment

- Golang, dep
- Android Sdk and Ndk
- Java Sdk
- Android Studio
- gomobile (correctly init'd with `gomobile init`)

If at any time you hit a gomobile compilation bug (due to missing strings.h), please check [comment by ivanalejandro0 on issue 22766 of golang/go github repo](https://github.com/golang/go/issues/22766#issuecomment-345137057).

Setting up the dev environment is finicky (the Android part anyway), so be ready to spend a couple of hours on this. At the very least, before starting a build you need to be able to do the following without any errors:

- `go version`
- `dep version`
- `gomobile version`
- `javac -version`
- `adb version`
- `adb devices` (show the device where you'll run the mobile app)
- `env | grep -E "(JAVA|GO|ANDROID)"` (you'll need `JAVA_HOME`, `GOPATH`, `GOROOT`, `ANDROID_HOME`)


## Build

### Android Library

Compile the Android lib with gomobile:

```
$ dep ensure
$ (cd pkg/mobilewallet && gomobile bind -target android -o ../../android/app/libs/mobilewallet.aar .)
```

### App with Android Studio

Import the `android` subdir as a project. Build and run on your phone/tablet/vm.

### App without Android Studio

*not yet available*

## Running

At least on my test phone, the operations are slow, so please be patient. Also, error handling is very primitive, so if you hit any snags you'll have to try and figure out how to fix it yourself.

- Create a wallet
- Open the wallet
- Generate a receive address
- Go to the testnet faucet and type it. Pay **very** close attention, as some letters may be a bit hard to distinguish
- Update the wallet to see the new balance (may need a few seconds/minutes to get the tx - no notifications yet. If the balance doesn't change, try sending to a different address, rescan, close and reopen app and wait some more)
- Send 1 DCR back to the configured address on the library

That's it!

# Future Work

As it is, this POC has fulfilled its purpose: prove that it is possible to build and run dcrwallet on a mobile platform. The only missing piece that I may still do is to start the wallet's grpc daemon and see if I can connect to it from the native code, which will simplify writing the official client.

Next steps to having a fully featured decred mobile wallet are (in no particular order):

- Waiting for SPV support (in the works)
- Refactoring dcrwallet wallet to be more library friendly (things like moving either config or wallet initialization into a self-contained package away from `package main`)
- Deciding on the actual tech stack of the mobile wallet (react native? Quasar? Phone gap? True native?)
- Doing all the work (that's the easy part :smirk:)
