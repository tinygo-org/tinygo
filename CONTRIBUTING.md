# How to contribute

Thank you for your interest in improving TinyGo.

We would like your help to make this project better, so we appreciate any contributions. See if one of the following descriptions matches your situation:

### New to TinyGo

We'd love to get your feedback on getting started with TinyGo. Run into any difficulty, confusion, or anything else? You are not alone. We want to know about your experience, so we can help the next people. Please open a Github issue with your questions, or you can also get in touch directly with us on our Slack channel at [https://gophers.slack.com/messages/CDJD3SUP6](https://gophers.slack.com/messages/CDJD3SUP6).

### Something in TinyGo is not working as you expect

Please open a Github issue with your problem, and we will be happy to assist.

### Something in Go that you want/need does not appear to be in TinyGo

We probably have not implemented it yet. Please take a look at our [Roadmap](https://github.com/tinygo-org/tinygo/wiki/Roadmap). Your pull request adding the functionality to TinyGo would be greatly appreciated.

Please open a Github issue. We want to help, and also make sure that there is no duplications of efforts. Sometimes what you need is already being worked on by someone else.

A long tail of small (and large) language features haven't been implemented yet. In almost all cases, the compiler will show a `todo:` error from `compiler/compiler.go` when you try to use it. You can try implementing it, or open a bug report with a small code sample that fails to compile.

### Some specific hardware you want to use does not appear to be in TinyGo

As above, we probably have not implemented it yet. Your contribution adding the hardware support to TinyGo would be greatly appreciated.

Please start by opening a Github issue. We want to help you to help us to help you.

Lots of targets/boards are still unsupported. Adding an architecture often requires a few compiler changes, but if the architecture is supported you can try implementing support for a new chip or board in `src/runtime`. For details, see [this wiki entry on adding archs/chips/boards](https://github.com/tinygo-org/tinygo/wiki/Adding-a-new-board).

Microcontrollers have lots of peripherals (I2C, SPI, ADC, etc.) and many don't have an implementation yet in the `machine` package. Adding support for new peripherals is very useful.

## How to use our Github repository

The `release` branch of this repo will always have the latest released version of TinyGo. All of the active development work for the next release will take place in the `dev` branch. TinyGo will use semantic versioning and will create a tag/release for each release.

Here is how to contribute back some code or documentation:

- Fork repo
- Create a feature branch off of the `dev` branch
- Make some useful change
- Make sure the tests still pass
- Submit a pull request against the `dev` branch.
- Be kind

## How to run tests

To run the tests:

```
make test
```
