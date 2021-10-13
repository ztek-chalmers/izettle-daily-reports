package main

import (
	"context"
	"fmt"
	"izettle-daily-reports/google"
	"izettle-daily-reports/izettle"
	"izettle-daily-reports/storage"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func main() {
	fmt.Printf("izettle-report-generator run at %s\n\n", time.Now())

	fmt.Print("Loading .env file... ")
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("FAILED: %s\n", err)
	} else {
		fmt.Println("DONE")
	}

	fmt.Print("Reading environment variables... ")
	applicationID := MustGetEnv("GDRIVE_CLIENT_ID")
	applicationSecret := MustGetEnv("GDRIVE_CLIENT_SECRET")
	driveDirectoryID := MustGetEnv("GDRIVE_FOLDER_ID")
	izettleSession := MustGetEnv("IZETTLE_SESSION")
	fmt.Println("DONE")

	fmt.Print("Logging in to your google account... ")
	token, err := google.Login(applicationID, applicationSecret, []string{drive.DriveScope})
	handleError(err)
	driveService, err := drive.NewService(context.Background(), option.WithTokenSource(*token))
	handleError(err)
	folders, err := storage.Children(driveService, driveDirectoryID)
	handleError(err)
	fmt.Println("DONE")

	fmt.Print("Logging in to your izettle account... ")
	izettleClient, err := izettle.Login(izettleSession)
	handleError(err)
	users, err := izettleClient.ListUsers()
	handleError(err)
	fmt.Println("DONE")

	fmt.Println("Looking for missing PDFs...")
	missingPDFs := make([]storage.MissingPDF, 0)
	for _, user := range users {
		name := user.Name
		fmt.Printf("Fetching reports for %s...\n", name)
		// TODO: make this configurable outside of the source code.
		if name == "zkk" {
			name = "ztyret"
		}
		folder, ok := folders[name]
		if !ok {
			handleError(fmt.Errorf("folder not found for user %s", name))
		}
		reports, err := izettleClient.ListReports(user)
		handleError(err)
		fmt.Printf(" - Comparing %d report(s) against Google Drive...\n", len(reports))
		missing, err := storage.MissingPDFs(driveService, folder, reports)
		handleError(err)
		fmt.Printf(" - Found %d missing report(s)\n", len(missing))
		for _, r := range missing {
			fmt.Printf("    * %s\n", r.Report.Date)
			missingPDFs = append(missingPDFs, r)
		}
	}
	fmt.Println()

	if len(missingPDFs) == 0 {
		fmt.Println("No files seem to be missing so we are done!")
		os.Exit(0)
	}

	fmt.Printf("Creating %d missing PDF(s)...\n", len(missingPDFs))
	for i, r := range missingPDFs {
		filePath := fmt.Sprintf("%s/%s/%s", r.Report.User.Name, r.Report.Date, storage.ReportFileName(r.Report))
		fmt.Printf(" - Processing report %d of %d (%s)\n", i, len(missingPDFs), filePath)
		fmt.Println("    * Generating...")
		pdf, err := izettleClient.DayReportToPDF(r.Report)
		handleError(err)
		fmt.Println("    * Uploading...")
		err = storage.UploadPDF(driveService, r, pdf)
		handleError(err)
	}
	fmt.Println()

	fmt.Println("Finished generating reports!")
	fmt.Println()
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func MustGetEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("Failed to get environment variable '%s'", name)
	}
	return value
}
