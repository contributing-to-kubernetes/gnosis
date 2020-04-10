# Welcome!

# Contributing to Kubernetes

## Welcome!

Here, you will find lots of examples of work that has been done in Kubernetes.
These examples are made for new contributors (even those who may not be new to
Go).

The aim of this is to, by revisiting work that has been completed or is in progress,
help you gain the hands-on experience necessary for you to become a Kubernetes maintainer!
We believe that through thorough analysis and understanding of the code going into Kubernetes,
we can tap into a valuable and readily available resource that will help us to be even
more effective contributors, for both experienced and inexperienced contributors.

## Table of Contents
* [General](./general): description of how the Kubernetes community is laid out
  and general information you may need to get started.
* Hands On Examples
  * [API Machinery: Fix bug in apiserver service cidr split](./kk-pr-85968)
  * [Network: Kubenet fetches gateway from CNI result instead of calculating gateway from pod cidr](./kk-pr-85993)
  * [Testing: Understanding E2E Tests](./e2e-tests)

## How can I make a PR?
- Create an issue, if there isn't already one. Following the guidelines of this template to make an issue:

```
PR
kubernetes/kubernetes#85898

SIG Labels
/sig node
/sig network
/area kubelet
```

- Create a feature branch (on the upstream repo) with the issue and it's number, so that it matches the issue you intend to write documentation on. If the issue is `"issue #23"`, your branch name should be `issue-23`.

- Make iterative changes to the branch by making PRs against that branch.

- When a PR is done being reviewed, merge it via the **squash and merge** github button, into the issue-X branch (to easily avoid any merge commit messages).

- Once `"issue-23"` is ready to be merged, clean it up and merge the final result into master.
