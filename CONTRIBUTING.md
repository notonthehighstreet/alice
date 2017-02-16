# Contributing

1. [Running Alice locally](#running-alice-locally)
2. [Testing](#testing)
3. [Submitting code](#submitting-code)

## Running Alice locally

Alice comes with a 'fake' inventory and monitor that you can use when running locally.

The fake inventory does nothing but keep a count of 'things' and allow itself to be scaled up and down.
The fake monitor will generate metrics for you in a sine wave pattern so you can watch scaling decisions being made
as the metric increases and drops back down again.

Here is a basic working config that will tie them together and should work locally on any machine.
```
interval: 5s
managers:
  test:
    monitor:
      name: fake
    inventory:
      name: fake
    strategy:
      name: threshold
      thresholds:
        my.arbitrary.metric.name:
          max: 65
          min: 35
```
With this file saved as `./config/config.yaml`, you should be able to run Alice.

## Testing

Make sure your code works with the [fakes](#running-alice-locally). If you're working on a new monitor, run
it with a fake inventory, and the reverse if you're working on an inventory.

Make sure you have an `*_test.go` file with reasonable coverage and that `go test -race -cover $(go list ./... | grep -v /vendor/)` passes.

## Submitting Code

We would love some extra hands to help improve the Alice.
If you want to help, please try to follow the following workflow to help make PRs accepted quickly.

1. Raise an issue
2. Create a branch
3. Write tests
4. Submit a PR
