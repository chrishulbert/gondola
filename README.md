# Gondola

*Note that this project isn't in a usable state yet*

Gondola is a media center that is designed to work on a cheap+silent single board computer (SBC) like
a [Chip](https://getchip.com/) or a [Raspberry Pi](https://www.raspberrypi.org/).

It accomplishes this feat by pre-processing your media into [HLS](https://developer.apple.com/streaming/),
then serving it using nginx. This can take a very long time eg overnight, so the recommended use case for this is to make backups of DVDs that you're likely to watch more than once. Eg your kids' movies, so you don't have to worry about the discs getting scratched.

## Features

* Cheap - you don't need to buy an expensive computer that's fast enough to transcode in real time.
* Not hot - my old media center in my garage gets quite hot, and I worry about it in summer! This one won't.
* Silent - my old media center spins its fans all day - this one won't, as most SBC's have no fan.
* Simple - therefore, hopefully more reliable than the other common alternatives.
* Seekable - because it pre-processes your media into HLS, which makes individual files for every few seconds, your media seeks perfectly (important for kids!).

## Drawbacks

* Media must be pre-processed, which can take a long time if it's high quality. Eg I tried a 2-hour 1080p movie, and my Chip took 2 days to transcode it. This is why I recommend this for movies you'll watch over and over again, eg your kids' movies.

## Name

The name is a tortured metaphor: A real gondola transports you down a stream; this Gondola transports your media by streaming it ;)