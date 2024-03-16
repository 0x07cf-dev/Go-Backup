package backup

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/0x07cf-dev/go-backup/internal/lang"
	"github.com/0x07cf-dev/go-backup/internal/logger"
)

type BackupErrorCode int8

type BackupError struct {
	Code    BackupErrorCode
	Source  string
	Message string
}

const (
	GenericError BackupErrorCode = iota
	CmdInvalid
	CmdFailed
	PathError
	UploadError
	DownloadError
)

var backupErrIDs = []string{
	"ErrorGeneric",
	"ErrorCmdInvalid",
	"ErrorCmdFailed",
	"ErrorPath",
	"ErrorUpload",
	"ErrorDownload",
}

func (e BackupErrorCode) ID() string {
	return backupErrIDs[e]
}

func (e BackupErrorCode) Error(source string, message string) BackupError {
	return BackupError{
		Code:    e,
		Source:  source,
		Message: message,
	}
}

func (e BackupError) Localize(langs ...string) string {
	template := map[string]string{
		"Source":  e.Source,
		"Message": e.Message,
	}
	return lang.GetManager().LocalizeTemplate(e.Code.ID(), template, langs...)
}

func getStatus(errCh chan BackupError, preErrCh chan BackupError, postErrCh chan BackupError, langs ...string) (string, string) {
	var status strings.Builder
	statusEmoji := "green_circle"

	// Check all errors
	var transferErrors []BackupError
	var preErrors []BackupError
	var postErrors []BackupError

	for err := range errCh {
		transferErrors = append(transferErrors, err)
	}
	for err := range preErrCh {
		preErrors = append(preErrors, err)
	}
	for err := range postErrCh {
		postErrors = append(postErrors, err)
	}

	failedTransfers := len(transferErrors)
	failedPre := len(preErrors)
	failedPost := len(postErrors)

	if failedTransfers == 0 && failedPre == 0 && failedPost == 0 {
		status.WriteString(lang.GetManager().Localize("Success", langs...))
	} else {
		possibleFails := cap(errCh) + cap(preErrCh) + cap(postErrCh)
		totalFails := failedTransfers + failedPre + failedPost
		totalFailRate := int(float32(totalFails) / float32(possibleFails) * 100)
		logger.Warnf("Failures: %d/%d (%d%%)", totalFails, possibleFails, totalFailRate)

		// Not a perfect run
		if totalFailRate > 0 {
			str := lang.GetManager().LocalizeTemplate("Fail", map[string]string{
				"Failed": strconv.Itoa(totalFailRate),
			}, langs...)
			status.WriteString(str + "\n\n")
		}

		if totalFailRate > 97 {
			return status.String(), "black_circle"
		}

		// Append upload errors
		if failedTransfers > 0 {
			transferFailRate := int8(float32(failedTransfers) / float32(cap(errCh)) * 100)
			logger.Warnf("Transfer failed: %d/%d (%d%%)\n", failedTransfers, cap(errCh), transferFailRate)

			if transferFailRate > 10 {
				str := lang.GetManager().LocalizeTemplate("FailedTransferNum", map[string]string{
					"Failed": fmt.Sprintf("%d%%", transferFailRate),
				}, langs...)
				status.WriteString(str + "\n")
			}
			for i, err := range transferErrors {
				status.WriteString(fmt.Sprintf("%d° | %s\n", i+1, err.Localize(langs...)))
			}
			status.WriteString("\n")
		}

		// Append pre errors
		if failedPre > 0 {
			logger.Warnf("Failed pre-transfer commands: %d/%d", failedPre, cap(preErrCh))
			str := lang.GetManager().LocalizeTemplate("FailedPreNum", map[string]string{
				"Failed": strconv.Itoa(failedPre),
			}, langs...)
			status.WriteString(str + "\n")

			for i, err := range preErrors {
				status.WriteString(fmt.Sprintf("%d° | %s\n", i+1, err.Localize(langs...)))
			}
		}

		// Append post errors
		if failedPost > 0 {
			logger.Warnf("Failed post-transfer commands: %d/%d", failedPost, cap(postErrCh))
			str := lang.GetManager().LocalizeTemplate("FailedPostNum", map[string]string{
				"Failed": strconv.Itoa(failedPost),
			}, langs...)
			status.WriteString(str + "\n")

			for i, err := range postErrors {
				status.WriteString(fmt.Sprintf("%d° | %s\n", i+1, err.Localize(langs...)))
			}
		}

		// Determine notification circle emoji color based on total fail rate
		colors := []string{"green_circle", "yellow_circle", "orange_circle", "red_circle"}
		thresh := []int{10, 50, 80, 100}
		lev := 0

		for i, th := range thresh {
			if totalFailRate <= th {
				lev = i
				break
			}
		}
		statusEmoji = colors[lev%len(colors)]
	}
	return status.String(), statusEmoji
}
