package main

import (
	"context"
	"fmt"
	"io"
	"izettle-daily-reports/google"
	"izettle-daily-reports/izettle"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func getTargetFolders(ds *drive.Service, dir string) (map[string]*drive.File, error) {
	folderList, err := ds.Files.List().
		Q(fmt.Sprintf("'%s' in parents", dir)).
		IncludeItemsFromAllDrives(true).
		SupportsTeamDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	folders := make(map[string]*drive.File)
	for _, folder := range folderList.Files {
		folders[folder.Name] = folder
	}
	return folders, nil
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type ReportFile struct {
	report izettle.Report
	dir    *drive.File
}

func getToday() string {
	year, month, day := time.Now().Date()
	return fmt.Sprintf("%d-%d-%d", year, month, day)
}
func missingReports(ds *drive.Service, root *drive.File, reports []izettle.Report) ([]ReportFile, error) {
	missing := make([]ReportFile, 0)
	today := getToday()

	reportYears := make(map[string][]izettle.Report)
	for _, report := range reports {
		year := strings.Split(report.From, "-")[0]
		reportYears[year] = append(reportYears[year], report)
	}
	yearDirList, err := ds.Files.List().
		Q(fmt.Sprintf("'%s' in parents", root.Id)).
		IncludeItemsFromAllDrives(true).
		SupportsTeamDrives(true).
		Do()
	if err != nil {
		return nil, err
	}
	dirYears := make(map[string]*drive.File)
	for _, file := range yearDirList.Files {
		dirYears[file.Name] = file
	}

	for year, reports := range reportYears {
		yearDir, ok := dirYears[year]
		if !ok {
			for _, report := range reports {
				if report.From == today {
					continue
				}
				missing = append(missing, ReportFile{report: report, dir: root})
			}
			continue
		}
		files, err := ds.Files.List().
			Q(fmt.Sprintf("'%s' in parents", yearDir.Id)).
			IncludeItemsFromAllDrives(true).
			SupportsTeamDrives(true).
			Do()
		if err != nil {
			return nil, err
		}
		filesByDate := make(map[string]*drive.File)
		for _, file := range files.Files {
			date, err := fileDate(file.Name)
			if err != nil {
				continue
			}
			filesByDate[date] = file
		}
		for _, report := range reports {
			if _, ok := filesByDate[report.From]; !ok {
				if report.From == today {
					continue
				}
				missing = append(missing, ReportFile{report: report, dir: yearDir})
			}
		}
	}

	return missing, nil
}

func uploadPDF(ds *drive.Service, report ReportFile, data io.Reader) error {
	isYear := regexp.MustCompile("\\d+")
	dir := report.dir
	if !isYear.Match([]byte(dir.Name)) {
		year := strings.Split(report.report.From, "-")[0]
		dir = &drive.File{
			MimeType: "application/vnd.google-apps.folder",
			Name:     year,
			Parents:  []string{report.dir.Id},
		}
		_, err := ds.Files.
			Create(dir).
			SupportsTeamDrives(true).
			Do()
		if err != nil {
			return err
		}
	}
	f := &drive.File{
		MimeType: "application/pdf",
		Name:     reportName(report.report),
		Parents:  []string{report.dir.Id},
	}
	_, err := ds.Files.
		Create(f).
		SupportsTeamDrives(true).
		Media(data).
		Do()
	return err
}

func reportName(report izettle.Report) string {
	return fmt.Sprintf("%s_%s.pdf", report.User.Name, report.From)
}

func fileDate(name string) (string, error) {
	getDate := regexp.MustCompile(".*_(\\d+-\\d+-\\d+).pdf")
	date := getDate.FindStringSubmatch(name)
	if date == nil {
		return "", fmt.Errorf("failed to get date from file")
	}
	return date[1], nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Failed to load .env file: %s", err)
	}
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	token, err := google.Login(clientID, clientSecret, []string{drive.DriveScope})
	handleError(err)
	driveService, err := drive.NewService(context.Background(), option.WithTokenSource(*token))
	handleError(err)
	dir := os.Getenv("GDRIVE_FOLDER_ID")
	folders, err := getTargetFolders(driveService, dir)
	handleError(err)

	user := os.Getenv("IZETTLE_EMAIL")
	password := os.Getenv("IZETTLE_PASSWORD")
	izettle, err := izettle.Login(user, password)
	handleError(err)
	users, err := izettle.ListUsers()
	handleError(err)

	toGenerate := make([]ReportFile, 0)
	for komitee, user := range users {
		folder := folders[komitee]
		reports, err := izettle.ListReports(user)
		handleError(err)
		missing, err := missingReports(driveService, folder, reports)
		handleError(err)
		for _, r := range missing {
			toGenerate = append(toGenerate, r)
		}
	}
	for _, r := range toGenerate {
		pdf, err := izettle.GenerateReport(r.report)
		handleError(err)
		err = uploadPDF(driveService, r, pdf)
		handleError(err)
	}
}
