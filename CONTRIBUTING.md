# Table of Contents
1. [Contributing to Tyk](#contributing-to-tyk)
2. [Our SLA for issues and bugs](#our-sla-for-issues-and-bugs)
3. [Filling an issue](#filling-an-issue)
4. [Contributor License Agreements](#contributor-license-agreements)
5. [Guidelines for Pull Requests](#guidelines-for-pull-requests)
6. [Project Structure](#project-structure)
7. [Building and Running test](#building-and-running-test)
8. [Coding Conventions](#coding-conventions)
9. [Resources](#resources)

# Contributing to Tyk

**First**: if you're unsure or afraid of anything, just ask or submit the issue or pull request anyway. You won't be yelled at for giving your best effort. The worst that can happen is that you'll be politely asked to change something. We appreciate any sort of contributions and don't want a wall of rules to get in the way of that.

However, for those individuals who want a bit more guidance on the best way to contribute to the project, read on. This document will cover what we're looking for. By addressing all the points we're looking for, it raises the chances we can quickly merge or address your contributions.

### Our SLA for issues and bugs
We do value the time each contributor spends contributing to this repo, and we work hard to make sure we respond to your issues and Pull request as soon as we can.

Below we have outlined 

### Filling an issue 
If you have a question, a problem, or an idea for a new capability, or if you think you found a bug, please [create an
issue](/issues/new).

### Contributor License Agreements

Before we can accept any PR the contributor needs to sign the [TYK CLA](./CLA.md).

Once you are CLA'ed, we'll be able to accept your pull requests. For any issues that you face during this process, please create a GitHub issue explaining the problem, and we will help get it sorted out.

### Guidelines for Pull Requests
We have created a few guidelines to help with creating PR. To make sure these requirements are followed we added them to the PR form as well:

1. When working on an existing issue, simply respond to the issue and express interest in working on it.  This helps other people know that the issue is active and hopefully prevents duplicated efforts.
2. For new ideas or breaking changes it is always better to open an issue and discuss your idea with our team first, before implementing it.
3. Create a small Pull request that addresses a single issue instead of multiple issues at the same time. This will make it possible for the PRS to be reviewed independently.
5. Make sure to run tests locally before submitting a pull request and verify that all of them are passing.
6. Documentation - a new capability or improvement needs to be exposed in the form of documentation which needs to be created before this PR is merged. Please open a ticket in [Tyk docs repo](https://github.com/TykTechnologies/tyk-docs/issues/new?assignees=&labels=enhancement&template=feature_request.md&title=) with all the relevant content and link it to the code PR. We are also happy work with you on creating docs so please let us know. Once the docs are ready and the review has been done we can merge the PR
7. Tips for making sure we review your pull request faster :
    1. Code is documented. Please comment on the code where possible, especially for fields and functions.
    2. Use meaningful commit messages.
    3. keep your pull request up to date with the upstream master to avoid merge conflicts.
    4. Provide a good PR description as a record of what change is being made and why it was made. Link to a GitHub issue if it exists.
    5. Tick all the relevant checkboxes in the PR form
    

### Coding Conventions
We haven't yet published an official code convention guide or a linter. For now please keep the current conventions.

### Development Setup
To contribute, follow these steps:

1.	Fork the repository on GitHub.
2.	Clone your fork to your local machine:
```git
git clone https://github.com/your-username/struct-viewer.git
```

3.	Create a new branch for your feature or bugfix:
```git
git checkout -b my-feature-branch
```

4.	Make your changes and commit them with clear messages:
```git
git commit -m "Add new feature"
```

5.	Push your changes to your fork:
```git
git push origin my-feature-branch
```

6. Submit a pull request to the main repository.


### Resources
- [How to Contribute to Open Source](https://opensource.guide/how-to-contribute/)
- [Using Pull Requests](https://help.github.com/articles/about-pull-requests/)


    
