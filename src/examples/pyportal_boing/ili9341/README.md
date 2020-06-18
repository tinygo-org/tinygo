TinyGo driver for TFT displays using ILI9341 driver chips.

These displays support 8-bit parallel, 16-bit parallel, or SPI interfaces.

Examples of such displays include:

 * [Adafruit PyPortal
 ](https://www.adafruit.com/product/4116)
 * [Adafruit 2.8" Touch Shield V2 (SPI)](http://www.adafruit.com/products/1651)
 * [Adafruit 2.4" TFT LCD with Touchscreen Breakout w/MicroSD Socket](https://www.adafruit.com/product/2478)
 * [2.8" TFT LCD with Touchscreen Breakout Board w/MicroSD Socket](https://www.adafruit.com/product/1770)
 * [2.2" 18-bit color TFT LCD display with microSD card breakout](https://www.adafruit.com/product/1770)
 * [TFT FeatherWing - 2.4" 320x240 Touchscreen For All Feathers](https://www.adafruit.com/product/3315)

Currently this driver only supports an 8-bit parallel interface using ATSAMD51
(this is the default configuration on PyPortal).  It should be relatively
straightforward to implement a more generic SPI-based interface as well.
Please see `parallel_atsamd51.go` for an example of what needs to be
implemented if you are interested in contributing.
