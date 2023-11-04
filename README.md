# Gondola

Sick of your kids' DVD's getting scratched and unusable?

Gondola is a media center that is designed to work from an old laptop, or a cheap+silent single board computer (SBC) like
a [Chip](https://getchip.com/) or a [Raspberry Pi](https://www.raspberrypi.org/).

It accomplishes this feat by pre-processing your media into [HLS](https://developer.apple.com/streaming/),
then serving it using Nginx. Pre-processing can take a very long time eg overnight on a Chip, so the recommended use case for this is to make backups of DVDs that you're likely to watch more than once. Eg your kids' movies, so you don't have to worry about the discs getting scratched.

## Features

* Cheap - you don't need to buy an expensive computer that's fast enough to transcode in real time.
* Not hot - my old media center in my garage gets quite hot, and I worry about it in summer! This one won't.
* Silent - my old media center spins its fans all day - this one won't, as most SBC's have no fan.
* Simple - therefore, hopefully more reliable than the other common alternatives.
* Seekable - because it pre-processes your media into HLS, which makes individual files for every few seconds, your media seeks perfectly (important for kids!).
* Just drop your eg VOB files into a 'New' folder using eg [ForkLift](http://www.binarynights.com/Forklift/), and it'll 
wait until transfer has complete to begin importing it automatically.
* Great if you've got limited bandwidth and can't let your kids watch eg Netflix.
* Much safer to have your kids watching your own library rather than randomly browsing Youtube - who knows what they'll come across.
* Doesn't poll the 'new' folder for new files, instead listens for updates. So you can use a normal non-SSD external hard drive and it'll allow it to go to sleep.

## Drawbacks

* Media must be pre-processed, which can take a long time if it's high quality. Eg I tried a 2-hour 1080p movie, and my Chip took 40 hours to transcode it. This is why I recommend this for movies you'll watch over and over again, eg your kids' movies. You will likely find it to be an order of magnitude faster if you use an old laptop, but it'll be noisier / use more electricity.

## Notes

* Gondola, after transcoding to HLS, removes the source file. The assumption is that the user ripped their original from their DVD so doesn't care to lose it. Plus this saves storage space.

## Config

Configuration is compulsory, and goes into `~/.gondola`

It uses TOML format (same as windows INI files). Options include:

`root = "~/Some/Folder/Where/I/Want/My/Data/To/Go"`

This allows you to disable the transcoding, which is useful to speed up dev.

`debugSkipHLS = true`

## File naming conventions

When you dump a movie into the 'New/Movies' folder, the following will work:

	* Big.Buck.Bunny.2008.1080p.blah.vob
	* Big Buck Bunny 2008 1080p blah.vob
	* Big.Buck.Bunny.vob
	* Big.Buck.Bunny.2008.deinterlace.vob
	* Big.Buck.Bunny.2008.scalecrop1080.vob
	* Big.Buck.Bunny.2008.scalecrop1920_940.vob
	* Big.Buck.Bunny.2008.scalecrop239letterbox1080.vob
	* Big.Buck.Bunny.2008.scalecrop239letterbox1920_940.vob
	* !Big Buck Bunny.2008.vob (this is for movies that cannot be found on TMDB; you will need to supply images yourself)

If it finds a year, it assumes the text to the left is the title. Text to the right is ignored, as it's usually resolution/codec/other stuff. Dots/periods are converted to spaces, which it then uses to search TMDB for the movie metadata.

If it cannot find a year, it still searches TMDB to find the movie, but it stands less of a chance finding the correct movie if there's no year.

If it finds 'deinterlace' then it uses FFMPEG to deinterlace the video. This is useful for old DVDs.

If 'scalecrop1080' is found, it scales to 1080p high, then takes only the center 1920 columns, discarding some content to the left and right outside of the 1920. This is handy when you have eg very-widescreen 4k input, and you want it to completely fill your TV, and prefer cropping off the right and left sides a little. Use this if there are no letterbox black bars baked into the input.

Use 'scalecrop1920_940' in the same way you'd use 'scalecrop1080' for wide 4k inputs that you'd like to crop a little off the sides, but is a compromise: You're still letterboxed, just that the black bars are half the height. Probably a good option for epic movies.

Use 'scalecrop239letterbox1080' for the same purpose as scalecrop1080, however this is for when letterbox black bars are baked into the input. This assumes the part we want to keep after cropping the black bars is 1:2.39 ratio.

Use 'scalecrop239letterbox1920_940' to get the same output as scalecrop1920_940, but if the input has the letterbox bars baked in.

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

## Installation on a CHIP

I don't really recommend using a CHIP, as I find it simply transcodes too slowly (13 hours to transcode 1 hour of movie). But if you don't mind, here you go:

* Buy a chip from [getchip.com](https://getchip.com/pages/store)
* Flash it to the latest 'headless' version [here](http://flash.getchip.com/). Headless leaves more resources available for Gondola.
* Plug it into your TV and connect a USB keyboard (alternatively, connect to your computer via the USB port and connect to it via serial using eg screen - this can be less reliable IME than connecting to TV). Username and password default to 'chip'.
* Change the hostname to 'gondola':
	* sudo nano /etc/hostname <- change 'chip' to 'gondola'
	* sudo nano /etc/hosts <- change 'chip' to 'gondola'
	* sudo reboot
* Set it up for your wifi, [read more here](http://docs.getchip.com/chip.html#wifi-connection).
	* nmcli device wifi list
	* sudo nmcli device wifi connect '(your wifi network name/SSID)' password '(your wifi password)' ifname wlan0
	* Eg `sudo nmcli device wifi connect 'MyWifi' password 'MyPassword' ifname wlan0`
	* nmcli device status
* Then you can ssh in from your mac's terminal:
	* `ssh chip@gondola`
	* If it didn't work, go back to the tv/serial, and install zeroconf so you won't need to know its IP address:
		* sudo apt-get update
		* sudo apt-get install avahi-daemon
* Once you can SSH in, you can disconnect it from the TV/Mac (for serial-over-usb) and plug it in to just a plain USB wall power adapter.
* Connecting to the Chip's USB input limits it to pulling 500mA, because the Chip's designers didn't want it to short-circuit people's laptops by pulling too much current. But this means it cannot reliably power an external USB flash drive. You have some options here:
	* You can use a powered USB hub to supply power to your storage drive.
	* If you strip a (good quality) USB cable and connect to CHG-IN(+) and GND(-) on the Chip, it will pull more power. You'll need to plug into a good quality 10 Watt or 2 Amp power adaptor (such as an iPad charger).
	* You can also try the 'no limit' setting, for the same effect as the above option without stripping wires:

	sudo axp209 --no-limit
	sudo apt-get update
	sudo apt-get upgrade
	sudo systemctl enable no-limit
	sudo reboot

	* Use a battery backup for load spikes, I believe [this will plug directly into the Chip](https://www.sparkfun.com/products/8483).
	* I've also tried soldering a capacitor between the +5V line on the USB and the ground to handle any current spikes. Only do this if you're confident!
* Whichever option above you choose, ensure you use a quality usb power supply (eg an apple/samsung genuine one). The cheap ones have uneven voltages and don't supply the promised amps, and your hardware will be flaky as a result (random crashes / data corruption). Also make sure you use a good quality USB cable, because the cheap ones don't have thick enough copper to deliver the necessary amps.
* Connect a large capacity USB drive to your Chip that you're happy to erase and reformat, and we'll configure it next. Make sure you've addressed your power sourcing in the steps above before you connect a drive or you'll get brownouts.
* Formatting steps:
	* I recommend a big flash drive because they're cheap and silent and don't need to turn on/off (so no slow startups). But you can use a normal hard drive if you like.
	* We need to format it as EXT. FAT32 is limited to 4GB files which will be a limitation, and ExFat isn't supported on Chip without a kernel recompile. It's a pity we can't use APFS yet.
	* Find the name of the device: `sudo fdisk -l | grep Disk`
	* Look for some device with lots of GB's. It'll be lower than you expect because of decimal vs binary GB measurements. Mine is `/dev/sda`.
	* If you like, you can use `sudo fdisk -l /dev/sda` (replace sda with whatever you discovered above) to see all the partitions currently on your USB drive. There's probably a FAT one.
	* We need to make a new partition table next:
		* `sudo fdisk /dev/sda` <- replace sda with whatever you discovered above.
		* `g` <- creates a gpt partition table
		* `n` <- creates a new partition. Accept all the defaults by hitting return a few times.
		* `p` <- prints the details, it should look like this, note the device name:

	Disk /dev/sda: 57.8 GiB, 62087233536 bytes, 121264128 sectors
	Units: sectors of 1 * 512 = 512 bytes
	Sector size (logical/physical): 512 bytes / 512 bytes
	I/O size (minimum/optimal): 512 bytes / 512 bytes
	Disklabel type: gpt
	Disk identifier: C7C3585E-CFC5-4A36-B61D-6B3B880F06DC

	Device     Start       End   Sectors  Size Type
	/dev/sda1   2048 121264094 121262047 57.8G Linux filesystem

		* `w` <- to write and exit
	* You should now be back at the normal command prompt.
	* Format it to EXT format like so: `sudo mkfs.ext4 /dev/sda1` <- replacing sda1 with whatever is above.
* Mounting steps:
	* `sudo mkdir /media/usb` <- creates a 'mount point'
	* `sudo chown chip /media/usb` <- allow chip to write to it.
	* `sudo chgrp chip /media/usb`
	* `sudo nano /etc/fstab`
		* Add the line `/dev/sda1 /media/usb ext4 defaults 0 0` <- replace /dev/sda1 with whatever is above.
	* Mount it automatically now: `sudo mount -a`
	* Test it worked: `mount | grep usb`
* Install go:
	* `sudo apt-get install git`
	* You'll need latest golang, the normal version won't compile with this error: `No such file or directory: textflag.h`
		* `sudo nano /etc/apt/sources.list`
			* Add a line: `deb http://httpredir.debian.org/debian experimental main`
			* Add a line: `deb http://httpredir.debian.org/debian unstable main`
		* `sudo apt-get update`
		* `sudo apt-get install golang`
		* It might pop up a blue screen asking to restart some service, just select yes.
	* `nano ~/.bash_profile`
		* add `export GOPATH=$HOME/go`
	* `source ~/.bash_profile` <- reload the profile
	* `env | grep go` <- test it worked
* Allow password-less sudo access to `lsof` so Gondola can use it to determine when uploads are complete:
	* `sudo apt-get install lsof` <- If lsof isn't already installed.
	* `sudo visudo -f /etc/sudoers.d/lsof`
		* add `chip ALL = (root) NOPASSWD: /usr/bin/lsof`
* Now install ffmpeg:
	* `sudo apt-get install ffmpeg`
* Now we can install Gondola:
	* `go get github.com/chrishulbert/gondola`
	* Add a configuration file:
		* `nano ~/.gondola`
		* Paste: `root = "/media/usb/Gondola"`, save and quit.
	* Test it: `~/go/bin/gondola` <- it should say 'Watching for changes'. Do Ctrl+C to close.
* Make it run as a service:
	* `sudo nano /lib/systemd/system/gondola.service`
	* Paste the following:

		[Unit]
		Description=Gondola media server

		[Service]
		PIDFile=/tmp/gondola.pid
		User=chip
		Group=chip
		ExecStart=/home/chip/go/bin/gondola

		[Install]
		WantedBy=multi-user.target

	* `sudo systemctl enable gondola` <- make it run on boot
	* `sudo systemctl start gondola` <- make it start now
	* `systemctl status gondola` <- it should be 'active (running)'
	* `sudo journalctl -u gondola` <- view its logs, should say 'watching for changes'
* Install Nginx:
	* `sudo apt-get install nginx`
	* `sudo nano /etc/nginx/sites-available/default`
		* Find 'root' and change line to: `root /media/usb/Gondola`
	* `sudo nginx -s reload` <- restart nginx.
	* Open [http://gondola](http://gondola) in Safari on your Mac/iPhone/iPad (Chrome doesn't support HLS) and you should see something!
* Upload your first media:
	* I recommend using [ForkLift](http://www.binarynights.com/Forklift/), but you can use any SCP-capable app on your mac/pc.
	* Go to favourites, click '+'. Protocol: 'SFTP'; Name: Gondola; Server: gondola; Username: chip; Password: chip; Remote path: /media/usb/Gondola
	* Connect, and drop something into `New/TV` or `New/Movies`, as per the file naming conventions described elsewhere here.
	* Check the logs on your Chip using `sudo journalctl -u gondola | tail`.
	* Have a look in the `Gondola/Staging` folder while it works.
	* Wait a (long) while for it to convert... For an idea, a 2 hour 1080p movie took over a day.
	* While it's converting you can use `top` to see that ffmpeg is hogging the CPU. Once it disappears from top, you'll know it's done.
	* Open [http://gondola](http://gondola) in Safari and you should be golden!

Good luck.

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
