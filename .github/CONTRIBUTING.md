# Contributing

Thank you for considering to help out with the source code! We welcome
contributions from anyone on the internet, and are grateful for even the
smallest of fixes!

If you'd like to contribute to Pactus, please fork, fix, commit and send a
pull request for the maintainers to review and merge into the main code base. If
you wish to submit more complex changes though, please check up with the core
devs first on [our discord channel](https://discord.gg/zPqWqV85ch) to
ensure those changes are in line with the general philosophy of the project
and/or get some early feedback which can make both your efforts much lighter as
well as our review and merge procedures quick and simple.

## Coding guidelines

Please make sure your contributions adhere to our coding guidelines:

 * Code must adhere to the official Go
[formatting](https://golang.org/doc/effective_go.html#formatting) guidelines
(i.e. uses [gofmt](https://golang.org/cmd/gofmt/) and `make fmt` can help you).
 * Code must be documented adhering to the official Go
[commentary](https://golang.org/doc/effective_go.html#commentary) guidelines.
 * Pull requests need to be based on and opened against the `main` branch.
 * Commit messages should be prefixed with the package(s) they modify.
   * E.g. "p2p, rpc: make trace configs optional"
* Error strings and log messages should not be capitalized (unless beginning with proper nouns or acronyms) or end with
punctuation

## Can I have feature X

Before you submit a feature request, please check and make sure that it isn't
possible through some other means. Please check our
[documentation page](https://pactus.org/guide/learn-index.html) for more info
and help.

## Configuration, dependencies, and tests

Please see the [Developers' Guide](https://pactus.org/dev/dev-index.html)
for more details on configuring your environment, managing project dependencies
and testing procedures.
