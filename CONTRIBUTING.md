# :star2: How to contribute to the Styra Controller Project :star2:

Thank you for considering to contribute to this project! We appreciate your
interest in helping make it better. Before you get started, please take a
moment to read through these guidelines to help make the contribution process
smooth and effective for everyone involved.

- [Code of Conduct](#code-of-conduct)
- [Questions about the Project](#questions-about-the-project)
- [Issues](#issues)
- [Pull requests](#pull-requests)
- [Commits](#commits)
- [Coding Conventions](#coding-conventions)
- [How to run the project](#how-to-run-the-project)
  - [Deploy to existing cluster](#deploy-to-existing-cluster)
  - [Deploy to local kind cluster](#deploy-to-local-kind-cluster)

## Code of Conduct

This project has a Code of Conduct. By participating in this project, you agree
to abide by its terms. You can find the Code of Conduct in the
[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) file.

## Questions about the Project

We use 
[GitHub Discussions](https://github.com/Bankdata/styra-controller/discussions)
for questions about the project. Before starting a new discussion, please check
if there is an existing discussion on the same topic. If there isn't a
discussion that covers what you're interested in, feel free to start a new one.
To help communicate the nature of your discussion, you can use
[gitmoji](https://gitmoji.dev/).

## Issues

If you encounter a problem or think you've found a bug, please let us know by
opening a new issue with as much detail as possible. You can search existing
issues to see if it has already been reported. If you want to propose a new
feature, please create an issue as well.

To help communicate the nature of the issue, please use a
[gitmoji](https://gitmoji.dev/).

Creating an issue first can save you time compared to making a pull request
without discussing the issue beforehand.

## Pull requests

If your PR addresses an existing issue, please refer to it. Using the
appropriate [gitmoji](https://gitmoji.dev/) can also help to communicate the
nature of the change.

We expect all code to be well-tested, and we kindly recommend that you review
your own code before submitting it for review. This will help catch any
mistakes early and reduce the reviewer's workload.

As you update your PR, please take the time to mark any resolved conversations
as such. If it's not clear why a conversation was resolved, please leave a
comment explaining the reasoning.

To ensure a clean git history, please squash all commits in your PR into a
single commit. Additionally, we require all commits to be signed with a GPG
signature.

If a pull request is still a work in progress, please mark it as a draft PR.

## Commits

For small changes, a one-line commit message is sufficient. However, for larger
changes, please use the following format for your commit message:

```
:emoji: Brief summary of the commit

A paragraph describing the changes made and why they were made.
```

Either way, remember to start the commit message with the appropriate
[gitmoji](https://gitmoji.dev/).

## Coding Conventions

Most of the coding conventions are enforced by the linter. However, if you have
any doubts, please take a look at the existing code and try to follow the same
coding style. Consistent coding conventions will make it easier for others to
understand and contribute to the project.

## How to run the project

To see the list of available targets for working with the project, you can
check the Makefile. You can also run `make help` to get a brief overview of the
available targets.

### Deploy to existing cluster

The Makefile has targets for building and deploying the controller to a
Kubernetes cluster.

1. Set desired docker image name using the IMG environment variable: `export IMG=docker-image-name`. This will be used across make targets that refer to
   image names.
2. Build the controller docker image: `make docker-build`
3. Push the controller docker image: `make docker-push`
4. Install the CRDs in the cluster: `make install`
5. Deploy the controller: `make deploy`. The configuration options for the
   controller are defined in `/config/default/config.yaml`. At the very least,
   you will need to provide the Styra API URL and a token.
6. Deploy instances of the CRDs. Examples are found in `/config/samples`.

### Deploy to local kind cluster

If you want to run the controller locally on your machine, this can be done
using [kind](https://kind.sigs.k8s.io/). The Makefile has a few targets for
easily using kind.

1. Set desired docker image name: `export IMG=docker-image-name`
2. Create a kind cluster with cert-manager installed: `make kind-create`
3. Build the controller docker image and load it into the cluster: 
   `make kind-load`
5. Install the CRDs in the cluster: `make install`
6. Deploy the controller: `make deploy`. The configuration options for the
   controller are defined in `/config/default/config.yaml`. At the very least,
   you will need to provide the Styra API URL and a token.
7. Deploy instances of the CRDs. Examples are found in `/config/samples`.
