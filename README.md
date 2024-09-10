# Gondola

Sick of your kids' DVD's getting scratched and unusable?

Gondola is a media center that is designed to work from an old laptop, or a cheap+silent single board computer (SBC) like a [Raspberry Pi](https://www.raspberrypi.org/), or a Mac Mini.

It accomplishes this feat by pre-processing your media into [HLS](https://developer.apple.com/streaming/),
then serving it using Nginx. Pre-processing can take a very long time eg overnight on a slow SBC, so the recommended use case for this is to make backups of DVDs that you're likely to watch more than once. Eg your kids' movies, so you don't have to worry about the discs getting scratched.

## Features

* Cheap - you don't need to buy an expensive computer that's fast enough to transcode in real time.
* Not hot - my old media center in my garage used to get quite hot, and I worried about it in summer! This one won't.
* Silent - my old media center spun its fans all day - this one won't, as most SBC's have no fan.
* Simple - therefore, hopefully more reliable than the other common alternatives.
* Seekable - because it pre-processes your media into HLS, which makes individual files for every few seconds, your media seeks perfectly (important for kids!).
* Just drop your eg VOB files into a 'New' folder using eg [ForkLift](http://www.binarynights.com/Forklift/), and it'll 
wait until transfer has complete to begin converting it automatically.
* Great if you've got limited bandwidth and can't let your kids watch eg Netflix.
* Much safer to have your kids watching your own library rather than randomly browsing Youtube - who knows what they'll come across.
* Doesn't poll the 'new' folder for new files, instead listens for updates. So you can use a normal non-SSD external hard drive and it'll allow it to go to sleep.

## Drawbacks

* Media must be pre-processed, which can take a long time if it's high quality. Eg I tried a 2-hour 1080p movie, and a slow SBC took 40 hours to transcode it. This is why I recommend this for movies you'll watch over and over again, eg your kids' movies. You will likely find it to be an order of magnitude faster if you use an old laptop, but it'll be noisier / use more electricity.
* You can run Gondola on a fast desktop then copy the resulting converted files across to mitigate this speed issue.

## Notes

* Gondola, after transcoding to HLS, removes the source file. The assumption is that the user ripped their original from their DVD so doesn't care to lose it. Plus this saves storage space.

## Config

Configuration is compulsory, and goes into `~/.gondola`

It uses TOML format (same as windows INI files). Options include:

`root = "~/Some/Folder/Where/I/Want/My/Data/To/Go/Gondola"`

This allows you to disable the transcoding, which is useful to speed up dev.

`debugSkipHLS = true`

## File naming conventions

When you dump a movie into the 'New/Movies' folder, the following will work:

	* Big.Buck.Bunny.2008.1080p.blah.vob
	* Big Buck Bunny 2008 1080p blah.vob
	* Big.Buck.Bunny.vob
	* Big.Buck.Bunny.2008.deinterlace.vob
	* Big.Buck.Bunny.2008.scalecrop1080.vob
	* Big.Buck.Bunny.2008.crop1920_940Ratio.vob
	* Big.Buck.Bunny.2008.scalecrop1920_940.vob
	* Big.Buck.Bunny.2008.scalecrop239letterbox1080.vob
	* Big.Buck.Bunny.2008.scalecrop239letterbox1920_940.vob
	* Big.Buck.Bunny.2008.scaleInside1920_1080MaintainingRatio.vob
	* !Big Buck Bunny.2008.vob (this is for movies that cannot be found on TMDB; you will need to supply images yourself)

If it finds a year, it assumes the text to the left is the title. Text to the right is ignored, as it's usually resolution/codec/other stuff. Dots/periods are converted to spaces, which it then uses to search TMDB for the movie metadata.

If it cannot find a year, it still searches TMDB to find the movie, but it stands less of a chance finding the correct movie if there's no year.

If it finds 'deinterlace' then it uses FFMPEG to deinterlace the video. This is useful for old DVDs.

If 'scalecrop1080' is found, it scales to 1080p high, then takes only the center 1920 columns, discarding some content to the left and right outside of the 1920. This is handy when you have eg very-widescreen 4k input, and you want it to completely fill your TV, and prefer cropping off the right and left sides a little. Use this if there are no letterbox black bars baked into the input.

Use 'scalecrop1920_940' in the same way you'd use 'scalecrop1080' for wide 4k inputs that you'd like to crop a little off the sides, but is a compromise: You're still letterboxed, just that the black bars are half the height. Probably a good option for epic movies.

Use 'crop1920_940Ratio' similarly to scalecrop1920_940 for inputs that are 1920 wide so you don't want to scale them down, and of eg 1:2.39 ratio, so it crops a bit off the left and right so that it fills the screen a bit more (but not entirely, so you still have the movie effect).

Use 'cropScaleDown4kWideToUnivisium' For if the input ratio is >= 1:2.1, crops down to 1:2 to fill the tv nicely, then scales to 1920w.

Use 'scalecrop239letterbox1080' for the same purpose as scalecrop1080, however this is for when letterbox black bars are baked into the input. This assumes the part we want to keep after cropping the black bars is 1:2.39 ratio.

Use 'scalecrop239letterbox1920_940' to get the same output as scalecrop1920_940, but if the input has the letterbox bars baked in.

Use 'crop240LetterboxThenUnivisium' to crop out baked-in 1:2.4 letterbox bars, then crop again to univisium 2:1. Useful for DVDs.

Use 'crop235LetterboxThenUnivisiumThen1920' for 4k inputs, to crop out baked-in 1:2.35 letterbox bars, then crop again to univisium 2:1, then scale to 1920x960.

Use 'crop4k240LetterboxThenUnivisiumThen1920' for 4k inputs, to crop out baked-in 1:2.40 letterbox bars, then crop again to univisium 2:1, then scale to 1920x960.

Use 'crop240LetterboxThen169' to crop out baked-in 1:2.40 letterbox bars, then crop again to 16:9.

Use 'crop235LetterboxThen169' to crop out baked-in 1:2.35 letterbox bars, then crop again to 16:9.

Use 'scaleInside1920_1080MaintainingRatio' to shrink 4k input to 1080p, maintaining aspect ratio, so the height will potentially be < 1080 if it's wider than 16:9. Nothing is cropped.

For TV shows placed in `New/TV` folder, use the following:

	* Some.TV.Show.S01E02.DVD.vob
	* Some TV Show S01E02 Blah blah blah.vob
	* Some TV Show - Episode Name.vob

So long as it can find 'SxxEyy' (for season x episode y), it assumes the show's title is to the left, and ignores anything to the right. It then searches TMDB to find the show's metadata.

If you use the last format (eg there's no SxEy, and there is a `-`), it takes a 'best guess' at which season and episode it is. This is useful for shows where the DVD order is different to the TV order. I actually recommend using this when you make backups of your DVDs. The episode name doesn't need to be an exact match, because it does a 'Levenshtein distance' calculation to make a best guess at which episode it is.

But it forces you to confirm it guessed correctly: the file is renamed to the best guess, with a `.remove if correct` extension attached. If you're happy with the guess, rename the file to remove the extension, and it'll process as usual. Eg if you upload `Seinfeld - Serenity.vob`, it'll rename it to `Seinfeld S09E03 The Serenity Now.Seinfeld - Serenity.vob.remove if correct`. The first half of that is the guessed episode's number and it's name according to TMDB, then the original name you gave the file, then the remove_if_correct extension for you to remove as a confirmation that you're happy.

### TV shows without TMDB lookup

Since the TMDB lookup tends to fail now, you can use the following naming convention:

	* !Show name - S1 Season 1 - E2 Episode title.deinterlace.vob

Obviously this will not be able to look up metadata, so it is up to you to provide images in this case.

## Name

The name is a (tortured) metaphor: A real gondola transports you down a stream; this Gondola transports your media by streaming it ;)

## Installation on an Orange Pi PC

An Orange Pi PC is a good choice because it has 1GB RAM, which is enough for 1080p transcoding; and it has a good power supply thus can power an external USB HDD.

* Buy stuff!
	* The Orange Pi itself: http://www.orangepi.org/orangepipc/
	* Buy a small self adhesive heatsink from eBay, and stick it on the Orange Pi's main chip.
		* Search eBay for `22x22x25mm adhesive heatsink`
	* Grab a name-brand SD card, 4GB will be fine, 2GB might be worth trying if you have one lying around (FWIW mine uses 1.6G once all configured).
		* Did i say make sure it's a name-brand one? Avoid eBay unless you're very sure its a genuine card.
* I power mine using a USB -> OrangePi barrel connector cable, connected to a 5V 2A iPad wall charger, and recommend a similar setup.
* Install the operating system:
	* Grab the software from here, I'm trying one of the 4.x 'nightly releases' because 3.x apparently has out-of-memory bugs, prefer the 'server' builds if you get a choice: https://www.armbian.com/orange-pi-pc/
	* Extract the 7z file using eg The Unarchiver from the mac app store
	* Copy the IMG file onto your sd card using Etcher.io
* Plug the card into your Orange Pi, and plug it into your TV & Ethernet & Keyboard & Power
	* Login as 'root' password '1234'
	* Change root password when prompted to 'gondola1'
	* Follow the steps to create the 'gondola' user
	* `sudo apt-get update` <- If you get a lock error, wait 5 mins and try again, it might be auto-updating.
	* `sudo apt-get install avahi-daemon` <- This gets bonjour working, so macs can find it.
	* `sudo nano /etc/hostname` <- change the host name 'gondola'.
	* `sudo nano /etc/hosts` <- change 'orangepipc' to 'gondola' in this file.
	* `sudo reboot now` <- applies the hostname change.
* Login remotely now and continue configuring:
	* `ssh gondola@gondola.local` <- login from your mac.
	* `sudo dpkg-reconfigure tzdata` <- set up the timezones.
* Configure the external USB hard drive:
	* `sudo mkdir /media/usb` <- Make a mount point for the external hdd.
	* `sudo chown gondola /media/usb` <- Set the permissions for the mount.
	* `sudo chgrp gondola /media/usb`
	* `sudo nano /etc/fstab` <- Configure the auto-mount.
		* Add the line `/dev/sda1 /media/usb auto defaults,noatime,uid=gondola,gid=gondola 0 0`
		* Note: `noatime` above makes it not write an updated time when you read, which is faster and causes less wear if it's a flash drive.
		* Note: `uid=gondola,gid=gondola` gives the proper permissions if it's a FAT drive.
	* `sudo mount -a` <- Auto mount.
* Install stuff:
	* Enable non-ancient apt (mainly for golang): `sudo nano /etc/apt/sources.list`
	  * Add: `deb http://httpredir.debian.org/debian experimental main`
	  * Add: `deb http://httpredir.debian.org/debian unstable main`
	* Install lots of stuff: `sudo apt-get install git golang lsof ffmpeg nginx`
* Configure Go:
  * `nano ~/.bash_profile`
	  * Add a line: `export GOPATH=$HOME/go`
  * `source ~/.bash_profile` <- reload the profile
	* `env | grep go` <- test it worked
*  Allow password-less sudo access to lsof so Gondola can use it to determine when uploads are complete:
	* `sudo visudo -f /etc/sudoers.d/lsof`
  * Add a line: `gondola ALL = (root) NOPASSWD: /usr/bin/lsof`
* Install Gondola itself:
	* `go get github.com/chrishulbert/gondola`
  * Add a configuration file:
		* `nano ~/.gondola`
		* Paste `root = "/media/usb/Gondola"`
  * Test it: `~/go/bin/gondola` <- it should say 'Watching for changes'. Do Ctrl+C to close.
	* Make it run as a service:
		* `sudo nano /lib/systemd/system/gondola.service`
		* Paste the following:

		    ```
	        [Unit]
	        Description=Gondola media server

		    [Service]
	        PIDFile=/tmp/gondola.pid
	        User=gondola
	        Group=gondola
	        ExecStart=/home/gondola/go/bin/gondola

	        [Install]
	        WantedBy=multi-user.target
			```

		* `sudo systemctl enable gondola` <- make it run on boot
		* `sudo systemctl start gondola` <- make it start now
		* `systemctl status gondola` <- it should be 'active (running)'
		* `sudo journalctl -u gondola` <- use this to view Gondola's logs, should say 'watching for changes'
* Configure Nginx:
  * `sudo nano /etc/nginx/sites-available/default`
    * Find `root /var/www/html;` and change line to: `root /media/usb/Gondola;` <- customise the path to suit where your hard drive mounts
  * `sudo nano /etc/nginx/nginx.conf`
    * Find `user www-data;` and change to `user gondola;` <- this allows nginx to read your external HDD
	* `sudo nginx -s reload` <- apply the nginx config changes.
  * Open http://gondola.local in Safari on your Mac/iPhone/iPad (Chrome doesn't support HLS) and you should see something!
  * Check your Nginx logs if something fails: `sudo cat /var/log/nginx/error.log`
* Final notes:
	* The Orange Pi PC takes 8.7h per hour of movie to transcode
	* Easy log access:
		* `nano ~/.bash_profile`
		* Add `alias l="sudo journalctl -u gondola | tail -n 100"`
		* `source ~/.bash_profile`
		* Then run `l` anytime you want to see the gondola logs.
	* Good luck!

## Installation on a laptop

Get a second hand laptop (you may have one lying around?). I bought a second hand HP Stream 11 for $120 - I recommend this laptop because it's cheap, small, fanless thus silent, and has a small power adapter thus won't use much electricity.

Install Ubuntu. In the installer, set the computer name to 'gondola', and the user 'gondola'.

If you use an HP Stream, some good instructions for installing and tweaking Ubuntu are [here](http://bernaerts.dyndns.org/linux/74-ubuntu/343-ubuntu-install-hp-stream-13).

In Ubuntu settings > Power, under 'when the lid is closed', set it to 'do nothing'. Now your laptop will continue running as a server when closed. Update: In newer versions, you may follow the instructions here to do the same: https://askubuntu.com/a/372616/655394

Connect your external hard drive (my laptop has one USB3 port and one USB2 - ensure you connect to the faster one). Find where it mounts: in my case, it mounts at /media/gondola/KRYTEN. (Kryten is the name of my external drive).

In the terminal, run the following:

    sudo apt-get install openssh-server

You should now be able to SSH in from your Mac/PC with the following terminal command:

    ssh gondola@gondola.local

If that succeeded, you may now connect your laptop to power, close the lid, and tuck it away somewhere with a bit of clear airflow, and follow the remaining instructions using remote SSH:

* Install go:
	* `sudo apt-get install git`
	* You'll need latest golang, the normal version won't compile with this error: `No such file or directory: textflag.h`
		* `sudo nano /etc/apt/sources.list`
			* Add a line: `deb http://httpredir.debian.org/debian experimental main`
			* Add a line: `deb http://httpredir.debian.org/debian unstable main`
		* `sudo apt-get update`
		* `sudo apt-get install golang`
	* `nano ~/.bash_profile`
		* add `export GOPATH=$HOME/go`
	* `source ~/.bash_profile` <- reload the profile
	* `env | grep go` <- test it worked
* Allow password-less sudo access to `lsof` so Gondola can use it to determine when uploads are complete:
	* `sudo apt-get install lsof` <- If lsof isn't already installed.
	* `sudo visudo -f /etc/sudoers.d/lsof`
		* add `gondola ALL = (root) NOPASSWD: /usr/bin/lsof`
* Now install ffmpeg:
	* `sudo apt-get install ffmpeg`
* Now we can install Gondola:
	* `go get github.com/chrishulbert/gondola`
	* Add a configuration file:
		* `nano ~/.gondola`
		* Paste: `root = "/media/gondola/KRYTEN/Gondola"`, customising the path to suit where your external drive mounts, save and quit.
	* Test it: `~/go/bin/gondola` <- it should say 'Watching for changes'. Do Ctrl+C to close.
* Make it run as a service:
	* `sudo nano /lib/systemd/system/gondola.service`
	* Paste the following:
	
	    ```
        [Unit]
        Description=Gondola media server
	 
	    [Service]
        PIDFile=/tmp/gondola.pid
        User=gondola
        Group=gondola
        ExecStart=/home/gondola/go/bin/gondola
	    
        [Install]
        WantedBy=multi-user.target
		```

	* `sudo systemctl enable gondola` <- make it run on boot
	* `sudo systemctl start gondola` <- make it start now
	* `systemctl status gondola` <- it should be 'active (running)'
	* `sudo journalctl -u gondola` <- view its logs, should say 'watching for changes'
* Install Nginx:
	* `sudo apt-get install nginx`
	* `sudo nano /etc/nginx/sites-available/default`
		* Find `root /var/www/html;` and change line to: `root /media/gondola/KRYTEN/Gondola;` - customise the path to suit where your hard drive mounts.
	* `sudo nano /etc/nginx/nginx.conf`
		* Find `user www-data;` and change to `user gondola;` - this allows nginx to read your external HDD.
	* `sudo nginx -s reload` <- restart nginx.
	* Open [http://gondola](http://gondola) in Safari on your Mac/iPhone/iPad (Chrome doesn't support HLS) and you should see something!
	* Check your nginx logs if something fails: `cat /var/log/nginx/error.log`
* Upload your first media:
	* I recommend using [ForkLift](http://www.binarynights.com/Forklift/), but you can use any SCP-capable app on your mac/pc. Cyberduck is also popular.
	* To make backups of your DVDs, I recommend [dvdbackup](https://wiki.archlinux.org/index.php/dvdbackup).
	* Go to favourites, click '+'. Protocol: 'SFTP'; Name: Gondola; Server: gondola.local; Username: gondola; Password: gondola; Remote path: /media/gondola/KRYTEN/Gondola
	* Connect, and drop something into `New/TV` or `New/Movies`, as per the file naming conventions described elsewhere here.
	* Check the logs on your Chip using `sudo journalctl -u gondola | tail`.
	* Have a look in the `Gondola/Staging` folder while it works.
	* Wait a (long) while for it to convert... For an idea, a 2 hour 1080p movie took over a day.
	* While it's converting you can use `top` to see that ffmpeg is hogging the CPU. Once it disappears from top, you'll know it's done.
	* Open [http://gondola.local](http://gondola.local) in Safari and you should be golden!

## Installation on a Mac Mini (as of 2024)

* Configure the Mac to not go to sleep
* Configure it to be accessible as 'gondola.local'
	* System Settings > General > Sharing > Scroll to bottom > Local Hostname > Edit
* Install golang from go.dev
* Allow password-less sudo access to `lsof` so Gondola can use it to determine when uploads are complete:
	* `whoami` <- figure out your username.
	* `which lsof` <- figure out where lsof is installed.
	* `sudo visudo -f /etc/sudoers.d/lsof`
		* paste `chris ALL = (root) NOPASSWD: /usr/sbin/lsof`
		* Replace 'chris' above with the result of 'whoami' before.
		* Change the path to lsof to match the result of 'which lsof' before.
		* Since this is 'vi', hit escape then type ':wq' then hit return.
* Install brew from `https://brew.sh`
* Now install ffmpeg:
	* `brew install ffmpeg`
* Now we can install Gondola:
	* `go get github.com/chrishulbert/gondola`
	* Add a configuration file:
		* `nano ~/.gondola`
		* Paste: `root = "/Users/chris/Gondola"`, customising the path to suit where your external drive mounts, save and quit.
	* Test it: `~/go/bin/gondola` <- it should say 'Watching for changes'. Do Ctrl+C to close.
* Make it run automatically:
	* Get the user and group:
	* get user: `id -un` eg chris
	* get group: `id -gn` eg staff
	* `sudo nano /Library/LaunchDaemons/gondola.plist`
	* Paste the following, careful to customise the paths and user and group:
	
	    ```
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key><string>au.com.splinter.gondola</string>
    <key>Program</key><string>/Users/chris/dev/gondola/gondola</string>
    <key>StandardOutPath</key><string>/tmp/gondola.log</string>
    <key>StandardErrorPath</key><string>/tmp/gondola.log</string>
    <key>UserName</key><string>chris</string>
    <key>GroupName</key><string>staff</string>
    <key>ExitTimeOut</key><integer>999111222</integer>
    <key>RunAtLoad</key><true/>
  </dict>
</plist>
		```
	* `sudo launchctl load /Library/LaunchDaemons/gondola.plist`
	* `sudo launchctl start /Library/LaunchDaemons/gondola.plist`
	* `launchctl print system/au.com.splinter.gondola`
	* Check for 'state = running'
	* If you want to change the plist, do:
		* `sudo launchctl unload /Library/LaunchDaemons/gondola.plist`
		* change it
		* `sudo launchctl load /Library/LaunchDaemons/gondola.plist`
	* Check its logs: `cat /tmp/gondola.log` it should say 'watching for changes'
* Configure the web server:
    * `go install github.com/m3ng9i/ran@latest`
	* `sudo nano /Library/LaunchDaemons/gondola.http.plist`
	* Paste the following, careful to customise the paths and user and group:
	
	    ```
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key><string>au.com.splinter.gondola.http</string>
    <key>Program</key><string>/Users/chris/go/bin/ran</string>
	<key>ProgramArguments</key>
    <array>
        <string>/Users/chris/go/bin/ran</string>
        <string>-port=80</string>
        <string>-listdir=true</string>
        <string>-gzip=false</string>
      </array>
	<key>WorkingDirectory</key>
    <string>/Volumes/MyExfat/Gondola</string>
    <key>StandardOutPath</key><string>/tmp/ran.log</string>
    <key>StandardErrorPath</key><string>/tmp/ran.err.log</string>
    <key>UserName</key><string>chris</string>
    <key>GroupName</key><string>staff</string>
    <key>ExitTimeOut</key><integer>999111222</integer>	
    <key>KeepAlive</key>
    <dict>
      <key>PathState</key>
      <dict>
        <key>/Volumes/MyExfat/Gondola</key>
        <true/>
      </dict>
    </dict>
  </dict>
</plist>
		```
	* `sudo launchctl load /Library/LaunchDaemons/gondola.http.plist`
	* If you get weird errors, unload and reload it.
	* Open [http://gondola](http://gondola) in Safari on your Mac/iPhone/iPad (Chrome doesn't support HLS) and you should see something!
* Install gondola bookmarker:
	* go install github.com/chrishulbert/gondola-bookmarker@latest
	* `sudo nano /Library/LaunchDaemons/gondola.bookmarker.plist`
	* Paste the following, careful to customise the paths and user and group:
	
	    ```
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key><string>au.com.splinter.gondola.bookmarker</string>
    <key>Program</key><string>/Users/chris/go/bin/gondola-bookmarker</string>
    <key>UserName</key><string>chris</string>
    <key>GroupName</key><string>staff</string>
    <key>ExitTimeOut</key><integer>999111222</integer>
    <key>RunAtLoad</key><true/>
  </dict>
</plist>
		```
	* `sudo launchctl load /Library/LaunchDaemons/gondola.bookmarker.plist`
	* `sudo launchctl list | grep bookmarker`
Good luck!!!

### Other tips

#### Transcoding:

Gondola reduces 5.1 to stereo (with some trickery to ensure the sub channel isn't lost). This is to save space and for simpler playback on my stereo system.

If you upload a huge VOB, don't worry: gondola compresses to a reasonable size during the HLS conversion process.

Speed-wise, converting an average length (90min) 1080p movie takes 7 hours 30 mins on an HP Stream 11. On a GetCHIP, it takes around 20 hours.

#### Viewing logs:

    sudo journalctl -u gondola

#### Updating:

    go get -u github.com/chrishulbert/gondola
    sudo systemctl restart gondola

#### Re-generating:

If you manually move the files around, you can force a metadata re-generation by restarting:

    sudo systemctl restart gondola

#### Serial connection:

To connect to the Chip via USB to your mac, do the following:

* Disconnect it, if connected
* Open the terminal
* Do: ls -1 /dev | grep usb
* Connect it
* Wait a minute or two
* Repeat: ls -1 /dev | grep usb
* See if there's a new device listed
* Connect: screen /dev/cu.usbmodemFA133 115200 <- replace the device with whatever you noticed it to be.
* Log in, do whatever you want to do.
* To quit screen, do: control-a-control-\

If screen fails because you quit without closing properly:
Use fuser to find who has the port open and kill it:
fuser /dev/cu.usbmodemFA133
returns: /dev/cu.usbmodemFA133: 95401
kill it: kill 95401

#### Docker:

I've tried to make this work with Docker, however Gondola needs root access to detect if any other processes are writing to files in the New folder, so I don't recommend spending time investigating this unless you find a solution to that.

## Acknowledgements

This product uses the <a href="https://www.themoviedb.org">TMDb</a> API but is not endorsed or certified by TMDb.

Some icons from Icons8.com
