Navigation: [DEDIS](https://github.com/dedis/doc/tree/master/README.md) ::
[Onet](../../README.md) ::
[Simulation](../README.md) ::
Mininet Simulation

# Mininet Simulation

Mininet allows to have a better control over the delay and bandwidth restrictions
during the simulation. While [Deterlab](DETERLAB.md) only has these restrictions
between servers, but not necessarily between all nodes, Mininet will set up the
same restriction between any two nodes that communcate with each other.

The Mininet simulation uses http://mininet.org/ to simulate a network where each
host represents a cothority node. It has been extensively tested on the
hardware of https://iccluster.epfl.ch, but should theoretically also run on
other kinds of hardware. As it needs to set up the mininet environment, it is
more fragile than the Deterlab simulations. But server availability in EPFL
is often better than on Deterlab.

Each T3-server on the iccluster is a 24-core server with 256GB of RAM and can
run around 300 cothority nodes simultaneously. So if you want to run a
simulation with 2000 nodes, you need at least 7 servers.

Unfortunately we still have the restriction that _all servers need to be on the
same subnet_! We know about the situation and hope to have a solution ready
sometime in the near future.

## Setting up ICCluster

Supposing you want to run your simulation using the iccluster-network, you
first need to reserve the machines. You have to do so one day in advance, as
it is not possible to reserve the machines on the same day. Once the machines
are ready, you need to install them.

1. In the ICCluster-admin interface, you need to go to `My Servers` -> `Setup`.
2. Chose the servers you reserved - take care, as there might be servers from
other people in the same lab!
3. `Choose a boot option` - please take _Ubuntu xenial 1604_ for best results
4. `Customization` - chose a password (it can be very simple) for the setup.
5. Optional: For easiest results, tick `Own %url-post-install-script%` and add
the following line to `My Servers` -> `Environment` -> `%url-post-install-script%`:
`https://raw.githubusercontent.com/dedis/onet/master/simul/platform/mininet/install_mininet.sh`
This will directly install mininet on the servers, else the simulation will have
to do so for you.
6. `Run Setup` and confirm the setup. THIS WILL DELETE ALL DATA ON THE SERVERS!

### Verifying everything is correctly set up

To ping server #33, you have to do the following:

```bash
ping iccluster033.iccluster.epfl.ch
```

If you are outside of EPFL, you need to use the VPN, as iccluster is only
reachable from inside EPFL.

You can also connect to the server with

```bash
ssh root@iccluster033.iccluster.epfl.ch
```

and enter your password that you gave in the `Customization` step above.

## Running a simulation

Now you are finally ready to run your simulation. We suppose that you have
your simulation running successfully under [Localhost](LOCALHOST.md). Then
all you need to do is:

```bash
go build && ./simul -platform mininet simul.toml
```

If it is the first time the simulation is run, it will ask you whether you want
to use iccluster or not. If you reply `Y` (or simply press `Enter`), you can
give the names of the servers you reserved and installed. An example is:
`31 32 33`. You don't need to enter leading `0`s, this will be automatically added,
as well as the `icculster.epfl.ch`.

During the first run, the simulation will make sure that mininet is correctly
installed on the servers and will try to install it if not.

### Debugging an installation

If something goes wrong, you can always try to run the
`$GOPATH/src/github.com/dedis/onet/simul/platform/mininet/setup_iccluster.sh`
bash file to have mininet correctly installed. A simple test is to ssh to the
remote machine and to run

```bash
mn -c
```

which should remove all running mininet sessions and return to the command line.

### Monitor port occupied

Sometimes the ssh-forwarding is misbehaving, and you cannot run your simulation.
The simplest solution is to restart the servers (not reinstall, just restart).
A faster option is to ssh to the first server and check if you can find the rogue
`sshd` process. Take care, there is a `sshd`-process that listens for incoming
connections - if you kill that one, you will have to restart the server, as you
won't have access to it...

### Network problems

An important restriction so far is the need of all servers to be on the same
subnet. To verify if this is the case, you can use the `host` or `dig`
command and verify that is the case.

### Other problems

You can always run the simulation with `-debug 3` to get more information and
eventually see what is going wrong:

```bash
go build && ./simul -platform mininet -debug 3 simul.toml
```

Then you will see more details of what is happening and eventually see what
you need to change.
