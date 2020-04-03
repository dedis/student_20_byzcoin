/*
Package onet is the Overlay Network which offers a simple framework for generating
your own distributed systems. It is based on a description of your protocol
and offers sending and receiving messages, handling trees and host-lists, and
easy deploying to Localhost, Deterlab or a real-system.

ONet is based on the following pieces:

    - Local* - offers the user-interface to the API for deploying your protocol locally and for testing
    - Node / ProtocolInstance - gives an interface to define your protocol
    - Server - hold states for the different parts of Onet
    - network - uses secured connections between hosts

If you just want to use an existing protocol, usually the ONet-part is enough.
If you want to create your own protocol, you have to learn how to use the
ProtocolInstance.
*/
package onet

// Version history notes:
// 1.2 (no comment)
// 2.0 first version where no base64 is allowed in {public,private}.toml files. Cothority's
//     run_conode.sh migrates from 1.2->2.0 format files.
// 3+  version is recorded in the build via the Go modules system and exposed
//     via rsc.io/goversion/version (see server.go)
