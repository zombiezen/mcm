# How to contribute

We'd love to accept your patches and contributions to this project. There are a
just a few small guidelines you need to follow.

## Contributor License Agreement

Contributions to any Google project must be accompanied by a Contributor License
Agreement. This is necessary because you own the copyright to your changes, even
after your contribution becomes part of this project. So this agreement simply
gives us permission to use and redistribute your contributions as part of the
project. Head over to <https://cla.developers.google.com/> to see your current
agreements on file or to sign a new one.

You generally only need to submit a CLA once, so if you've already submitted one
(even if it was for a different project), you probably don't need to do it
again.

## Code reviews

All submissions, including submissions by project members, require review. We
use GitHub pull requests for this purpose. Consult [GitHub Help] for more
information on using pull requests.

As a policy, mcm should always build cleanly and pass tests on every commit.  We
run a [Travis build] that checks before and after merges to enforce this policy.
However, as a courtesy to other contributors, please run `./bazel test //...`
before sending a pull request (this is what the Travis build does).

[GitHub Help]: https://help.github.com/articles/about-pull-requests/
[Travis build]: https://travis-ci.org/zombiezen/mcm
