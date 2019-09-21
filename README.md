# iZettle Daily Reports

This is a small tool to generate daily iZettle reports which then are uploaded to Google Drive. The report generator
can be run with `./run.sh`

## Installation

The report generator requires a go version `>1.13` so a installation script is included for installing
go on a Raspberry PI since the default package repository only includes an outdated version. 
To install go run `./raspberypi-install.sh`

## Configuration

The application can be configured by setting the following environment variables 

```bash
# Your iZettle email adress
IZETTLE_EMAIL=
# Your iZettle password
IZETTLE_PASSWORD=
# The folder where the reports will be generated, this is the folder ID which can be found in the folder url
# https://drive.google.com/drive/u/0/folders/GDRIVE_FOLDER_ID
GDRIVE_FOLDER_ID=
# The client id and secret can be retreived when creating a new application and oAuth2 client
# 1. Create a new application with application type other
#    https://console.developers.google.com/projectcreate
# 2. Add the Google Drive API 
#    https://console.developers.google.com/apis/library/drive.googleapis.com
# 3. Create a new oAuth2 Client 
#    https://console.developers.google.com/apis/credentials
CLIENT_ID=
# The secret which is shown when creating the application
CLIENT_SECRET=
```