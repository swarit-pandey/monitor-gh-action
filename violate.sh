#!/bin/bash

echo "Starting KubeArmor policy test script..."

# Test umount
sudo umount /mnt 2>/dev/null

# Test chage
sudo chage -l root

# Test chown
sudo chown root:root test.txt 2>/dev/null

# Test usermod
sudo usermod -L testuser 2>/dev/null

# Test sudo
sudo echo "Testing sudo"

# Test crontab
crontab -l

# Test sudoedit
sudo -e /etc/hosts 2>/dev/null

# Test pam_timestamp_check
sudo /usr/sbin/pam_timestamp_check 2>/dev/null

# Test chsh
sudo chsh -s /bin/bash testuser 2>/dev/null

# Test newgrp
newgrp users

# Test sensitive syscalls
touch test_file
rm test_file
mkdir test_dir
rmdir test_dir

echo "Test script completed."
