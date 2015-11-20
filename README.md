# go-circle

This is a very incomplete wrapper for the CircleCI API. Currently we use it to
fetch the latest build for a branch.

You should treat the API as very unstable, library API's that grow from one or
two methods to the whole API tend to not be designed very well, so probably at
some point you will have to create a Client instance or something.

## Token Management

This library will look for your Circle API token in `~/cfg/circleci` and (if
that does not exist, in `~/.circlerc`). The configuration file should look like
this:

```toml
[organizations]

    [organizations.Shyp]
    token = "aabbccddeeff00"
```

You can specify any org name you want.

## Installation

If you just want the binary, [download it from Equinox.io][download] and place
the `circle` file somewhere on your `$PATH`.

If you want to install the project, first set your `$GOPATH` in your
environment (I set it to `~/code/go`), then run

```
go install github.com/Shyp/go-circle/...
```

This should place a `circle` binary in `$GOPATH/bin`, so for me,
`~/code/go/bin/circle`.

[download]: https://dl.equinox.io/shyp/circle/stable

## Wait for tests to pass/fail on a branch

If you want to be notified when your tests finish running, run `circle wait
[branchname]`. The interface for that will certainly change as well; we should
be able to determine which organization/project to run tests for by checking
your Git remotes.

It's pretty neat! Here's a screenshot.

<img src="https://monosnap.com/file/49h2NvVwxDBtHWlphAGiqzdJFDB7xy.png"
alt="CircleCI screenshot">
