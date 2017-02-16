<img src="http://2.bp.blogspot.com/-cDi1mp5gxI8/Vhppy38tWfI/AAAAAAAAY4g/XvWB3QG_a-s/s1600/Filler02%2BDrink%2BEat%2BMe_FINALsml.jpg" width="256">
# Alice

Alice is an autoscaler used at [Notonthehighstreet](http://www.notonthehighstreet.com). It is designed to scale up and down
server and application instances based on daily demand, and is designed to be flexible and easy to extend.

Alice is written in Go and runs as a standalone background process in our production environment. Using metrics from a
monitoring provider, it can make decisions on how to scale an inventory of resources. An inventory can be any set of
resources that provide the capacity to deal with current demand - for example, the servers in a server farm, or the
number of docker containers running on an instance of your application.

To support different monitoring providers and resources types, we have to write a monitor or inventory plugin to talk to
the different backend APIs.

The currently supported backends are:

 - **Monitors**: Datadog, Stats directly from Mesos
 - **Inventories**: Marathon applications, AWS EC2 instances (via autoscaling groups)
 
It's relatively easy to write plugins for additional backends as required.

In addition to this, Alice is also easy to integrate with Slack and Fluentd.

## Build

We use [glide](https://github.com/Masterminds/glide) for vendoring dependencies, so once you have cloned this repository
you will want to install it. This is usually fairly straightfoward:

OSX: `brew install glide`

Ubuntu: `sudo add-apt-repository ppa:masterminds/glide && sudo apt-get update &&sudo apt-get install glide`

Once this is installed you can install all the dependencies (into `./vendor` by default):

`glide install`

To build Alice, just `go build`.

## Configuration

When Alice starts up it reads `./config/config.yaml` which should be relative to the executable's working directory.
Copying and editing the `config/config.yaml.dist` file is a good place to start.

The main body of the configuration is under the `managers` section of the config file. Alice can manage multiple
resource inventories at a time. A manager is a grouping of an inventory to be managed, a monitor from which to collect
metrics, and a strategy which determines what scaling should be done on the inventory taking the current metrics into
account.

A simple manager configuration might look like this:

```
managers:
  # Unique name of this manager
  my_web_application:
    monitor:
      # A datadog plugin example
      name: datadog
      api_key: xxxxxx
      app_key: xxxxxx
      time_period: 5m  # Most recent data point must be within this time
      metrics:
        active_users:
          query: avg:application.users.active{*}

    inventory:
      # A marathon application plugin example
      name: marathon
      settle_down_period: 3m
      url: http://marathon.example.com:8080
      app: www_example_com  # Application ID in marathon
      minimum_instances: 1
      maximum_instances: 10

    strategy:
       # A ratio strategy plugin example
      name: ratio
      ratios:
        active_users:
          metric: 100  # Keep one application instance for every 100 active users
          inventory: 1
```

## How to test the software

The tests for Alice can be run using `go test` like this: `go test -race -cover $(go list ./... | grep -v /vendor/)`

Alternatively you can run `make test`

## Getting help

If you have questions, concerns, bug reports, etc, please file an issue in this repository's Issue Tracker.

## Getting involved

If you'd like to get involved and make Alice better, we'd be happy to accept your help!

Please read [CONTRIBUTING](CONTRIBUTING.md) for more information on how to contribute to Alice.


