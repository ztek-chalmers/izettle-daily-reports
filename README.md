# iZettle Daily Reports

This is a small tool to generate daily iZettle reports which then are uploaded to Google Drive.

## Configuration

The application can be configured by setting the following environment variables 

```bash
IZETTLE_EMAIL=# Your iZettle email adress
IZETTLE_PASSWORD=# Your iZettle password
GDRIVE_FOLDER_ID=# The folder where the reports will be generated
CLIENT_ID=# The client id for the application which can be created at https://console.developers.google.com
CLIENT_SECRET=# The client secret which can be accessde when creating a new application
```