#!/bin/bash

# copy systemd service unit file into the correct directory and enable on boot
sudo cp discord-voice-monitor.service /lib/systemd/system/
sudo systemctl enable discord-voice-monitor
sudo service discord-voice-monitor start